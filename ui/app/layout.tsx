import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Eko — Visual Memory System",
  description: "Time-travel through your AI coding sessions. Inspect, diff, and restore every snapshot.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
