import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { getJob } from "@/lib/repositories-api";

type UseJobStatusOptions = {
  pollIntervalMs?: number;
  timeoutMs?: number;
};

export function useJobStatus(jobId: string | null, options?: UseJobStatusOptions) {
  const pollIntervalMs = options?.pollIntervalMs ?? 2000;
  const timeoutMs = options?.timeoutMs ?? 60000;
  const [timedOut, setTimedOut] = useState(false);
  const startedAtRef = useRef<number | null>(null);

  useEffect(() => {
    if (!jobId) {
      startedAtRef.current = null;
      setTimedOut(false);
      return;
    }
    startedAtRef.current = Date.now();
    setTimedOut(false);
  }, [jobId]);

  const query = useQuery({
    queryKey: ["job-status", jobId],
    queryFn: () => getJob(jobId ?? ""),
    enabled: Boolean(jobId),
    refetchInterval: (query) => {
      const status = query.state.data?.job.status;
      if ((status === "queued" || status === "running") && !timedOut) {
        return pollIntervalMs;
      }
      return false;
    },
  });

  useEffect(() => {
    if (!jobId || timedOut) {
      return;
    }
    const status = query.data?.job.status;
    if (status === "succeeded" || status === "failed") {
      return;
    }

    const id = window.setInterval(() => {
      if (!startedAtRef.current) {
        return;
      }
      if (Date.now()-startedAtRef.current >= timeoutMs) {
        setTimedOut(true);
      }
    }, 1000);
    return () => {
      window.clearInterval(id);
    };
  }, [jobId, query.data?.job.status, timedOut, timeoutMs]);

  return {
    ...query,
    timedOut,
  };
}
