"use client";

import { useFirewallRules } from "@/hooks/firewallRules";
import { normalizeError } from "@/lib/normalizeError";

export default function FirewallPage() {
  const { rules, error, isLoading } = useFirewallRules();
  const message = error ? normalizeError(error) : "";

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">Firewall</h1>
        <p className="text-sm text-neutral-400">iptablesのWILLIAM_FWDルールを表示します。</p>
      </header>

      {message && <div className="rounded-md border border-red-400/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">{message}</div>}

      <section className="space-y-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">WILLIAM_FWD</h2>
          {isLoading && <span className="text-xs text-neutral-500">Refreshing...</span>}
        </div>
        <pre className="whitespace-pre-wrap rounded-lg border border-neutral-800 bg-neutral-950 px-4 py-3 text-xs text-neutral-200">
          {rules || "ルールがありません。"}
        </pre>
      </section>
    </div>
  );
}
