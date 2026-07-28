"use client";

import type { Session } from "@supabase/supabase-js";

import { apiFetch } from "@/lib/api/client";

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  default_policy_effect: string;
  role: string;
};

type ListWorkspacesResponse = {
  workspaces: Workspace[];
};

type CreateWorkspaceResponse = {
  workspace: Workspace;
};

export async function listWorkspaces(session: Session) {
  return apiFetch<ListWorkspacesResponse>("/api/v1/workspaces", session);
}

export async function createWorkspace(session: Session, name: string) {
  return apiFetch<CreateWorkspaceResponse>("/api/v1/workspaces", session, {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}
