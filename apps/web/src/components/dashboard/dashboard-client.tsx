"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { Session } from "@supabase/supabase-js";

import { APIKeyGenerator } from "@/components/dashboard/api-key-generator";
import { PolicyManager } from "@/components/dashboard/policy-manager";
import { WorkspaceOnboarding } from "@/components/dashboard/workspace-onboarding";
import type { APIKey } from "@/lib/api/api-keys";
import { listWorkspaces, type Workspace } from "@/lib/api/workspaces";
import { getSupabaseBrowserClient } from "@/lib/supabase/client";

export function DashboardClient() {
  const router = useRouter();
  const [session, setSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [workspaceError, setWorkspaceError] = useState<string | null>(null);
  const [generatedToken, setGeneratedToken] = useState<APIKey | null>(null);

  useEffect(() => {
    const supabase = getSupabaseBrowserClient();

    async function loadSession() {
      const {
        data: { session: currentSession },
      } = await supabase.auth.getSession();

      if (!currentSession) {
        router.replace("/login");
        return;
      }

      setSession(currentSession);
      await loadWorkspaces(currentSession);
      setIsLoading(false);
    }

    void loadSession();

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, updatedSession) => {
      if (!updatedSession) {
        router.replace("/login");
        return;
      }

      setSession(updatedSession);
      void loadWorkspaces(updatedSession).finally(() => setIsLoading(false));
    });

    return () => {
      subscription.unsubscribe();
    };
  }, [router]);

  async function signOut() {
    const supabase = getSupabaseBrowserClient();
    await supabase.auth.signOut();
    router.replace("/login");
    router.refresh();
  }

  async function loadWorkspaces(activeSession: Session) {
    try {
      const response = await listWorkspaces(activeSession);
      setWorkspaces(response.workspaces);
      setWorkspaceError(null);
    } catch (error) {
      setWorkspaceError(
        error instanceof Error ? error.message : "Workspaces could not be loaded.",
      );
    }
  }

  if (isLoading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f7f1e7] px-6">
        <div className="rounded-[1.75rem] border border-black/10 bg-white px-6 py-5 text-sm text-black/65 shadow-[0_18px_50px_rgba(0,0,0,0.08)]">
          Loading dashboard...
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[linear-gradient(180deg,#f8f3eb_0%,#efe7dc_100%)] px-6 py-8 md:px-10">
      <div className="mx-auto max-w-6xl">
        <header className="flex flex-col gap-4 rounded-[2rem] border border-black/10 bg-[#16302b] px-6 py-6 text-[#f5ecde] shadow-[0_24px_60px_rgba(0,0,0,0.12)] md:flex-row md:items-end md:justify-between">
          <div className="space-y-3">
            <p className="text-sm font-semibold uppercase tracking-[0.24em] text-[#9bc5bb]">
              HumanGate Dashboard
            </p>
            <div>
              <h1 className="text-3xl font-semibold md:text-4xl">
                Signed in and ready to build the approval workflow.
              </h1>
              <p className="mt-2 text-sm leading-6 text-[#dbcdbb]">
                Continue by creating a workspace endpoint and wiring API key
                creation from the UI.
              </p>
            </div>
          </div>

          <div className="rounded-[1.5rem] border border-white/10 bg-white/5 px-5 py-4 text-sm">
            <p className="text-[#9bc5bb]">Authenticated as</p>
            <p className="mt-1 font-medium text-white">
              {session?.user.email ?? session?.user.id}
            </p>
          </div>
        </header>

        <section className="mt-8 grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
          <div className="space-y-6">
            {workspaceError ? (
              <div className="rounded-[2rem] border border-[#b74b2a]/20 bg-[#fff2ec] p-6 text-sm text-[#8d3419] shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
                {workspaceError}
              </div>
            ) : null}

            {session && workspaces.length === 0 ? (
              <WorkspaceOnboarding
                onCreated={async (token) => {
                  setGeneratedToken(token ?? null);
                  await loadWorkspaces(session);
                }}
                session={session}
              />
            ) : (
              <div className="rounded-lg border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
                <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#b74b2a]">
                  Your workspaces
                </p>
                <div className="mt-5 grid gap-4">
                  {workspaces.map((workspace) => (
                    <div
                      key={workspace.id}
                      className="rounded-lg border border-black/10 bg-[#fbf8f2] p-4"
                    >
                      <div className="flex items-start justify-between gap-4">
                        <div>
                          <p className="text-lg font-semibold text-[#15110d]">
                            {workspace.name}
                          </p>
                          <p className="mt-1 text-sm text-black/55">
                            slug: {workspace.slug}
                          </p>
                        </div>
                        <div className="rounded-full border border-[#1f6f78]/20 bg-[#eef8f9] px-3 py-1 text-xs font-semibold uppercase tracking-[0.16em] text-[#13505a]">
                          {workspace.role}
                        </div>
                      </div>
                      <p className="mt-3 text-sm text-black/65">
                        Default policy behavior:{" "}
                        <span className="font-medium">
                          {workspace.default_policy_effect}
                        </span>
                      </p>
                      {session ? (
                        <div className="mt-4 space-y-4">
                          <APIKeyGenerator
                            initialToken={
                              generatedToken?.workspace_id === workspace.id
                                ? generatedToken
                                : null
                            }
                            session={session}
                            workspace={workspace}
                          />
                          <PolicyManager session={session} workspace={workspace} />
                        </div>
                      ) : null}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="rounded-[2rem] border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#b74b2a]">
                Next implementation steps
              </p>
              <div className="mt-5 grid gap-4">
                {[
                  "Create API keys directly from the app",
                  "Expose approval requests in the inbox",
                  "Add policy creation and routing",
                  "Connect the first workflow integration",
                ].map((item) => (
                  <div
                    key={item}
                    className="rounded-[1.5rem] border border-black/10 bg-[#fbf8f2] px-4 py-4 text-sm text-black/70"
                  >
                    {item}
                  </div>
                ))}
              </div>
            </div>
          </div>

          <aside className="space-y-6">
            <div className="rounded-[2rem] border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#1f6f78]">
                Session details
              </p>
              <dl className="mt-4 space-y-3 text-sm">
                <div>
                  <dt className="text-black/45">User ID</dt>
                  <dd className="mt-1 break-all text-black/80">
                    {session?.user.id}
                  </dd>
                </div>
                <div>
                  <dt className="text-black/45">Email</dt>
                  <dd className="mt-1 text-black/80">
                    {session?.user.email ?? "Unavailable"}
                  </dd>
                </div>
              </dl>
            </div>

            <div className="rounded-[2rem] border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#1f6f78]">
                Quick actions
              </p>
              <div className="mt-4 flex flex-col gap-3">
                <Link
                  className="rounded-2xl border border-black/10 px-4 py-3 text-sm font-medium text-black/75 transition hover:border-[#1f6f78] hover:text-[#1f6f78]"
                  href="/inbox"
                >
                  Open inbox
                </Link>
                <button
                  className="rounded-2xl bg-[#15110d] px-4 py-3 text-sm font-semibold text-white transition hover:bg-black"
                  onClick={signOut}
                  type="button"
                >
                  Sign out
                </button>
              </div>
            </div>
          </aside>
        </section>
      </div>
    </main>
  );
}
