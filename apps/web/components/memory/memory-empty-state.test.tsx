import React from "react";
import { fireEvent, render, screen } from "@testing-library/react";

import { MemoryEmptyState } from "@/components/memory/memory-empty-state";

describe("MemoryEmptyState", () => {
  it("renders CTA and handles click", () => {
    const onGenerate = vi.fn();
    render(<MemoryEmptyState onGenerateMemory={onGenerate} isGenerating={false} generationStatus={null} generationError={null} />);

    expect(screen.getByText("No memory entries yet")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate memory" }));
    expect(onGenerate).toHaveBeenCalled();
  });
});
