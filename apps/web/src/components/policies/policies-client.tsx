"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { Session } from "@supabase/supabase-js";

import { PolicyManager } from "@/components/dashboard/policy-manager";
import { listWorkspaces, type Workspace } from "@/lib/api/workspaces";
import { getSupabaseBrowserClient } from "@/lib/supabase/client";

export function PoliciesClient() {
  const router = useRouter();
  const [session, setSession] = useState<Session | null>(null);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [selectedWorkspaceID, setSelectedWorkspaceID] = useState("");
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

      try {
        const response = await listWorkspaces(currentSession);
        setWorkspaces(response.workspaces);
        setSelectedWorkspaceID(response.workspaces[0]?.id ?? "");
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : "Workspaces could not be loaded.");
      } finally {
        setIsLoading(false);
      }
    }

    void loadInitialState();
  }, [router]);

  const selectedWorkspace = workspaces.find((workspace) => workspace.id === selectedWorkspaceID);

  if (isLoading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f4f5f0] px-6">
        <div className="rounded-lg border border-black/10 bg-white px-5 py-4 text-sm text-black/65 shadow-sm">
          Loading policies...
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f4f5f0] text-[#15110d]">
      <header className="border-b border-black/10 bg-white">
        <div className="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-5 md:flex-row md:items-center md:justify-between">
          <div>
            <Link className="text-sm font-medium text-[#1f6f78]" href="/dashboard">
              Back to dashboard
            </Link>
            <h1 className="mt-2 text-2xl font-semibold">Policies</h1>
            <p className="mt-1 text-sm text-black/55">
              Decide which workflow actions are allowed, reviewed, blocked, or expired.
            </p>
          </div>
          <Link
            className="rounded-md bg-[#15110d] px-4 py-3 text-sm font-semibold text-white transition hover:bg-black"
            href="/inbox"
          >
            Open inbox
          </Link>
        </div>
      </header>

      <section className="mx-auto max-w-6xl px-6 py-6">
        {error ? (
          <div className="mb-5 rounded-lg border border-[#b74b2a]/20 bg-[#fff2ec] p-4 text-sm text-[#8d3419]">
            {error}
          </div>
        ) : null}

        {workspaces.length === 0 ? (
          <div className="rounded-lg border border-black/10 bg-white p-6 text-sm text-black/65 shadow-sm">
            Create a workspace before managing policies.
          </div>
        ) : (
          <div className="grid gap-5 lg:grid-cols-[280px_1fr]">
            <aside className="rounded-lg border border-black/10 bg-white p-4 shadow-sm">
              <label className="block text-xs font-semibold uppercase text-black/45">
                Workspace
              </label>
              <select
                className="mt-2 w-full rounded-md border border-black/10 bg-white px-3 py-2 text-sm outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
                onChange={(event) => setSelectedWorkspaceID(event.target.value)}
                value={selectedWorkspaceID}
              >
                {workspaces.map((workspace) => (
                  <option key={workspace.id} value={workspace.id}>
                    {workspace.name}
                  </option>
                ))}
              </select>

              <div className="mt-5 rounded-md bg-[#f4f5f0] p-3 text-sm text-black/60">
                Lower priority numbers run first. The first matching active policy wins.
              </div>
            </aside>

            {session && selectedWorkspace ? (
              <PolicyManager key={selectedWorkspace.id} session={session} workspace={selectedWorkspace} />
            ) : null}
          </div>
        )}
      </section>
    </main>
  );
}
