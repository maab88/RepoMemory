import React from "react";
import { render, screen } from "@testing-library/react";

import { StatusBadge } from "@/components/status-badge";

describe("StatusBadge", () => {
  it("renders the label", () => {
    render(<StatusBadge label="healthy" />);

    expect(screen.getByText("healthy")).toBeInTheDocument();
  });
});