import { useMemo } from "react";
import useSWR from "swr";

import type { AdminPeer } from "@/gen/proto/admin/v1/admin_pb";
import { createAdminClient } from "@/lib/adminClient";

export function useAdminPeers(interfaceId: string | null) {
  const client = useMemo(() => createAdminClient(), []);
  const filter = interfaceId ?? "";

  const { data, error, isLoading, mutate } = useSWR<AdminPeer[]>(
    interfaceId === null ? null : ["admin-peers", filter],
    async () => {
      const response = await client.listPeers({ interfaceId: filter });
      return response.peers;
    },
  );

  return {
    peers: data ?? [],
    error,
    isLoading,
    mutate,
  } as const;
}
