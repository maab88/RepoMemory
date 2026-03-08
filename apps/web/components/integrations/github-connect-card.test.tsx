import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { GitHubConnectCard } from "@/components/integrations/github-connect-card";

const startGitHubConnectMock = vi.fn();
const assignMock = vi.fn();

vi.mock("@/lib/github-api", () => ({
  startGitHubConnect: (input?: { organizationId?: string }) => startGitHubConnectMock(input),
}));

function renderCard() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={client}>
      <GitHubConnectCard />
    </QueryClientProvider>
  );
}

describe("GitHubConnectCard", () => {
  beforeEach(() => {
    startGitHubConnectMock.mockReset();
    assignMock.mockReset();
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { assign: assignMock },
    });
  });

  it("renders CTA and description", () => {
    renderCard();

    expect(screen.getByRole("heading", { name: "Connect your GitHub account" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Connect GitHub" })).toBeInTheDocument();
  });

  it("starts oauth and redirects", async () => {
    startGitHubConnectMock.mockResolvedValue({ redirectUrl: "https://github.com/login/oauth/authorize?state=abc" });

    renderCard();
    fireEvent.click(screen.getByRole("button", { name: "Connect GitHub" }));

    await waitFor(() => {
      expect(startGitHubConnectMock).toHaveBeenCalledTimes(1);
      expect(assignMock).toHaveBeenCalledWith("https://github.com/login/oauth/authorize?state=abc");
    });
  });
});