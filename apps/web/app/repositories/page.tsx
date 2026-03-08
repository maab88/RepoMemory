import { RepositoryImportCard } from "@/components/repositories/repository-import-card";

export default function RepositoriesPage() {
  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h2 className="text-3xl font-semibold tracking-tight">Repositories</h2>
        <p className="text-slate-600">Browse GitHub repositories and import them into your organizations.</p>
      </div>
      <RepositoryImportCard />
    </section>
  );
}