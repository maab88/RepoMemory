import React from "react";
import { render, screen } from "@testing-library/react";
import HealthPage from "../app/health/page";

describe("HealthPage", () => {
  it("renders health status", () => {
    render(<HealthPage />);

    expect(screen.getByRole("heading", { name: "Web Health" })).toBeInTheDocument();
    expect(screen.getByText("status: ok")).toBeInTheDocument();
  });
});