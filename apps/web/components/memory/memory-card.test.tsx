import React from "react";
import { fireEvent, render, screen } from "@testing-library/react";

import { MemoryCard } from "@/components/memory/memory-card";

describe("MemoryCard", () => {
  it("renders key fields and opens detail", () => {
    const onOpen = vi.fn();
    render(
      <MemoryCard
        entry={{
          id: "mem-1",
          repositoryId: "repo-1",
          organizationId: "org-1",
          type: "pr_summary",
          title: "Refactored retry flow",
          summary: "Moved retry scheduling into worker.",
          whyItMatters: "Reduces duplicate retries.",
          impactedAreas: ["billing", "sync"],
          risks: ["retry timing drift"],
          followUps: ["monitor failure rate"],
          generatedBy: "deterministic",
          sourceUrl: "https://github.com/org/repo/pull/123",
          createdAt: "2026-03-07T12:00:00Z",
        }}
        onOpen={onOpen}
      />
    );

    expect(screen.getByText("Refactored retry flow")).toBeInTheDocument();
    expect(screen.getByText(/Why it matters:/)).toBeInTheDocument();
    expect(screen.getByText("billing")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open source link" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Open detail" }));
    expect(onOpen).toHaveBeenCalledWith("mem-1");
  });

  it("renders hotspot type label", () => {
    render(
      <MemoryCard
        entry={{
          id: "mem-2",
          repositoryId: "repo-1",
          organizationId: "org-1",
          type: "hotspot",
          title: "Recurring sync instability",
          summary: "Multiple recent sync issues and PRs indicate churn.",
          whyItMatters: "",
          impactedAreas: ["sync"],
          risks: [],
          followUps: [],
          generatedBy: "deterministic",
          sourceUrl: "",
          createdAt: "2026-03-08T10:00:00Z",
        }}
        onOpen={() => {}}
      />
    );

    expect(screen.getByText("Hotspot")).toBeInTheDocument();
  });
});
