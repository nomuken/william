import { useMemo } from "react";
import useSWR from "swr";

import { createAdminClient } from "@/lib/adminClient";

export function useFirewallRules() {
  const client = useMemo(() => createAdminClient(), []);

  const { data, error, isLoading, mutate } = useSWR(
    "firewall-rules",
    async () => {
      const response = await client.getFirewallRules({});
      return response.rules;
    },
    {
      refreshInterval: 60_000,
    },
  );

  return {
    rules: data ?? "",
    error,
    isLoading,
    mutate,
  } as const;
}
