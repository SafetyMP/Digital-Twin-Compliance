import { AppShell, consoleUrlsFromEnv } from "@digital-twin/console-shell";
import type { Metadata } from "next";
import { IBM_Plex_Sans } from "next/font/google";
import "./globals.css";

const plex = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Report Console · Digital Twin",
  description: "Regulatory reporting · Phase 5 report lifecycle",
  icons: { icon: "/favicon.svg" },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const urls = consoleUrlsFromEnv();
  return (
    <html lang="en" className={plex.className}>
      <body>
        <AppShell activeApp="report" urls={urls}>
          {children}
        </AppShell>
      </body>
    </html>
  );
}
