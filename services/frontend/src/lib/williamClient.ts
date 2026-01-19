import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { WilliamService } from "@/gen/proto/server/v1/server_pb";

const defaultBaseUrl = "http://localhost:8080";

// createWilliamClient wires the Connect RPC client for the public API.
export function createWilliamClient(baseUrl?: string) {
  const transport = createConnectTransport({
    baseUrl: baseUrl ?? process.env.NEXT_PUBLIC_WILLIAM_API_BASE_URL ?? defaultBaseUrl,
  });

  return createClient(WilliamService, transport);
}
