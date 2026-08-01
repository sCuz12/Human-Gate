"use client";

import type { Session } from "@supabase/supabase-js";

import { apiFetch } from "@/lib/api/client";

export type PolicyCondition = {
  field: "action.type" | "source.platform" | "context.reversible";
  operator: "equals";
  value: string | boolean;
};

export type Policy = {
  id: string;
  workspace_id: string;
  name: string;
  description?: string;
  priority: number;
  is_active: boolean;
  version_id: string;
  version_number: number;
  conditions: PolicyCondition[];
  effect: "allow" | "require_approval" | "block";
  approval_settings: {
    deadline_seconds?: number;
  };
  deadline_seconds?: number;
  created_at: string;
};

type ListPoliciesResponse = {
  policies: Policy[];
};

type CreatePolicyResponse = {
  policy: Policy;
};

export async function listPolicies(session: Session, workspaceID: string) {
  return apiFetch<ListPoliciesResponse>(
    `/api/v1/policies?workspace_id=${encodeURIComponent(workspaceID)}`,
    session,
  );
}

export async function createPolicy(
  session: Session,
  input: {
    workspaceID: string;
    name: string;
    description?: string;
    priority: number;
    isActive: boolean;
    conditions: PolicyCondition[];
    effect: Policy["effect"];
    deadlineSeconds: number;
  },
) {
  return apiFetch<CreatePolicyResponse>("/api/v1/policies", session, {
    method: "POST",
    body: JSON.stringify({
      workspace_id: input.workspaceID,
      name: input.name,
      description: input.description ?? "",
      priority: input.priority,
      is_active: input.isActive,
      conditions: input.conditions,
      effect: input.effect,
      deadline_seconds: input.deadlineSeconds,
    }),
  });
}

export async function updatePolicy(
  session: Session,
  input: {
    workspaceID: string;
    policyID: string;
    name: string;
    description?: string;
    priority: number;
    isActive: boolean;
    conditions: PolicyCondition[];
    effect: Policy["effect"];
    deadlineSeconds: number;
  },
) {
  return apiFetch<CreatePolicyResponse>(`/api/v1/policies/${input.policyID}`, session, {
    method: "PATCH",
    body: JSON.stringify({
      workspace_id: input.workspaceID,
      name: input.name,
      description: input.description ?? "",
      priority: input.priority,
      is_active: input.isActive,
      conditions: input.conditions,
      effect: input.effect,
      deadline_seconds: input.deadlineSeconds,
    }),
  });
}

export async function deletePolicy(
  session: Session,
  input: {
    workspaceID: string;
    policyID: string;
  },
) {
  await apiFetch<void>(
    `/api/v1/policies/${input.policyID}?workspace_id=${encodeURIComponent(input.workspaceID)}`,
    session,
    {
      method: "DELETE",
    },
  );
}
