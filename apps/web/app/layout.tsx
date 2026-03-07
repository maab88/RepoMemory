import type { Metadata } from "next";
import "./globals.css";
import Providers from "./providers";

export const metadata: Metadata = {
  title: "RepoMemory",
  description: "Engineering memory for teams",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>
          <div className="min-h-screen bg-slate-50">
            <header className="border-b border-slate-200 bg-white">
              <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-4">
                <h1 className="text-xl font-semibold tracking-tight">RepoMemory</h1>
              </div>
            </header>
            <main className="mx-auto max-w-5xl px-4 py-8">{children}</main>
          </div>
        </Providers>
      </body>
    </html>
  );
}