"use client";

import type { Session } from "@supabase/supabase-js";

import { apiFetch } from "@/lib/api/client";

export type APIKey = {
  id: string;
  workspace_id: string;
  name: string;
  prefix: string;
  scopes: string[];
  created_at: string;
  secret: string;
};

type CreateAPIKeyResponse = {
  api_key: APIKey;
};

export async function createAPIKey(
  session: Session,
  input: {
    workspaceID: string;
    name: string;
    scopes?: string[];
  },
) {
  return apiFetch<CreateAPIKeyResponse>("/api/v1/api-keys", session, {
    method: "POST",
    body: JSON.stringify({
      workspace_id: input.workspaceID,
      name: input.name,
      scopes: input.scopes ?? ["approval_requests:create"],
    }),
  });
}
