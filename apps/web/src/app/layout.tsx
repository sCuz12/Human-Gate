import type { Metadata } from "next";
import Script from "next/script";
import "./globals.css";

export const metadata: Metadata = {
  title: "Decree",
  description: "Centralized approval workflows for AI agents and automation.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const enableAnalytics = process.env.NODE_ENV === "production";

  return (
    <html lang="en">
      <head>
        {enableAnalytics ? (
          <Script
            defer
            src="https://cloud.umami.is/script.js"
            data-website-id="1195b258-0f5f-43e1-a6b8-83bfe360c631"
            strategy="afterInteractive"
          />
        ) : null}
      </head>
      <body>
        {children}
      </body>
    </html>
  );
}
