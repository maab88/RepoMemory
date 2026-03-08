import type { Digest } from "@repomemory/contracts";

import { DigestCard } from "@/components/digests/digest-card";

type DigestListProps = {
  digests: Digest[];
  selectedDigestId: string | null;
  onSelectDigest: (digestId: string) => void;
};

export function DigestList({ digests, selectedDigestId, onSelectDigest }: DigestListProps) {
  return (
    <div className="space-y-3">
      {digests.map((digest) => (
        <DigestCard key={digest.id} digest={digest} selected={digest.id === selectedDigestId} onSelect={onSelectDigest} />
      ))}
    </div>
  );
}
