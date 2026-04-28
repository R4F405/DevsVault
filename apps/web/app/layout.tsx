import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "DevsVault",
  description: "Secrets manager for local development and production"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}