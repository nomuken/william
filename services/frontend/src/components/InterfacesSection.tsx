import type { WireguardInterface } from "@/gen/proto/server/v1/server_pb";

type InterfacesSectionProps = {
  interfaces: WireguardInterface[];
  selectedInterface: string;
  loading: boolean;
  onSelectInterface: (value: string) => void;
};

export function InterfacesSection({
  interfaces,
  selectedInterface,
  loading,
  onSelectInterface,
}: InterfacesSectionProps) {
  const interfaceList = Array.isArray(interfaces) ? interfaces : [];
  return (
    <section className="rounded-xl border border-neutral-800 bg-neutral-900/60 p-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Wireguard Interfaces</h2>
      </div>
      <div className="mt-4 space-y-3">
        {interfaceList.length === 0 ? (
          <p className="text-sm text-neutral-400">Interfaceが見つかりません。</p>
        ) : (
          interfaceList.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => onSelectInterface(item.id)}
              disabled={loading}
              className={`w-full rounded-lg border px-4 py-3 text-left text-sm transition ${
                selectedInterface === item.id
                  ? "border-neutral-500 bg-neutral-900/80"
                  : "border-neutral-800 bg-neutral-950 hover:border-neutral-600"
              }`}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium">{item.name}</span>
                <span className="text-xs text-neutral-400">{item.address}</span>
              </div>
              <div className="mt-2 text-xs text-neutral-400">
                Listen: {item.listenPort} / MTU {item.mtu}
              </div>
            </button>
          ))
        )}
      </div>
    </section>
  );
}
