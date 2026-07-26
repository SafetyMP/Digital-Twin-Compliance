import { AppShell, consoleUrlsFromEnv } from "@digital-twin/console-shell";
import type { Metadata } from "next";
import { IBM_Plex_Sans } from "next/font/google";
import "./globals.css";

const plex = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Alert Console · Digital Twin",
  description: "Monitoring & alerts · Phase 2 compliance alert feed",
  icons: { icon: "/favicon.svg" },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const urls = consoleUrlsFromEnv();
  return (
    <html lang="en" className={plex.className}>
      <body>
        <AppShell activeApp="alert" urls={urls}>
          {children}
        </AppShell>
      </body>
    </html>
  );
}
