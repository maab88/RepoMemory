import type { ImportedRepository } from "@repomemory/contracts";

type RepositoryImportTableProps = {
  repositories: ImportedRepository[];
};

export function RepositoryImportTable({ repositories }: RepositoryImportTableProps) {
  if (repositories.length === 0) {
    return null;
  }

  return (
    <div className="overflow-hidden rounded-xl border border-emerald-200 bg-emerald-50">
      <div className="border-b border-emerald-200 px-4 py-3">
        <h3 className="text-sm font-semibold text-emerald-900">Imported repositories</h3>
      </div>
      <table className="min-w-full divide-y divide-emerald-200 text-sm">
        <thead className="bg-emerald-100/80 text-left text-xs uppercase tracking-wide text-emerald-800">
          <tr>
            <th className="px-4 py-2">Repository</th>
            <th className="px-4 py-2">Branch</th>
            <th className="px-4 py-2">Visibility</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-emerald-200 text-emerald-900">
          {repositories.map((repo) => (
            <tr key={repo.id}>
              <td className="px-4 py-2 font-medium">{repo.fullName}</td>
              <td className="px-4 py-2">{repo.defaultBranch}</td>
              <td className="px-4 py-2">{repo.private ? "Private" : "Public"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}