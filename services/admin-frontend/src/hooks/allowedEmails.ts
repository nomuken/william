import { useMemo } from "react";
import useSWR from "swr";

import type { AllowedEmail } from "@/gen/proto/admin/v1/admin_pb";
import { createAdminClient } from "@/lib/adminClient";

export function useAllowedEmails(interfaceId: string) {
  const client = useMemo(() => createAdminClient(), []);

  const { data, error, isLoading, mutate } = useSWR<AllowedEmail[]>(
    interfaceId ? ["allowed-emails", interfaceId] : null,
    async () => {
      const response = await client.listAllowedEmails({ interfaceId });
      return response.emails;
    },
  );

  return {
    emails: data ?? [],
    error,
    isLoading,
    mutate,
  } as const;
}
