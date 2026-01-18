import { useMemo } from "react";
import useSWR from "swr";

import type { AdminWireguardInterface } from "@/gen/proto/admin/v1/admin_pb";
import { createAdminClient } from "@/lib/adminClient";

const emptyRequest = {};

export function useAdminInterfaces() {
  const client = useMemo(() => createAdminClient(), []);

  const { data, error, isLoading, mutate } = useSWR<AdminWireguardInterface[]>(
    "admin-interfaces",
    async () => {
      const response = await client.listInterfaces(emptyRequest);
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
