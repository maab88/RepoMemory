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
                <nav className="flex items-center gap-4 text-sm">
                  <a href="/organizations" className="font-medium text-slate-600 hover:text-slate-900">
                    Organizations
                  </a>
                  <a href="/settings/integrations/github" className="font-medium text-slate-600 hover:text-slate-900">
                    Integrations
                  </a>
                </nav>
              </div>
            </header>
            <main className="mx-auto max-w-5xl px-4 py-8">{children}</main>
          </div>
        </Providers>
      </body>
    </html>
  );
}
