type ConnectionSectionProps = {
  email: string;
  loading: boolean;
  onEmailChange: (value: string) => void;
  onLoadMyPeer: () => void;
  onDeletePeer: () => void;
};

export function ConnectionSection({
  email,
  loading,
  onEmailChange,
  onLoadMyPeer,
  onDeletePeer,
}: ConnectionSectionProps) {
  return (
    <section className="rounded-xl border border-neutral-800 bg-neutral-900/60 p-6">
      <h2 className="text-lg font-semibold">接続情報</h2>
      <div className="mt-4 grid gap-4 md:grid-cols-2">
        <label className="flex flex-col gap-2 text-sm">
          X-Email
          <input
            className="rounded-md border border-neutral-700 bg-neutral-950 px-3 py-2 text-sm"
            placeholder="you@example.com"
            value={email}
            onChange={(event) => onEmailChange(event.target.value)}
          />
        </label>
        <div className="flex items-end gap-2">
          <button
            type="button"
            className="rounded-md bg-neutral-200 px-4 py-2 text-sm font-medium text-neutral-900 hover:bg-neutral-100"
            onClick={onLoadMyPeer}
            disabled={loading}
          >
            自分のPeerを取得
          </button>
          <button
            type="button"
            className="rounded-md border border-neutral-700 px-4 py-2 text-sm font-medium text-neutral-200 hover:border-neutral-500"
            onClick={onDeletePeer}
            disabled={loading}
          >
            Peer削除
          </button>
        </div>
      </div>
    </section>
  );
}
