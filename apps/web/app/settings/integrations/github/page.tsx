import { GitHubConnectCard } from "@/components/integrations/github-connect-card";

export default function GitHubIntegrationSettingsPage() {
  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h2 className="text-3xl font-semibold tracking-tight">Integrations</h2>
        <p className="text-slate-600">Manage external providers connected to your RepoMemory account.</p>
      </div>
      <GitHubConnectCard />
    </section>
  );
}