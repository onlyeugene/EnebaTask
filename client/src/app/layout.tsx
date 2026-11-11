import type { Metadata } from "next";
import "@fontsource/metropolis/400.css";
import "@fontsource/metropolis/600.css";
import "@fontsource/metropolis/700.css";
import "@fontsource/metropolis/800.css";
import "./globals.css";
import { Providers } from "@/lib/providers";

export const metadata: Metadata = {
  title: "Eugene | Game keys marketplace",
  description: "Demo marketplace for digital game keys. Not affiliated with Eneba.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
