import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Alert Console",
  description: "Phase 2 compliance alert feed",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="border-b border-amber-500/40 bg-amber-950/40 px-4 py-2 text-sm text-amber-200">
          Dev mode — no authentication (Phase 2)
        </div>
        {children}
      </body>
    </html>
  );
}
