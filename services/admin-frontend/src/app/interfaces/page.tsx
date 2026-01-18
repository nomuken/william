"use client";

import { useEffect, useMemo, useState } from "react";
import { createAdminClient } from "@/lib/adminClient";
import { useAdminInterfaces } from "@/hooks/interfaces";
import { useInterfaceRoutes } from "@/hooks/interfaceRoutes";
import { normalizeError } from "@/lib/normalizeError";

export default function InterfacesPage() {
  const client = useMemo(() => createAdminClient(), []);
  const { interfaces, error: interfacesError, isLoading: interfacesLoading, mutate: mutateInterfaces } =
    useAdminInterfaces();
  const [selectedInterfaceId, setSelectedInterfaceId] = useState("");
  const {
    routes: interfaceRoutes,
    error: interfaceRoutesError,
    isLoading: interfaceRoutesLoading,
    mutate: mutateInterfaceRoutes,
  } = useInterfaceRoutes(selectedInterfaceId);
  const [newInterfaceRoute, setNewInterfaceRoute] = useState("");
  const [actionError, setActionError] = useState("");
  const [isMutating, setIsMutating] = useState(false);
  const [isRoutesModalOpen, setIsRoutesModalOpen] = useState(false);

  const [createForm, setCreateForm] = useState({
    name: "",
    address: "",
    listenPort: "",
    mtu: "",
    endpoint: "",
  });
  const emptyUpdateForm = {
    id: "",
    name: "",
    address: "",
    listenPort: "",
    mtu: "",
    endpoint: "",
  };
  const [updateForm, setUpdateForm] = useState(emptyUpdateForm);
  const [isUpdateModalOpen, setIsUpdateModalOpen] = useState(false);

  const firstError = interfacesError ?? interfaceRoutesError;
  const error = actionError || (firstError ? normalizeError(firstError) : "");
  const loading = isMutating || interfacesLoading || interfaceRoutesLoading;

  useEffect(() => {
    if (!selectedInterfaceId && interfaces.length > 0) {
      setSelectedInterfaceId(interfaces[0].id);
      return;
    }
    if (selectedInterfaceId && !interfaces.some((item) => item.id === selectedInterfaceId)) {
      setSelectedInterfaceId("");
    }
  }, [interfaces, selectedInterfaceId]);

  const handleOpenUpdateModal = (item: (typeof interfaces)[number]) => {
    setUpdateForm({
      id: item.id,
      name: item.name ?? "",
      address: item.address ?? "",
      listenPort: String(item.listenPort ?? ""),
      mtu: String(item.mtu ?? ""),
      endpoint: item.endpoint ?? "",
    });
    setSelectedInterfaceId(item.id);
    setActionError("");
    setIsUpdateModalOpen(true);
  };

  const handleCloseUpdateModal = () => {
    setIsUpdateModalOpen(false);
    setUpdateForm(emptyUpdateForm);
    setActionError("");
  };

  const handleOpenRoutesModal = (item: (typeof interfaces)[number]) => {
    setSelectedInterfaceId(item.id);
    setNewInterfaceRoute("");
    setActionError("");
    setIsRoutesModalOpen(true);
  };

  const handleCloseRoutesModal = () => {
    setIsRoutesModalOpen(false);
    setNewInterfaceRoute("");
    setActionError("");
  };

  const handleCreateInterface = async () => {
    setIsMutating(true);
    setActionError("");
    try {
      const response = await client.createInterface({
        name: createForm.name,
        address: createForm.address,
        listenPort: Number(createForm.listenPort),
        mtu: Number(createForm.mtu),
        endpoint: createForm.endpoint,
      });
      setCreateForm({ name: "", address: "", listenPort: "", mtu: "", endpoint: "" });
      await mutateInterfaces();
      if (response.interface?.id) {
        setSelectedInterfaceId(response.interface.id);
      }
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleUpdateInterface = async () => {
    if (!updateForm.id) {
      setActionError("Interfaceを選択してください。");
      return;
    }
    if (!window.confirm("このInterfaceを更新します。よろしいですか？")) {
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.updateInterface({
        id: updateForm.id,
        name: updateForm.name,
        address: updateForm.address,
        listenPort: Number(updateForm.listenPort),
        mtu: Number(updateForm.mtu),
        endpoint: updateForm.endpoint,
      });
      setUpdateForm(emptyUpdateForm);
      setIsUpdateModalOpen(false);
      await mutateInterfaces();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleDeleteInterface = async (interfaceId: string) => {
    if (!window.confirm("このInterfaceを削除します。よろしいですか？")) {
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.deleteInterface({ id: interfaceId });
      if (selectedInterfaceId === interfaceId) {
        setSelectedInterfaceId("");
      }
      await mutateInterfaces();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleCreateInterfaceRoute = async () => {
    if (!selectedInterfaceId) {
      setActionError("Interfaceを選択してください。");
      return;
    }
    if (!window.confirm("このRouteを追加します。よろしいですか？")) {
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.createInterfaceRoute({
        interfaceId: selectedInterfaceId,
        cidr: newInterfaceRoute,
      });
      setNewInterfaceRoute("");
      await mutateInterfaceRoutes();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleDeleteInterfaceRoute = async (cidr: string) => {
    if (!selectedInterfaceId) {
      setActionError("Interfaceを選択してください。");
      return;
    }
    if (!window.confirm("このRouteを削除します。よろしいですか？")) {
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.deleteInterfaceRoute({ interfaceId: selectedInterfaceId, cidr });
      await mutateInterfaceRoutes();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">Interfaces</h1>
        <p className="text-sm text-neutral-400">Wireguard interfaceとルートを管理します。</p>
      </header>

      {error && <div className="rounded-md border border-red-400/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">{error}</div>}

      <section className="space-y-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
        <h2 className="text-lg font-semibold">Create Interface</h2>
        <div className="grid gap-3 md:grid-cols-5">
          <input
            className="rounded-md bg-neutral-950 px-3 py-2 text-sm"
            placeholder="Name"
            value={createForm.name}
            onChange={(event) => setCreateForm({ ...createForm, name: event.target.value })}
          />
          <input
            className="rounded-md bg-neutral-950 px-3 py-2 text-sm"
            placeholder="CIDR Address"
            value={createForm.address}
            onChange={(event) => setCreateForm({ ...createForm, address: event.target.value })}
          />
          <input
            className="rounded-md bg-neutral-950 px-3 py-2 text-sm"
            placeholder="Listen Port"
            value={createForm.listenPort}
            onChange={(event) => setCreateForm({ ...createForm, listenPort: event.target.value })}
          />
          <input
            className="rounded-md bg-neutral-950 px-3 py-2 text-sm"
            placeholder="MTU"
            value={createForm.mtu}
            onChange={(event) => setCreateForm({ ...createForm, mtu: event.target.value })}
          />
          <input
            className="rounded-md bg-neutral-950 px-3 py-2 text-sm"
            placeholder="Endpoint"
            value={createForm.endpoint}
            onChange={(event) => setCreateForm({ ...createForm, endpoint: event.target.value })}
          />
        </div>
        <button
          className="rounded-md bg-neutral-200 px-4 py-2 text-sm text-neutral-900"
          onClick={handleCreateInterface}
          disabled={loading}
        >
          Create Interface
        </button>
      </section>

      <section className="space-y-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
        <h2 className="text-lg font-semibold">Interfaces</h2>
        <div className="space-y-2 text-sm">
          {interfaces.length === 0 && <p className="text-neutral-400">Interfaceがありません。</p>}
          {interfaces.map((item) => (
            <div key={item.id} className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-neutral-800 bg-neutral-950 px-4 py-2">
              <div className="space-y-1">
                <div className="font-medium">{item.name}</div>
                <div className="text-xs text-neutral-400">ID: {item.id}</div>
                <div className="text-xs text-neutral-400">
                  {item.address} / port {item.listenPort} / mtu {item.mtu}
                </div>
                <div className="text-xs text-neutral-400">Endpoint: {item.endpoint}</div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  className="rounded-md border border-neutral-600 px-3 py-1"
                  onClick={() => handleOpenUpdateModal(item)}
                  disabled={loading}
                >
                  Edit
                </button>
                <button
                  className="rounded-md border border-neutral-600 px-3 py-1"
                  onClick={() => handleOpenRoutesModal(item)}
                  disabled={loading}
                >
                  Routes
                </button>
                <button
                  className="rounded-md border border-red-400 px-3 py-1 text-red-200"
                  onClick={() => handleDeleteInterface(item.id)}
                  disabled={loading}
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      </section>

      {isUpdateModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4">
          <div className="w-full max-w-2xl space-y-4 rounded-xl border border-neutral-800 bg-neutral-950 p-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">Update Interface</h2>
                <p className="text-xs text-neutral-400">ID: {updateForm.id}</p>
              </div>
              <button className="text-sm text-neutral-400" onClick={handleCloseUpdateModal}>
                Close
              </button>
            </div>
            <div className="grid gap-3 md:grid-cols-2">
              <input
                className="rounded-md bg-neutral-900 px-3 py-2 text-sm"
                placeholder="Name"
                value={updateForm.name}
                onChange={(event) => setUpdateForm({ ...updateForm, name: event.target.value })}
              />
              <input
                className="rounded-md bg-neutral-900 px-3 py-2 text-sm"
                placeholder="CIDR Address"
                value={updateForm.address}
                onChange={(event) => setUpdateForm({ ...updateForm, address: event.target.value })}
              />
              <input
                className="rounded-md bg-neutral-900 px-3 py-2 text-sm"
                placeholder="Listen Port"
                value={updateForm.listenPort}
                onChange={(event) => setUpdateForm({ ...updateForm, listenPort: event.target.value })}
              />
              <input
                className="rounded-md bg-neutral-900 px-3 py-2 text-sm"
                placeholder="MTU"
                value={updateForm.mtu}
                onChange={(event) => setUpdateForm({ ...updateForm, mtu: event.target.value })}
              />
              <input
                className="rounded-md bg-neutral-900 px-3 py-2 text-sm md:col-span-2"
                placeholder="Endpoint"
                value={updateForm.endpoint}
                onChange={(event) => setUpdateForm({ ...updateForm, endpoint: event.target.value })}
              />
            </div>
            <div className="flex items-center justify-end gap-2">
              <button
                className="rounded-md border border-neutral-700 px-4 py-2 text-sm"
                onClick={handleCloseUpdateModal}
                disabled={loading}
              >
                Cancel
              </button>
              <button
                className="rounded-md bg-neutral-200 px-4 py-2 text-sm text-neutral-900"
                onClick={handleUpdateInterface}
                disabled={loading}
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}

      {isRoutesModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4">
          <div className="w-full max-w-2xl space-y-4 rounded-xl border border-neutral-800 bg-neutral-950 p-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">Interface Routes</h2>
                <p className="text-xs text-neutral-400">ID: {selectedInterfaceId}</p>
              </div>
              <button className="text-sm text-neutral-400" onClick={handleCloseRoutesModal}>
                Close
              </button>
            </div>

            <div className="flex flex-wrap gap-2">
              <input
                className="flex-1 rounded-md bg-neutral-900 px-3 py-2 text-sm"
                placeholder="CIDR (例: 10.0.0.0/24)"
                value={newInterfaceRoute}
                onChange={(event) => setNewInterfaceRoute(event.target.value)}
              />
              <button
                className="rounded-md bg-neutral-200 px-4 py-2 text-sm text-neutral-900"
                onClick={handleCreateInterfaceRoute}
                disabled={loading}
              >
                Add Route
              </button>
            </div>

            <div className="space-y-2 text-sm">
              {interfaceRoutes.length === 0 && <p className="text-neutral-400">Routeはありません。</p>}
              {interfaceRoutes.map((route) => (
                <div
                  key={`${route.interfaceId}-${route.cidr}`}
                  className="flex items-center justify-between rounded-md border border-neutral-800 bg-neutral-900 px-4 py-2"
                >
                  <span>{route.cidr}</span>
                  <button
                    className="rounded-md border border-red-400 px-3 py-1 text-red-200"
                    onClick={() => handleDeleteInterfaceRoute(route.cidr)}
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
