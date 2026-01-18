type PeerConfigSectionProps = {
  peerConfig: string;
  peerId: string;
  hasPeer: boolean;
  actionLabel: string | null;
  statusLabel: string | null;
  statusToneClassName: string;
  onAction?: () => void;
  actionDisabled?: boolean;
  onDownload?: () => void;
};

export function PeerConfigSection({
  peerConfig,
  peerId,
  hasPeer,
  actionLabel,
  statusLabel,
  statusToneClassName,
  onAction,
  actionDisabled,
  onDownload,
}: PeerConfigSectionProps) {
  const showConfig = peerConfig.length > 0;
  const emptyMessage = hasPeer ? "まだConfigを表示していません。" : "まだPeerは作成されていません。";

  return (
    <section className="rounded-xl border border-neutral-800 bg-neutral-900/60 p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">Peer Config</h2>
          <p className="mt-2 text-sm text-neutral-400">生成された設定はQRコードなどに変換して配布できます。</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {actionLabel && (
            <button
              className="rounded-md bg-neutral-200 px-4 py-2 text-sm text-neutral-900"
              onClick={onAction}
              disabled={actionDisabled}
            >
              {actionLabel}
            </button>
          )}
          {showConfig && (
            <button
              className="rounded-md border border-neutral-600 px-4 py-2 text-sm text-neutral-100"
              onClick={onDownload}
            >
              Download
            </button>
          )}
        </div>
      </div>
      <div className="mt-4 rounded-lg border border-neutral-800 bg-neutral-950 p-4">
        <pre className="whitespace-pre-wrap text-xs text-neutral-100">
          {showConfig ? peerConfig : emptyMessage}
        </pre>
      </div>
      {peerId && statusLabel ? (
        <div className="mt-2 flex items-center gap-2 text-xs text-neutral-400">
          <span className={`text-base ${statusToneClassName}`}>•</span>
          <span>{statusLabel}</span>
        </div>
      ) : null}
    </section>
  );
}
