"use client";

import { useMemo } from "react";
import { useAdminInterfaces } from "@/hooks/interfaces";
import { useAdminPeers } from "@/hooks/peers";
import { usePeerStats } from "@/hooks/peerStats";
import { normalizeError } from "@/lib/normalizeError";

function formatBytes(value: number) {
  if (value < 1024) {
    return `${value} B`;
  }
  const units = ["KB", "MB", "GB", "TB"];
  let size = value / 1024;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`;
}

function formatHandshake(timestamp: number) {
  if (!timestamp) {
    return { label: "never", className: "text-neutral-400" };
  }
  const now = Math.floor(Date.now() / 1000);
  const diff = Math.max(0, now - timestamp);
  if (diff < 60) {
    return { label: `${diff}s ago`, className: "text-emerald-300" };
  }
  if (diff < 3600) {
    return { label: `${Math.floor(diff / 60)}m ago`, className: "text-amber-300" };
  }
  return { label: `${Math.floor(diff / 3600)}h ago`, className: "text-neutral-400" };
}

export default function StatusPage() {
  const { interfaces, error: interfacesError } = useAdminInterfaces();
  const { peers, error: peersError } = useAdminPeers("");
  const { stats, error: statsError, isLoading } = usePeerStats();
  const interfaceNameById = useMemo(() => {
    const map = new Map<string, string>();
    interfaces.forEach((item) => map.set(item.id, item.name));
    return map;
  }, [interfaces]);
  const peerById = useMemo(() => {
    const map = new Map<string, { email: string; interfaceId: string }>();
    peers.forEach((peer) => {
      map.set(peer.peerId, { email: peer.email, interfaceId: peer.interfaceId });
    });
    return map;
  }, [peers]);

  const message = statsError || peersError || interfacesError ? normalizeError(statsError ?? peersError ?? interfacesError) : "";

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">Status</h1>
        <p className="text-sm text-neutral-400">Peerの通信量とHandshakeを表示します。</p>
      </header>

      {message && <div className="rounded-md border border-red-400/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">{message}</div>}

      <section className="rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Peer Status</h2>
          {isLoading && <span className="text-xs text-neutral-500">Refreshing...</span>}
        </div>

        {stats.length === 0 ? (
          <p className="text-sm text-neutral-400">Peerが見つかりません。</p>
        ) : (
          <div className="space-y-2 text-sm">
            {stats.map((stat) => {
              const handshake = formatHandshake(stat.lastHandshakeAt);
              const peerInfo = peerById.get(stat.peerId);
              const interfaceId = peerInfo?.interfaceId ?? stat.interfaceId;
              const interfaceName = interfaceNameById.get(interfaceId) ?? interfaceId;
              return (
                <div
                  key={stat.peerId}
                  className="flex flex-wrap items-center justify-between gap-4 rounded-md border border-neutral-800 bg-neutral-950 px-4 py-3"
                >
                  <div className="space-y-1">
                    <div className="font-medium text-neutral-100">
                      {peerInfo?.email ?? "Unknown Email"} · {interfaceName}
                    </div>
                    <div className="text-xs text-neutral-400">PubKey: {stat.peerId}</div>
                  </div>
                  <div className="flex flex-wrap gap-6 text-xs text-neutral-300">
                    <span>RX: {formatBytes(Number(stat.rxBytes))}</span>
                    <span>TX: {formatBytes(Number(stat.txBytes))}</span>
                    <span className={handshake.className}>Handshake: {handshake.label}</span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
