import React from "react";
import { render, screen } from "@testing-library/react";

import OnboardingRepositoriesPage from "@/app/onboarding/repositories/page";

vi.mock("@/components/repositories/repository-import-card", () => ({
  RepositoryImportCard: () => <div data-testid="repository-import-card">Repository Import Card</div>,
}));

describe("OnboardingRepositoriesPage", () => {
  it("renders repository import onboarding page", () => {
    render(<OnboardingRepositoriesPage />);

    expect(screen.getByRole("heading", { name: "Import repositories" })).toBeInTheDocument();
    expect(screen.getByTestId("repository-import-card")).toBeInTheDocument();
  });
});