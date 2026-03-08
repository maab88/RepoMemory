import React from "react";
import { fireEvent, render, screen } from "@testing-library/react";

import { RepositorySelector, type SelectableGitHubRepository } from "@/components/repositories/repository-selector";

const repositories: SelectableGitHubRepository[] = [
  {
    githubRepoId: "1",
    ownerLogin: "octocat",
    name: "repo-memory",
    fullName: "octocat/repo-memory",
    private: true,
    defaultBranch: "main",
    htmlUrl: "https://github.com/octocat/repo-memory",
    description: "Internal tools",
  },
  {
    githubRepoId: "2",
    ownerLogin: "acme",
    name: "infra",
    fullName: "acme/infra",
    private: false,
    defaultBranch: "trunk",
    htmlUrl: "https://github.com/acme/infra",
  },
];

describe("RepositorySelector", () => {
  it("search filters list", () => {
    const onSelectionChange = vi.fn();

    render(<RepositorySelector repositories={repositories} selectedIds={[]} onSelectionChange={onSelectionChange} />);

    fireEvent.change(screen.getByLabelText("Search repositories"), { target: { value: "infra" } });

    expect(screen.getByText("acme/infra")).toBeInTheDocument();
    expect(screen.queryByText("octocat/repo-memory")).not.toBeInTheDocument();
  });

  it("multi-select updates selected count", () => {
    function Wrapper() {
      const [selected, setSelected] = React.useState<string[]>([]);
      return <RepositorySelector repositories={repositories} selectedIds={selected} onSelectionChange={setSelected} />;
    }

    render(<Wrapper />);

    fireEvent.click(screen.getByLabelText("Select octocat/repo-memory"));
    fireEvent.click(screen.getByLabelText("Select acme/infra"));

    expect(screen.getByText("Selected: 2")).toBeInTheDocument();
  });
});