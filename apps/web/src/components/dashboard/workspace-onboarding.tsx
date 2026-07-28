"use client";

import { useState } from "react";
import type { Session } from "@supabase/supabase-js";

import { createAPIKey, type APIKey } from "@/lib/api/api-keys";
import { createWorkspace } from "@/lib/api/workspaces";

type WorkspaceOnboardingProps = {
  session: Session;
  onCreated: (generatedToken?: APIKey) => Promise<void> | void;
};

export function WorkspaceOnboarding({
  session,
  onCreated,
}: WorkspaceOnboardingProps) {
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const workspaceResponse = await createWorkspace(session, name);
      const tokenResponse = await createAPIKey(session, {
        workspaceID: workspaceResponse.workspace.id,
        name: `${workspaceResponse.workspace.name} workflow key`,
        scopes: ["approval_requests:create"],
      });
      setName("");
      await onCreated(tokenResponse.api_key);
    } catch (submissionError) {
      setError(
        submissionError instanceof Error
          ? submissionError.message
          : "Workspace could not be created.",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="rounded-[2rem] border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
      <p className="text-sm font-semibold uppercase tracking-[0.18em] text-[#b74b2a]">
        Create your first workspace
      </p>
      <h2 className="mt-3 text-2xl font-semibold text-[#15110d]">
        Start with a single team workspace.
      </h2>
      <p className="mt-2 text-sm leading-6 text-black/65">
        The creator becomes the initial owner. After that you can create API
        keys, define policies, and invite approvers.
      </p>

      <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
        <label className="block space-y-2">
          <span className="text-sm font-medium text-[#15110d]">
            Workspace name
          </span>
          <input
            required
            className="w-full rounded-2xl border border-black/10 bg-[#fbf8f2] px-4 py-3 outline-none transition focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
            onChange={(event) => setName(event.target.value)}
            placeholder="Support Operations"
            type="text"
            value={name}
          />
        </label>

        {error ? (
          <p className="rounded-2xl border border-[#b74b2a]/20 bg-[#fff2ec] px-4 py-3 text-sm text-[#8d3419]">
            {error}
          </p>
        ) : null}

        <button
          className="rounded-2xl bg-[#15110d] px-4 py-3 text-sm font-semibold text-white transition hover:bg-black disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isSubmitting}
          type="submit"
        >
          {isSubmitting ? "Creating workspace..." : "Create workspace"}
        </button>
      </form>
    </div>
  );
}
