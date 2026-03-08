import React from "react";
import { render, screen } from "@testing-library/react";

import { OrganizationsView } from "@/components/organizations/organizations-view";

describe("OrganizationsView", () => {
  it("renders empty organizations state", () => {
    render(<OrganizationsView organizations={[]} />);

    expect(screen.getByTestId("org-empty-state")).toBeInTheDocument();
    expect(screen.getByText("No organizations yet")).toBeInTheDocument();
  });

  it("renders organizations list", () => {
    render(
      <OrganizationsView
        organizations={[
          { id: "1", name: "Acme", slug: "acme", role: "owner" },
          { id: "2", name: "Beta", slug: "beta", role: "member" },
        ]}
      />
    );

    expect(screen.getByText("Acme")).toBeInTheDocument();
    expect(screen.getByText("Beta")).toBeInTheDocument();
    expect(screen.getByText("Role: owner")).toBeInTheDocument();
  });
});