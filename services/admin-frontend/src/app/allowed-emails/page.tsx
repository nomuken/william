"use client";

import { useEffect, useState } from "react";
import { createAdminClient } from "@/lib/adminClient";
import { useAllowedEmails } from "@/hooks/allowedEmails";
import { useAdminInterfaces } from "@/hooks/interfaces";
import { normalizeError } from "@/lib/normalizeError";

export default function AllowedEmailsPage() {
  const client = createAdminClient();
  const { interfaces, error: interfacesError } = useAdminInterfaces();
  const [selectedInterfaceId, setSelectedInterfaceId] = useState("");
  const {
    emails: allowedEmails,
    error: allowedEmailsError,
    isLoading: allowedEmailsLoading,
    mutate: mutateAllowedEmails,
  } = useAllowedEmails(selectedInterfaceId);
  const [newAllowedEmail, setNewAllowedEmail] = useState("");
  const [actionError, setActionError] = useState("");
  const [isMutating, setIsMutating] = useState(false);

  useEffect(() => {
    if (!selectedInterfaceId && interfaces.length > 0) {
      setSelectedInterfaceId(interfaces[0].id);
      return;
    }
    if (selectedInterfaceId && !interfaces.some((item) => item.id === selectedInterfaceId)) {
      setSelectedInterfaceId("");
    }
  }, [interfaces, selectedInterfaceId]);

  const error = actionError || (interfacesError ?? allowedEmailsError ? normalizeError(interfacesError ?? allowedEmailsError) : "");
  const loading = isMutating || allowedEmailsLoading;

  const handleCreateAllowedEmail = async () => {
    if (!selectedInterfaceId) {
      setActionError("Interfaceを選択してください。");
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.createAllowedEmail({
        interfaceId: selectedInterfaceId,
        email: newAllowedEmail,
      });
      setNewAllowedEmail("");
      await mutateAllowedEmails();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  const handleDeleteAllowedEmail = async (email: string) => {
    if (!selectedInterfaceId) {
      setActionError("Interfaceを選択してください。");
      return;
    }
    setIsMutating(true);
    setActionError("");
    try {
      await client.deleteAllowedEmail({ interfaceId: selectedInterfaceId, email });
      await mutateAllowedEmails();
    } catch (err) {
      setActionError(normalizeError(err));
    } finally {
      setIsMutating(false);
    }
  };

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">Allowed Emails</h1>
        <p className="text-sm text-neutral-400">Interface単位の許可Emailを管理します。</p>
      </header>

      {error && <div className="rounded-md border border-red-400/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">{error}</div>}

      <section className="space-y-4 rounded-xl border border-neutral-800 bg-neutral-900/40 p-6">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Select Interface</h2>
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

        <div className="flex flex-wrap gap-2">
          <input
            className="flex-1 rounded-md bg-neutral-950 px-3 py-2 text-sm"
            placeholder="email@example.com"
            value={newAllowedEmail}
            onChange={(event) => setNewAllowedEmail(event.target.value)}
          />
          <button
            className="rounded-md bg-neutral-200 px-4 py-2 text-sm text-neutral-900"
            onClick={handleCreateAllowedEmail}
            disabled={loading}
          >
            Add Email
          </button>
        </div>

        <div className="space-y-2 text-sm">
          {allowedEmails.length === 0 && <p className="text-neutral-400">登録済みEmailはありません。</p>}
          {allowedEmails.map((item) => (
            <div
              key={`${item.interfaceId}-${item.email}`}
              className="flex items-center justify-between rounded-md border border-neutral-800 bg-neutral-950 px-4 py-2"
            >
              <span>{item.email}</span>
              <button
                className="rounded-md border border-red-400 px-3 py-1 text-red-200"
                onClick={() => handleDeleteAllowedEmail(item.email)}
                disabled={loading}
              >
                Remove
              </button>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
