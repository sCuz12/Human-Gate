"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { Session } from "@supabase/supabase-js";

import {
  listApprovalRequests,
  type ApprovalRequest,
} from "@/lib/api/approval-requests";
import { listWorkspaces, type Workspace } from "@/lib/api/workspaces";
import { getSupabaseBrowserClient } from "@/lib/supabase/client";

const statusLabels: Record<string, string> = {
  received: "Received",
  evaluating: "Evaluating",
  allowed: "Allowed",
  blocked: "Blocked",
  pending: "Pending",
  approved: "Approved",
  approved_with_changes: "Approved with changes",
  rejected: "Rejected",
  expired: "Expired",
  cancelled: "Cancelled",
};

export function InboxClient() {
  const router = useRouter();
  const [session, setSession] = useState<Session | null>(null);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [selectedWorkspaceID, setSelectedWorkspaceID] = useState("");
  const [requests, setRequests] = useState<ApprovalRequest[]>([]);
  const [statusFilter, setStatusFilter] = useState("all");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const supabase = getSupabaseBrowserClient();

    async function loadInitialState() {
      const {
        data: { session: currentSession },
      } = await supabase.auth.getSession();

      if (!currentSession) {
        router.replace("/login");
        return;
      }

      setSession(currentSession);
      await loadWorkspacesAndRequests(currentSession);
      setIsLoading(false);
    }

    void loadInitialState();
  }, [router]);

  async function loadWorkspacesAndRequests(activeSession: Session) {
    try {
      const workspaceResponse = await listWorkspaces(activeSession);
      setWorkspaces(workspaceResponse.workspaces);

      const firstWorkspace = workspaceResponse.workspaces[0];
      if (!firstWorkspace) {
        setRequests([]);
        return;
      }

      setSelectedWorkspaceID(firstWorkspace.id);
      await loadRequests(activeSession, firstWorkspace.id);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "Inbox could not be loaded.");
    }
  }

  async function loadRequests(activeSession: Session, workspaceID: string) {
    const response = await listApprovalRequests(activeSession, {
      workspaceID,
      limit: 50,
    });
    setRequests(response.requests);
    setError(null);
  }

  async function changeWorkspace(workspaceID: string) {
    setSelectedWorkspaceID(workspaceID);
    if (!session) {
      return;
    }

    try {
      await loadRequests(session, workspaceID);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "Requests could not be loaded.");
    }
  }

  const filteredRequests = useMemo(() => {
    if (statusFilter === "all") {
      return requests;
    }

    return requests.filter((request) => request.status === statusFilter);
  }, [requests, statusFilter]);

  if (isLoading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f4f5f0] px-6">
        <div className="rounded-lg border border-black/10 bg-white px-5 py-4 text-sm text-black/65 shadow-sm">
          Loading inbox...
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f4f5f0] text-[#15110d]">
      <header className="border-b border-black/10 bg-white">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-6 py-5 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase text-[#1f6f78]">
              Approval Inbox
            </p>
            <h1 className="mt-1 text-2xl font-semibold">Requests</h1>
          </div>

          <div className="flex flex-col gap-3 sm:flex-row">
            <Link
              className="rounded-md border border-black/10 bg-white px-3 py-2 text-sm font-medium text-black/70 transition hover:border-[#1f6f78] hover:text-[#1f6f78]"
              href="/dashboard"
            >
              Dashboard
            </Link>
            <button
              className="rounded-md bg-[#15110d] px-3 py-2 text-sm font-semibold text-white transition hover:bg-black"
              onClick={() => session && loadRequests(session, selectedWorkspaceID)}
              type="button"
            >
              Refresh
            </button>
          </div>
        </div>
      </header>

      <section className="mx-auto max-w-7xl px-6 py-6">
        {error ? (
          <div className="mb-5 rounded-lg border border-[#b74b2a]/20 bg-[#fff2ec] p-4 text-sm text-[#8d3419]">
            {error}
          </div>
        ) : null}

        {workspaces.length === 0 ? (
          <div className="rounded-lg border border-black/10 bg-white p-6 shadow-sm">
            <p className="text-sm text-black/65">
              Create a workspace before reviewing approval requests.
            </p>
          </div>
        ) : (
          <div className="grid gap-5 lg:grid-cols-[260px_1fr]">
            <aside className="space-y-4">
              <div className="rounded-lg border border-black/10 bg-white p-4 shadow-sm">
                <label className="block text-xs font-semibold uppercase text-black/45">
                  Workspace
                </label>
                <select
                  className="mt-2 w-full rounded-md border border-black/10 bg-white px-3 py-2 text-sm outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
                  onChange={(event) => void changeWorkspace(event.target.value)}
                  value={selectedWorkspaceID}
                >
                  {workspaces.map((workspace) => (
                    <option key={workspace.id} value={workspace.id}>
                      {workspace.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="rounded-lg border border-black/10 bg-white p-4 shadow-sm">
                <p className="text-xs font-semibold uppercase text-black/45">
                  Status
                </p>
                <div className="mt-3 grid gap-2">
                  {["all", "pending", "approved", "rejected", "blocked", "allowed"].map((status) => (
                    <button
                      className={`rounded-md px-3 py-2 text-left text-sm transition ${
                        statusFilter === status
                          ? "bg-[#15110d] text-white"
                          : "bg-[#f4f5f0] text-black/70 hover:bg-[#e7ebe2]"
                      }`}
                      key={status}
                      onClick={() => setStatusFilter(status)}
                      type="button"
                    >
                      {status === "all" ? "All" : statusLabels[status] ?? status}
                    </button>
                  ))}
                </div>
              </div>
            </aside>

            <div className="rounded-lg border border-black/10 bg-white shadow-sm">
              <div className="grid grid-cols-[1.5fr_120px_120px_160px] gap-4 border-b border-black/10 px-4 py-3 text-xs font-semibold uppercase text-black/45 max-md:hidden">
                <span>Request</span>
                <span>Status</span>
                <span>Source</span>
                <span>Created</span>
              </div>

              {filteredRequests.length === 0 ? (
                <div className="p-6 text-sm text-black/60">
                  No approval requests match this view.
                </div>
              ) : (
                <div className="divide-y divide-black/10">
                  {filteredRequests.map((request) => (
                    <Link
                      className="grid gap-3 px-4 py-4 transition hover:bg-[#f9faf7] md:grid-cols-[1.5fr_120px_120px_160px] md:items-center"
                      href={`/inbox/${request.id}?workspace_id=${request.workspace_id}`}
                      key={request.id}
                    >
                      <div className="min-w-0">
                        <p className="truncate text-sm font-semibold text-[#15110d]">
                          {request.title}
                        </p>
                        <p className="mt-1 text-xs text-black/50">
                          {request.action_type}
                        </p>
                        {request.context?.reason ? (
                          <p className="mt-2 line-clamp-2 text-sm text-black/65">
                            {request.context.reason}
                          </p>
                        ) : null}
                      </div>

                      <div>
                        <span className="inline-flex rounded-md border border-black/10 bg-[#f4f5f0] px-2 py-1 text-xs font-medium text-black/70">
                          {statusLabels[request.status] ?? request.status}
                        </span>
                      </div>

                      <p className="text-sm text-black/65">{request.source_platform}</p>

                      <time className="text-sm text-black/55">
                        {new Intl.DateTimeFormat(undefined, {
                          dateStyle: "medium",
                          timeStyle: "short",
                        }).format(new Date(request.created_at))}
                      </time>
                    </Link>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </section>
    </main>
  );
}
