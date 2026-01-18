import { useMemo } from "react";
import useSWR from "swr";

import type { PeerStatus } from "@/gen/proto/server/v1/server_pb";
import { createWilliamClient } from "@/lib/williamClient";

type RequestOptions = { headers?: Headers } | undefined;

export function usePeerStatuses(enabled: boolean, requestOptions: RequestOptions) {
  const client = useMemo(() => createWilliamClient(), []);

  const { data, error, isLoading, mutate } = useSWR<PeerStatus[]>(
    enabled ? "peer-statuses" : null,
    async () => {
      const response = await client.listPeerStatuses({}, requestOptions);
      return response.statuses;
    },
    {
      refreshInterval: 10_000,
    },
  );

  return {
    statuses: data ?? [],
    error,
    isLoading,
    mutate,
  } as const;
}
