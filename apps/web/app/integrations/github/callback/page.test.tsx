import React from "react";
import { render, screen } from "@testing-library/react";

import { GitHubCallbackStatus } from "@/components/integrations/github-callback-status";

const useGitHubCallbackMock = vi.fn();

vi.mock("@/lib/hooks/use-github-callback", () => ({
  useGitHubCallback: (code?: string, state?: string) => useGitHubCallbackMock(code, state),
}));

describe("GitHub callback states", () => {
  it("renders loading state", () => {
    useGitHubCallbackMock.mockReturnValue({ isPending: true, error: null, data: null });

    render(<GitHubCallbackStatus code="code123" state="state123" />);

    expect(screen.getByRole("heading", { name: "Finishing GitHub connection..." })).toBeInTheDocument();
  });

  it("renders success state", () => {
    useGitHubCallbackMock.mockReturnValue({
      isPending: false,
      error: null,
      data: {
        connected: true,
        account: { id: "1", githubLogin: "octocat", githubUserId: "1", connectedAt: "2026-03-07T12:00:00Z" },
      },
    });

    render(<GitHubCallbackStatus code="code123" state="state123" />);

    expect(screen.getByRole("heading", { name: "GitHub connected" })).toBeInTheDocument();
    expect(screen.getByText(/octocat/)).toBeInTheDocument();
  });

  it("renders failure and retry action", () => {
    useGitHubCallbackMock.mockReturnValue({ isPending: false, error: new Error("failed"), data: null });

    render(<GitHubCallbackStatus code="code123" state="state123" />);

    expect(screen.getByRole("heading", { name: "GitHub connection failed" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Retry GitHub connect" })).toBeInTheDocument();
  });
});