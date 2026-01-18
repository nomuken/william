"use client";

import { useEffect, useState } from "react";

import { useAdminInterfaces } from "@/hooks/interfaces";
import { useWireguardConfigs } from "@/hooks/wireguardConfigs";
import { normalizeError } from "@/lib/normalizeError";

export default function ConfigPage() {
  const { interfaces, error: interfacesError } = useAdminInterfaces();
  const [selectedInterfaceId, setSelectedInterfaceId] = useState("");
  const { configs, error: configsError, isLoading } = useWireguardConfigs(selectedInterfaceId);

  useEffect(() => {
    if (!selectedInterfaceId && interfaces.length > 0) {
      setSelectedInterfaceId(interfaces[0].id);
      return;
    }
    if (selectedInterfaceId && !interfaces.some((item) => item.id === selectedInterfaceId)) {
      setSelectedInterfaceId("");
    }
  }, [interfaces, selectedInterfaceId]);

  const message = interfacesError || configsError ? normalizeError(interfacesError ?? configsError) : "";
  const activeConfig = configs.find((config) => config.interfaceId === selectedInterfaceId);

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">Config</h1>
        <p className="text-sm text-neutral-400">現在のWireguard設定を表示します。</p>
      </header>

      {message && <div className="rounded-md border border-red-400/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">{message}</div>}

      <section className="space-y-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold">Interface Config</h2>
          <select
            className="rounded-md bg-neutral-950 px-3 py-2 text-sm"
            value={selectedInterfaceId}
            onChange={(event) => setSelectedInterfaceId(event.target.value)}
          >
            <option value="">Select interface</option>
            {interfaces.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </div>

        {isLoading && <p className="text-xs text-neutral-500">Loading config...</p>}
        {!isLoading && !selectedInterfaceId && (
          <p className="text-sm text-neutral-400">Interfaceを選択してください。</p>
        )}
        {!isLoading && selectedInterfaceId && !activeConfig && (
          <p className="text-sm text-neutral-400">Configが見つかりません。</p>
        )}
        {activeConfig && (
          <pre className="whitespace-pre-wrap rounded-lg border border-neutral-800 bg-neutral-950 px-4 py-3 text-xs text-neutral-200">
            {activeConfig.config}
          </pre>
        )}
      </section>
    </div>
  );
}
