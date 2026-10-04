"""AIMeter Asynchronous Non-blocking Background Reporter."""

import atexit
import logging
import queue
import threading
import time
from typing import Any, Dict, List, Optional
import httpx

logger = logging.getLogger("aimeter")


class BackgroundReporter:
    """Non-blocking background worker that aggregates and batches usage events."""

    def __init__(
        self,
        endpoint_url: str,
        batch_size: int = 50,
        flush_interval_seconds: float = 0.5,
        max_queue_size: int = 10000,
        headers: Optional[Dict[str, str]] = None,
    ):
        self.endpoint_url = endpoint_url
        self.batch_size = batch_size
        self.flush_interval = flush_interval_seconds
        self.headers = headers or {"Content-Type": "application/json"}

        self._queue: queue.Queue = queue.Queue(maxsize=max_queue_size)
        self._lock = threading.Lock()
        self._stop_event = threading.Event()
        self._flush_event = threading.Event()

        self._client = httpx.Client(
            timeout=5.0,
            limits=httpx.Limits(max_keepalive_connections=5, max_connections=10),
        )

        self._thread = threading.Thread(
            target=self._worker_loop,
            name="aimeter-reporter-thread",
            daemon=True,
        )
        self._thread.start()

        # Register atexit handler for clean flush
        atexit.register(self.shutdown)

    def enqueue(self, event: Dict[str, Any]) -> bool:
        """Enqueue an event without blocking the caller."""
        try:
            self._queue.put_nowait(event)
            if self._queue.qsize() >= self.batch_size:
                self._flush_event.set()
            return True
        except queue.Full:
            logger.warning("[AIMeter] Ingestion queue is full. Dropping usage event to protect memory.")
            return False

    def flush(self, timeout: float = 3.0) -> None:
        """Trigger an immediate batch send and block until queue is drained."""
        self._flush_event.set()
        start = time.time()
        while not self._queue.empty() and (time.time() - start < timeout):
            time.sleep(0.05)

    def shutdown(self) -> None:
        """Flush remaining events and terminate worker."""
        if self._stop_event.is_set():
            return
        self._stop_event.set()
        self._flush_event.set()
        self.flush(timeout=2.0)
        try:
            self._client.close()
        except Exception:
            pass

    def _worker_loop(self) -> None:
        last_flush = time.time()

        while not self._stop_event.is_set():
            # Wait for either flush trigger, timeout, or queue build-up
            triggered = self._flush_event.wait(timeout=self.flush_interval)
            if triggered:
                self._flush_event.clear()

            now = time.time()
            if self._queue.qsize() >= self.batch_size or (now - last_flush) >= self.flush_interval or self._stop_event.is_set():
                self._drain_and_send()
                last_flush = now

        # Final drain after stop signal
        self._drain_and_send()

    def _drain_and_send(self) -> None:
        batch: List[Dict[str, Any]] = []

        while len(batch) < self.batch_size * 2:
            try:
                item = self._queue.get_nowait()
                batch.append(item)
            except queue.Empty:
                break

        if not batch:
            return

        try:
            resp = self._client.post(self.endpoint_url, json=batch, headers=self.headers)
            if resp.status_code >= 400:
                logger.warning(
                    f"[AIMeter] Ingestion endpoint returned {resp.status_code}: {resp.text[:200]}"
                )
        except Exception as e:
            logger.warning(f"[AIMeter] Failed to send usage batch to {self.endpoint_url}: {e}")
        finally:
            for _ in batch:
                self._queue.task_done()
