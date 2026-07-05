import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Simulation Console",
  description: "Phase 4 deterministic stress simulation",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="border-b border-violet-500/40 bg-violet-950/40 px-4 py-2 text-sm text-violet-200">
          Simulation Console — ECB Adverse scenario (Phase 4)
        </div>
        {children}
      </body>
    </html>
  );
}
