"use client";

import { useEffect, useState } from "react";
import { createAdminClient } from "@/lib/adminClient";
import { useAdminInterfaces } from "@/hooks/interfaces";
import { useAdminPeers } from "@/hooks/peers";
import { usePeerRoutes } from "@/hooks/peerRoutes";
import { normalizeError } from "@/lib/normalizeError";

export default function PeersPage() {
  const client = createAdminClient();
  const { interfaces, error: interfacesError } = useAdminInterfaces();
  const [selectedInterfaceId, setSelectedInterfaceId] = useState("");
  const [selectedPeerId, setSelectedPeerId] = useState("");
  const [newPeerRoute, setNewPeerRoute] = useState("");
  const [isRoutesModalOpen, setIsRoutesModalOpen] = useState(false);
  const { peers, error: peersError, isLoading: peersLoading, mutate: mutatePeers } =
    useAdminPeers(selectedInterfaceId);
  const {
    routes: peerRoutes,
    error: peerRoutesError,
    isLoading: peerRoutesLoading,
    mutate: mutatePeerRoutes,
  } = usePeerRoutes(selectedPeerId);
  const [actionError, setActionError] = useState("");
  const [isMutating, setIsMutating] = useState(false);

  useEffect(() => {
    if (selectedPeerId && !peers.some((item) => item.peerId === selectedPeerId)) {
      setSelectedPeerId("");
      setIsRoutesModalOpen(false);
    }
  }, [peers, selectedPeerId]);

  const error =
    actionError ||
    (peersError ?? peerRoutesError ?? interfacesError
      ? normalizeError(peersError ?? peerRoutesError ?? interfacesError)
      : "");
  const loading = isMutating || peersLoading || peerRoutesLoading;

  const handleDeletePeer = async (peerId: string) => {
    setIsMutating(true);
    setActionError("");
    try {
      await client.deletePeer({ peerId });
      await mutatePeers();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleOpenRoutesModal = (peerId: string) => {
    setSelectedPeerId(peerId);
    setNewPeerRoute("");
    setActionError("");
    setIsRoutesModalOpen(true);
  };

  const handleCloseRoutesModal = () => {
    setIsRoutesModalOpen(false);
    setSelectedPeerId("");
    setNewPeerRoute("");
    setActionError("");
  };

  const handleCreatePeerRoute = async () => {
    if (!selectedPeerId) {
      setActionError("Peerを選択してください。");
      return;
    }
    if (!window.confirm("このRouteを追加します。よろしいですか？")) {
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.createPeerRoute({ peerId: selectedPeerId, cidr: newPeerRoute });
      setNewPeerRoute("");
      await mutatePeerRoutes();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleDeletePeerRoute = async (cidr: string) => {
    if (!selectedPeerId) {
      setActionError("Peerを選択してください。");
      return;
    }
    if (!window.confirm("このRouteを削除します。よろしいですか？")) {
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.deletePeerRoute({ peerId: selectedPeerId, cidr });
      await mutatePeerRoutes();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">Peers</h1>
        <p className="text-sm text-neutral-400">Peerの管理と許可ルート設定を行います。</p>
      </header>

      {error && <div className="rounded-md border border-red-400/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">{error}</div>}

      <section className="space-y-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold">Peers</h2>
          <select
            className="rounded-md bg-neutral-950 px-3 py-2 text-sm"
            value={selectedInterfaceId}
            onChange={(event) => setSelectedInterfaceId(event.target.value)}
          >
            <option value="">All Interfaces</option>
            {interfaces.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name || item.id}
              </option>
            ))}
          </select>
        </div>

        <div className="space-y-2 text-sm">
          {peers.length === 0 && <p className="text-neutral-400">Peerがありません。</p>}
          {peers.map((peer) => (
            <div key={peer.peerId} className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-neutral-800 bg-neutral-950 px-4 py-2">
              <div className="space-y-1">
                <div className="font-medium">{peer.email}</div>
                <div className="text-xs text-neutral-400">PubKey: {peer.peerId}</div>
                <div className="text-xs text-neutral-400">Allowed IP: {peer.allowedIp}</div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  className="rounded-md border border-neutral-600 px-3 py-1"
                  onClick={() => handleOpenRoutesModal(peer.peerId)}
                  disabled={loading}
                >
                  Routes
                </button>
                <button
                  className="rounded-md border border-red-400 px-3 py-1 text-red-200"
                  onClick={() => handleDeletePeer(peer.peerId)}
                  disabled={loading}
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      </section>

      {isRoutesModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4">
          <div className="w-full max-w-2xl space-y-4 rounded-xl border border-neutral-800 bg-neutral-950 p-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">Peer Routes</h2>
                <p className="text-xs text-neutral-400">PubKey: {selectedPeerId}</p>
              </div>
              <button className="text-sm text-neutral-400" onClick={handleCloseRoutesModal}>
                Close
              </button>
            </div>

            <div className="flex flex-wrap gap-2">
              <input
                className="flex-1 rounded-md bg-neutral-900 px-3 py-2 text-sm"
                placeholder="CIDR (例: 10.0.0.0/24)"
                value={newPeerRoute}
                onChange={(event) => setNewPeerRoute(event.target.value)}
              />
              <button
                className="rounded-md bg-neutral-200 px-4 py-2 text-sm text-neutral-900"
                onClick={handleCreatePeerRoute}
                disabled={loading}
              >
                Add Route
              </button>
            </div>

            <div className="space-y-2 text-sm">
              {peerRoutes.length === 0 && <p className="text-neutral-400">Routeはありません。</p>}
              {peerRoutes.map((route) => (
                <div
                  key={`${route.peerId}-${route.cidr}`}
                  className="flex items-center justify-between rounded-md border border-neutral-800 bg-neutral-900 px-4 py-2"
                >
                  <span>{route.cidr}</span>
                  <button
                    className="rounded-md border border-red-400 px-3 py-1 text-red-200"
                    onClick={() => handleDeletePeerRoute(route.cidr)}
                    disabled={loading}
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>

            <div className="flex justify-end">
              <button
                className="rounded-md border border-neutral-700 px-4 py-2 text-sm"
                onClick={handleCloseRoutesModal}
                disabled={loading}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
