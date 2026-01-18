"use client";

import { Code, ConnectError } from "@connectrpc/connect";
import { useEffect, useMemo, useState } from "react";
import { ConnectionSection } from "@/components/ConnectionSection";
import { ErrorBanner } from "@/components/ErrorBanner";
import { InterfacesSection } from "@/components/InterfacesSection";
import { PeerConfigSection } from "@/components/PeerConfigSection";
import { usePeerStatuses } from "@/hooks/peerStatuses";
import { useWireguardInterfaces } from "@/hooks/interfaces";
import { normalizeError } from "@/lib/normalizeError";
import { createWilliamClient } from "@/lib/williamClient";

const emptyRequest = {};

export default function Home() {
  const client = useMemo(() => createWilliamClient(), []);
  const [email, setEmail] = useState("");
  const [selectedInterface, setSelectedInterface] = useState("");
  const [peerConfig, setPeerConfig] = useState("");
  const [peerId, setPeerId] = useState("");
  const [peerExists, setPeerExists] = useState<boolean | null>(null);
  const [actionError, setActionError] = useState("");
  const [isMutating, setIsMutating] = useState(false);
  const {
    interfaces,
    error: interfacesError,
    isLoading: interfacesLoading,
  } = useWireguardInterfaces(email);

  const requestHeaders = useMemo(() => (email ? new Headers({ "X-Email": email }) : undefined), [email]);
  const interfaceList = Array.isArray(interfaces) ? interfaces : [];
  const showConnectionSection = process.env.NODE_ENV === "development";
  const requestOptions = requestHeaders ? { headers: requestHeaders } : undefined;
  const statusEnabled = Boolean(peerId) && (!showConnectionSection || email.length > 0);
  const { statuses, error: statusesError } = usePeerStatuses(statusEnabled, requestOptions);
  const error = actionError || (interfacesError || statusesError ? normalizeError(interfacesError ?? statusesError) : "");
  const loading = isMutating || interfacesLoading;

  useEffect(() => {
    if (showConnectionSection && !email) {
      setSelectedInterface("");
      setPeerExists(null);
      setPeerConfig("");
      setPeerId("");
      return;
    }
    if (!selectedInterface && interfaceList.length > 0) {
      setSelectedInterface(interfaceList[0].id);
      return;
    }
    if (selectedInterface && !interfaceList.some((item) => item.id === selectedInterface)) {
      setSelectedInterface("");
    }
  }, [email, interfaceList, selectedInterface, showConnectionSection]);

  const fetchPeerForInterface = async (displayConfig: boolean) => {
    if (!selectedInterface) {
      setPeerExists(null);
      return;
    }
    if (showConnectionSection && !email) {
      setActionError("X-Emailに設定するメールを入力してください。");
      return;
    }

    setIsMutating(true);
    setActionError("");
    try {
      const response = await client.getMyWireguardPeerByInterface(
        { interfaceId: selectedInterface },
        requestOptions,
      );
      setPeerExists(true);
      if (displayConfig) {
        setPeerConfig(response.peerConfig);
        setPeerId(response.peerId);
      }
    } catch (err) {
      if (err instanceof ConnectError && err.code === Code.NotFound) {
        setPeerExists(false);
        if (displayConfig) {
          setPeerConfig("");
          setPeerId("");
        }
        return;
      }
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  useEffect(() => {
    if (!selectedInterface) {
      setPeerExists(null);
      setPeerConfig("");
      setPeerId("");
      return;
    }
    setPeerConfig("");
    setPeerId("");
    void fetchPeerForInterface(true);
  }, [selectedInterface, email, showConnectionSection]);

  const handleCreatePeer = async () => {
    if (!selectedInterface) {
      setActionError("Interfaceを選択してください。");
      return;
    }
    if (showConnectionSection && !email) {
      setActionError("X-Emailに設定するメールを入力してください。");
      return;
    }

    setIsMutating(true);
    setActionError("");
    try {
      const response = await client.createWireguardPeer(
        { wireguardInterfaceId: selectedInterface },
        requestOptions,
      );
      setPeerConfig(response.peerConfig);
      setPeerId(response.peerId);
      setPeerExists(true);
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleLoadMyPeer = async () => {
    if (showConnectionSection && !email) {
      setActionError("X-Emailに設定するメールを入力してください。");
      return;
    }

    setIsMutating(true);
    setActionError("");
    try {
      const response = await client.getMyWireguardPeer(emptyRequest, requestOptions);
      setPeerConfig(response.peerConfig);
      setPeerId(response.peerId);
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleDeletePeer = async () => {
    if (!peerId) {
      setActionError("削除対象のPeer IDがありません。");
      return;
    }
    if (showConnectionSection && !email) {
      setActionError("X-Emailに設定するメールを入力してください。");
      return;
    }

    setIsMutating(true);
    setActionError("");
    try {
      await client.deleteWireguardPeer({ peerId }, requestOptions);
      setPeerConfig("");
      setPeerId("");
      setPeerExists(false);
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleDownloadPeerConfig = () => {
    if (!peerConfig) {
      return;
    }
    const blob = new Blob([peerConfig], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "peer.conf";
    link.click();
    URL.revokeObjectURL(url);
  };

  const hasPeer = peerExists === true;
  const actionLabel = peerExists === false ? "Create Peer" : null;

  const formatBytes = (bytes: number) => {
    if (!Number.isFinite(bytes) || bytes <= 0) {
      return "0 B";
    }
    const units = ["B", "KB", "MB", "GB", "TB"];
    let index = 0;
    let value = bytes;
    while (value >= 1024 && index < units.length - 1) {
      value /= 1024;
      index += 1;
    }
    return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
  };

  const peerStatus = statuses.find((status) => status.peerId === peerId);
  const hasRecentHandshake = peerStatus ? Date.now() / 1000 - Number(peerStatus.lastHandshakeAt) <= 600 : false;
  const statusToneClassName = hasRecentHandshake ? "text-emerald-300" : "text-amber-300";
  const statusLabel = peerStatus
    ? `↑:${formatBytes(Number(peerStatus.txBytes))} / ↓:${formatBytes(Number(peerStatus.rxBytes))}`
    : null;

  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-100">
      <main className="mx-auto flex w-full max-w-5xl flex-col gap-8 px-6 py-10">
        <header className="space-y-2">
          <h1 className="text-3xl font-semibold">William Wireguard Console</h1>
          <p className="text-sm text-neutral-400">Connect RPCでWireguard InterfaceとPeerを操作します。</p>
        </header>

        {showConnectionSection && (
          <ConnectionSection
            email={email}
            loading={loading}
            onEmailChange={setEmail}
            onLoadMyPeer={handleLoadMyPeer}
            onDeletePeer={handleDeletePeer}
          />
        )}

        <InterfacesSection
          interfaces={interfaceList}
          selectedInterface={selectedInterface}
          onSelectInterface={setSelectedInterface}
          loading={interfacesLoading}
        />


        <PeerConfigSection
          peerConfig={peerConfig}
          peerId={peerId}
          hasPeer={hasPeer}
          actionLabel={actionLabel}
          statusLabel={statusLabel}
          statusToneClassName={statusToneClassName}
          onAction={handleCreatePeer}
          actionDisabled={loading || !selectedInterface}
          onDownload={handleDownloadPeerConfig}
        />

        <ErrorBanner message={error} />
      </main>
    </div>
  );
}
