import type { NextConfig } from "next";

const backendUrl = process.env.AIMETER_BACKEND_URL || "http://localhost:8080";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/v1/:path*",
        destination: `${backendUrl}/api/v1/:path*`,
      },
      {
        // Backend readiness probe, used by the console's status indicator.
        source: "/api/status",
        destination: `${backendUrl}/readyz`,
      },
    ];
  },
};

export default nextConfig;
