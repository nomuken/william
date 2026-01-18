import { useMemo } from "react";
import useSWR from "swr";

import type { PeerRoute } from "@/gen/proto/admin/v1/admin_pb";
import { createAdminClient } from "@/lib/adminClient";

export function usePeerRoutes(peerId: string) {
  const client = useMemo(() => createAdminClient(), []);

  const { data, error, isLoading, mutate } = useSWR<PeerRoute[]>(
    peerId ? ["peer-routes", peerId] : null,
    async () => {
      const response = await client.listPeerRoutes({ peerId });
      return response.routes;
    },
  );

  return {
    routes: data ?? [],
    error,
    isLoading,
    mutate,
  } as const;
}
