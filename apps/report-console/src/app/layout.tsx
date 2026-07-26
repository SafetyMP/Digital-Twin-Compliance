import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Report Console",
  description: "Phase 5 regulatory report lifecycle",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="border-b border-teal-500/40 bg-teal-950/40 px-4 py-2 text-sm text-teal-200">
          Report Console — draft → validated → submitted (Phase 5)
        </div>
        {children}
      </body>
    </html>
  );
}
