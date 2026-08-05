import type { Metadata } from "next";
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
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
