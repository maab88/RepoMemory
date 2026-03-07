import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { CreateOrganizationForm } from "@/components/organizations/create-organization-form";

const createOrganizationMock = vi.fn();
const assignMock = vi.fn();

vi.mock("@/lib/organizations-api", () => ({
  createOrganization: (input: { name: string }) => createOrganizationMock(input),
}));

function renderForm() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={client}>
      <CreateOrganizationForm />
    </QueryClientProvider>
  );
}

describe("CreateOrganizationForm", () => {
  beforeEach(() => {
    createOrganizationMock.mockReset();
    assignMock.mockReset();
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { assign: assignMock },
    });
  });

  it("shows validation error for short names", async () => {
    renderForm();

    fireEvent.change(screen.getByLabelText("Organization name"), { target: { value: "A" } });
    fireEvent.click(screen.getByRole("button", { name: "Create organization" }));

    expect(await screen.findByText("Organization name must be between 2 and 80 characters.")).toBeInTheDocument();
    expect(createOrganizationMock).not.toHaveBeenCalled();
  });

  it("creates organization and navigates on success", async () => {
    createOrganizationMock.mockResolvedValue({ id: "org-123", name: "Acme", slug: "acme", role: "owner" });
    renderForm();

    fireEvent.change(screen.getByLabelText("Organization name"), { target: { value: "Acme" } });
    fireEvent.click(screen.getByRole("button", { name: "Create organization" }));

    await waitFor(() => {
      expect(createOrganizationMock).toHaveBeenCalledWith({ name: "Acme" });
      expect(assignMock).toHaveBeenCalledWith("/organizations/org-123");
    });
  });
});