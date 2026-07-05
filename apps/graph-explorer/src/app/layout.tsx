import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Graph Explorer",
  description: "Phase 4 exposure network visualization",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="border-b border-sky-500/40 bg-sky-950/40 px-4 py-2 text-sm text-sky-200">
          Graph Explorer — institution exposure network (Phase 4)
        </div>
        {children}
      </body>
    </html>
  );
}
