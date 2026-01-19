import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { WilliamAdminService } from "@/gen/proto/admin/v1/admin_pb";

const defaultBaseUrl = "http://localhost:8081/api";

// createAdminClient wires the Connect RPC client for the admin API.
export function createAdminClient(baseUrl?: string) {
  const transport = createConnectTransport({
    baseUrl: baseUrl ?? process.env.NEXT_PUBLIC_WILLIAM_ADMIN_API_BASE_URL ?? defaultBaseUrl,
  });

  return createClient(WilliamAdminService, transport);
}
