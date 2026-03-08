import { RepositoryImportCard } from "@/components/repositories/repository-import-card";

export default function OnboardingRepositoriesPage() {
  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h2 className="text-3xl font-semibold tracking-tight">Import repositories</h2>
        <p className="text-slate-600">Step 2: choose GitHub repositories to bring into RepoMemory.</p>
      </div>
      <RepositoryImportCard />
    </section>
  );
}