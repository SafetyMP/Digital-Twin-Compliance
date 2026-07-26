import { AppShell, consoleUrlsFromEnv } from "@digital-twin/console-shell";
import type { Metadata } from "next";
import { IBM_Plex_Sans } from "next/font/google";
import "./globals.css";

const plex = IBM_Plex_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Simulation Console · Digital Twin",
  description: "Graph & simulation · Phase 4 deterministic stress simulation",
  icons: { icon: "/favicon.svg" },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const urls = consoleUrlsFromEnv();
  return (
    <html lang="en" className={plex.className}>
      <body>
        <AppShell activeApp="simulation" urls={urls}>
          {children}
        </AppShell>
      </body>
    </html>
  );
}
