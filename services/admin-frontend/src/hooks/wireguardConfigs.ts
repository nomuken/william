import { useMemo } from "react";
import useSWR from "swr";

import type { WireguardConfig } from "@/gen/proto/admin/v1/admin_pb";
import { createAdminClient } from "@/lib/adminClient";

export function useWireguardConfigs(interfaceId: string) {
  const client = useMemo(() => createAdminClient(), []);

  const { data, error, isLoading, mutate } = useSWR<WireguardConfig[]>(
    interfaceId ? ["wireguard-configs", interfaceId] : null,
    async () => {
      const response = await client.listWireguardConfigs({ interfaceId });
      return response.configs;
    },
  );

  return {
    configs: data ?? [],
    error,
    isLoading,
    mutate,
  } as const;
}
