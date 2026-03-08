import type { MemoryEntry } from "@repomemory/contracts";

import { MemoryCard } from "@/components/memory/memory-card";

type MemoryTimelineProps = {
  entries: MemoryEntry[];
  onOpenDetail: (memoryId: string) => void;
};

export function MemoryTimeline({ entries, onOpenDetail }: MemoryTimelineProps) {
  return (
    <div className="space-y-4">
      {entries.map((entry) => (
        <MemoryCard key={entry.id} entry={entry} onOpen={onOpenDetail} />
      ))}
    </div>
  );
}
