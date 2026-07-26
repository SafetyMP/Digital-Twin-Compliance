import { AppShell, consoleUrlsFromEnv } from "@digital-twin/console-shell";
import type { Metadata } from "next";
import { IBM_Plex_Sans } from "next/font/google";
import "./globals.css";

const plex = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Audit Explorer · Digital Twin",
  description: "Policy & audit · Phase 3 tamper-evident ledger search",
  icons: { icon: "/favicon.svg" },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const urls = consoleUrlsFromEnv();
  return (
    <html lang="en" className={plex.className}>
      <body>
        <AppShell activeApp="audit" urls={urls}>
          {children}
        </AppShell>
      </body>
    </html>
  );
}
