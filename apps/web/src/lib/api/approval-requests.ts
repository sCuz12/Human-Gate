"use client";

import type { Session } from "@supabase/supabase-js";

import { apiFetch } from "@/lib/api/client";

export type ApprovalRequest = {
  id: string;
  workspace_id: string;
  action_type: string;
  title: string;
  description?: string;
  status: string;
  decision_required: boolean;
  source_platform: string;
  source_workflow_id: string;
  source_execution_id: string;
  original_action: {
    parameters?: Record<string, unknown>;
  };
  context: {
    reason?: string;
    reversible?: boolean;
  };
  created_at: string;
  resolved_at?: string;
  expires_at?: string;
};

export type DecisionDelivery = {
  id: string;
  decision_id: string;
  status: string;
  attempt_count: number;
  next_attempt_at?: string;
  last_attempt_at?: string;
  last_response_code?: number;
  last_error?: string;
  delivered_at?: string;
  acknowledged_at?: string;
  updated_at: string;
};

type ListApprovalRequestsResponse = {
  requests: ApprovalRequest[];
};

export async function listApprovalRequests(
  session: Session,
  input: {
    workspaceID: string;
    limit?: number;
    offset?: number;
  },
) {
  const params = new URLSearchParams({
    workspace_id: input.workspaceID,
    limit: String(input.limit ?? 50),
    offset: String(input.offset ?? 0),
  });

  return apiFetch<ListApprovalRequestsResponse>(
    `/api/v1/approval-requests?${params.toString()}`,
    session,
  );
}

type GetApprovalRequestResponse = {
  request: ApprovalRequest;
};

export async function getApprovalRequest(
  session: Session,
  input: {
    workspaceID: string;
    requestID: string;
  },
) {
  const params = new URLSearchParams({
    workspace_id: input.workspaceID,
  });

  return apiFetch<GetApprovalRequestResponse>(
    `/api/v1/approval-requests/${input.requestID}?${params.toString()}`,
    session,
  );
}

type GetApprovalRequestDeliveryResponse = {
  delivery: DecisionDelivery;
};

export async function getApprovalRequestDelivery(
  session: Session,
  input: {
    workspaceID: string;
    requestID: string;
  },
) {
  const params = new URLSearchParams({
    workspace_id: input.workspaceID,
  });

  return apiFetch<GetApprovalRequestDeliveryResponse>(
    `/api/v1/approval-requests/${input.requestID}/delivery?${params.toString()}`,
    session,
  );
}

export async function decideApprovalRequest(
  session: Session,
  input: {
    workspaceID: string;
    requestID: string;
    decision: "approve" | "reject";
    comment?: string;
  },
) {
  return apiFetch<GetApprovalRequestResponse>(
    `/api/v1/approval-requests/${input.requestID}/${input.decision}`,
    session,
    {
      method: "POST",
      timeoutMs: 15_000,
      body: JSON.stringify({
        workspace_id: input.workspaceID,
        comment: input.comment ?? "",
      }),
    },
  );
}
