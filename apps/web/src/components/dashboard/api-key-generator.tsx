"use client";

import { useState } from "react";
import type { Session } from "@supabase/supabase-js";

import { createAPIKey, type APIKey } from "@/lib/api/api-keys";
import type { Workspace } from "@/lib/api/workspaces";

type APIKeyGeneratorProps = {
  session: Session;
  workspace: Workspace;
  initialToken?: APIKey | null;
};

export function APIKeyGenerator({
  session,
  workspace,
  initialToken = null,
}: APIKeyGeneratorProps) {
  const [token, setToken] = useState<APIKey | null>(initialToken);
  const [name, setName] = useState(`${workspace.name} workflow key`);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleGenerate(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setNotice(null);
    setIsSubmitting(true);

    try {
      const response = await createAPIKey(session, {
        workspaceID: workspace.id,
        name,
        scopes: ["approval_requests:create"],
      });
      setToken(response.api_key);
    } catch (submissionError) {
      setError(
        submissionError instanceof Error
          ? submissionError.message
          : "API key could not be created.",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  async function copyToken() {
    if (!token?.secret) {
      return;
    }

    await navigator.clipboard.writeText(token.secret);
    setNotice("Token copied.");
  }

  return (
    <div className="rounded-lg border border-black/10 bg-white p-5 shadow-[0_14px_35px_rgba(0,0,0,0.06)]">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-[#1f6f78]">
            Workflow token
          </p>
          <h3 className="mt-2 text-lg font-semibold text-[#15110d]">
            Generate an API key
          </h3>
          <p className="mt-1 text-sm leading-6 text-black/60">
            Use this token from n8n, LangGraph, or another workflow client.
          </p>
        </div>
      </div>

      <form className="mt-5 flex flex-col gap-3 sm:flex-row" onSubmit={handleGenerate}>
        <input
          required
          className="min-w-0 flex-1 rounded-md border border-black/10 bg-[#fbf8f2] px-3 py-2 text-sm outline-none transition focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
          onChange={(event) => setName(event.target.value)}
          type="text"
          value={name}
        />
        <button
          className="rounded-md bg-[#15110d] px-4 py-2 text-sm font-semibold text-white transition hover:bg-black disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isSubmitting}
          type="submit"
        >
          {isSubmitting ? "Generating..." : "Generate token"}
        </button>
      </form>

      {error ? (
        <p className="mt-4 rounded-md border border-[#b74b2a]/20 bg-[#fff2ec] px-3 py-2 text-sm text-[#8d3419]">
          {error}
        </p>
      ) : null}

      {token ? (
        <div className="mt-4 rounded-md border border-[#1f6f78]/20 bg-[#eef8f9] p-3">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase text-[#13505a]">
                Save this token now
              </p>
              <code className="mt-2 block break-all rounded-md bg-white px-3 py-2 text-xs text-[#15110d]">
                {token.secret}
              </code>
            </div>
            <button
              className="rounded-md border border-[#1f6f78]/30 px-3 py-2 text-sm font-semibold text-[#13505a] transition hover:bg-white"
              onClick={copyToken}
              type="button"
            >
              Copy
            </button>
          </div>
          {notice ? <p className="mt-2 text-sm text-[#13505a]">{notice}</p> : null}
        </div>
      ) : null}
    </div>
  );
}
