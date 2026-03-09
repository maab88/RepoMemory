import React from "react";
import { render, screen } from "@testing-library/react";

const getServerSessionMock = vi.fn();
const redirectMock = vi.fn();

vi.mock("next-auth/next", () => ({
  getServerSession: (...args: unknown[]) => getServerSessionMock(...args),
}));

vi.mock("next/navigation", () => ({
  redirect: (...args: unknown[]) => redirectMock(...args),
}));

vi.mock("@/components/auth/sign-in-form", () => ({
  SignInForm: ({ enableDevCredentials }: { enableDevCredentials: boolean }) => (
    <div>
      <h1>Sign in to RepoMemory</h1>
      <p>dev-enabled:{String(enableDevCredentials)}</p>
    </div>
  ),
}));

import SignInPage from "./page";

describe("SignInPage", () => {
  it("renders sign-in form for unauthenticated users", async () => {
    getServerSessionMock.mockResolvedValue(null);

    render(await SignInPage());

    expect(screen.getByRole("heading", { name: "Sign in to RepoMemory" })).toBeInTheDocument();
  });

  it("redirects authenticated users", async () => {
    getServerSessionMock.mockResolvedValue({ user: { id: "user-1" } });

    await SignInPage();

    expect(redirectMock).toHaveBeenCalledWith("/organizations");
  });
});
