import { useMemo } from "react";
import useSWR from "swr";

import type { WireguardInterface } from "@/gen/proto/server/v1/server_pb";
import { createWilliamClient } from "@/lib/williamClient";

const emptyRequest = {};

export function useWireguardInterfaces(email: string) {
  const client = useMemo(() => createWilliamClient(), []);

  const { data, error, isLoading, mutate } = useSWR<WireguardInterface[]>(
    email ? ["wireguard-interfaces", email] : null,
    async () => {
      const headers = new Headers({ "X-Email": email });
      const response = await client.listWireguardInterfaces(emptyRequest, { headers });
      return response.interfaces;
    },
  );

  return {
    interfaces: data ?? [],
    error,
    isLoading,
    mutate,
  } as const;
}
