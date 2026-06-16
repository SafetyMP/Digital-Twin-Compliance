import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Audit Explorer",
  description: "Phase 3 tamper-evident audit ledger search",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="border-b border-emerald-500/40 bg-emerald-950/40 px-4 py-2 text-sm text-emerald-200">
          Audit Explorer — hash-chain integrity badges (Phase 3)
        </div>
        {children}
      </body>
    </html>
  );
}
