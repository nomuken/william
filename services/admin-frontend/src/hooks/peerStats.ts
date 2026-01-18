import { useMemo } from "react";
import useSWR from "swr";

import type { PeerStat } from "@/gen/proto/admin/v1/admin_pb";
import { createAdminClient } from "@/lib/adminClient";

export function usePeerStats() {
  const client = useMemo(() => createAdminClient(), []);

  const { data, error, isLoading, mutate } = useSWR<PeerStat[]>(
    "peer-stats",
    async () => {
      const response = await client.listPeerStats({});
      return response.stats;
    },
    { refreshInterval: 5000 },
  );

  return {
    stats: data ?? [],
    error,
    isLoading,
    mutate,
  } as const;
}
