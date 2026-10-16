import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Navbar } from "@/components/Navbar";
import { Suspense } from "react";
import { AuthGate } from "@/components/AuthGate";
import { FeedbackHost } from "@/components/FeedbackHost";
import { CommandPalette } from "@/components/CommandPalette";

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" });
const mono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono" });

export const metadata: Metadata = {
  title: "AI Meter | AI Usage & Cost Control Plane",
  description: "Independent AI Usage & Cost Control Plane for Enterprises and Multi-Agent Workflows",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} ${mono.variable} font-sans bg-zinc-950 text-zinc-100 min-h-screen antialiased flex flex-col`}>
        <Navbar />
        <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Pages read shared filters from the URL (useSearchParams), which needs a Suspense boundary */}
          <Suspense fallback={null}>
            <AuthGate>{children}</AuthGate>
            <CommandPalette />
          </Suspense>
          <FeedbackHost />
        </main>
      </body>
    </html>
  );
}
