import { useMemo } from "react";
import useSWR from "swr";

import type { InterfaceRoute } from "@/gen/proto/admin/v1/admin_pb";
import { createAdminClient } from "@/lib/adminClient";

export function useInterfaceRoutes(interfaceId: string) {
  const client = useMemo(() => createAdminClient(), []);

  const { data, error, isLoading, mutate } = useSWR<InterfaceRoute[]>(
    interfaceId ? ["interface-routes", interfaceId] : null,
    async () => {
      const response = await client.listInterfaceRoutes({ interfaceId });
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
