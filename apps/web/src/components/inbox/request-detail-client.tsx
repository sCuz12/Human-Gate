"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import type { Session } from "@supabase/supabase-js";

import {
  decideApprovalRequest,
  getApprovalRequestAuditEvents,
  getApprovalRequestDelivery,
  getApprovalRequest,
  type ApprovalRequestAuditEvent,
  type ApprovalRequest,
  type DecisionDelivery,
} from "@/lib/api/approval-requests";
import { getSupabaseBrowserClient } from "@/lib/supabase/client";

const statusLabels: Record<string, string> = {
  pending: "Pending",
  approved: "Approved",
  rejected: "Rejected",
  blocked: "Blocked",
  allowed: "Allowed",
  expired: "Expired",
  cancelled: "Cancelled",
};

const deliveryStatusLabels: Record<string, string> = {
  not_required: "Not required",
  pending: "Pending",
  attempting: "Attempting",
  delivered: "Delivered",
  acknowledged: "Acknowledged",
  retrying: "Retrying",
  permanently_failed: "Permanently failed",
};

export function RequestDetailClient() {
  const params = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const router = useRouter();
  const workspaceID = searchParams.get("workspace_id") ?? "";
  const [session, setSession] = useState<Session | null>(null);
  const [request, setRequest] = useState<ApprovalRequest | null>(null);
  const [delivery, setDelivery] = useState<DecisionDelivery | null>(null);
  const [auditEvents, setAuditEvents] = useState<ApprovalRequestAuditEvent[]>([]);
  const [comment, setComment] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState<"approve" | "reject" | null>(null);

  useEffect(() => {
    const supabase = getSupabaseBrowserClient();

    async function load() {
      const {
        data: { session: currentSession },
      } = await supabase.auth.getSession();

      if (!currentSession) {
        router.replace("/login");
        return;
      }

      setSession(currentSession);
      try {
        const response = await getApprovalRequest(currentSession, {
          workspaceID,
          requestID: params.id,
        });
        setRequest(response.request);
        await loadAuditEvents(currentSession, response.request.workspace_id, response.request.id);
        if (response.request.status !== "pending") {
          await loadDelivery(currentSession, response.request.workspace_id, response.request.id);
        }
        setError(null);
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : "Request could not be loaded.");
      } finally {
        setIsLoading(false);
      }
    }

    void load();
  }, [params.id, router, workspaceID]);

  async function loadDelivery(currentSession: Session, currentWorkspaceID: string, requestID: string) {
    try {
      const response = await getApprovalRequestDelivery(currentSession, {
        workspaceID: currentWorkspaceID,
        requestID,
      });
      setDelivery(response.delivery);
    } catch (deliveryError) {
      if (deliveryError instanceof Error && deliveryError.message.includes("not found")) {
        setDelivery(null);
        return;
      }
      throw deliveryError;
    }
  }

  async function loadAuditEvents(currentSession: Session, currentWorkspaceID: string, requestID: string) {
    const response = await getApprovalRequestAuditEvents(currentSession, {
      workspaceID: currentWorkspaceID,
      requestID,
    });
    setAuditEvents(response.audit_events);
  }

  async function decide(decision: "approve" | "reject") {
    if (!session || !request) {
      return;
    }

    setIsSubmitting(decision);
    setError(null);

    try {
      const response = await decideApprovalRequest(session, {
        workspaceID: request.workspace_id,
        requestID: request.id,
        decision,
        comment,
      });
      setRequest(response.request);
      setComment("");
      await Promise.all([
        loadDelivery(session, response.request.workspace_id, response.request.id),
        loadAuditEvents(session, response.request.workspace_id, response.request.id),
      ]);
    } catch (decisionError) {
      const message = decisionError instanceof Error ? decisionError.message : "Decision could not be saved.";
      setError(message.includes("timed out") ? "Decision request timed out. Check the API logs." : message);
    } finally {
      setIsSubmitting(null);
    }
  }

  if (isLoading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f4f5f0] px-6">
        <div className="rounded-lg border border-black/10 bg-white px-5 py-4 text-sm text-black/65 shadow-sm">
          Loading request...
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f4f5f0] text-[#15110d]">
      <header className="border-b border-black/10 bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-5">
          <div>
            <Link className="text-sm font-medium text-[#1f6f78]" href="/inbox">
              Back to inbox
            </Link>
            <h1 className="mt-2 text-2xl font-semibold">
              {request?.title ?? "Approval request"}
            </h1>
          </div>
          {request ? (
            <span className="rounded-md border border-black/10 bg-[#f4f5f0] px-3 py-2 text-sm font-medium text-black/70">
              {statusLabels[request.status] ?? request.status}
            </span>
          ) : null}
        </div>
      </header>

      <section className="mx-auto grid max-w-6xl gap-6 px-6 py-6 lg:grid-cols-[1fr_320px]">
        {error ? (
          <div className="rounded-lg border border-[#b74b2a]/20 bg-[#fff2ec] p-4 text-sm text-[#8d3419] lg:col-span-2">
            {error}
          </div>
        ) : null}

        {!request ? (
          <div className="rounded-lg border border-black/10 bg-white p-6 text-sm text-black/65">
            Request not found.
          </div>
        ) : (
          <>
            <div className="space-y-6">
              <section className="rounded-lg border border-black/10 bg-white p-5 shadow-sm">
                <p className="text-xs font-semibold uppercase text-black/45">Action</p>
                <dl className="mt-4 grid gap-4 sm:grid-cols-2">
                  <div>
                    <dt className="text-sm text-black/45">Type</dt>
                    <dd className="mt-1 text-sm font-medium">{request.action_type}</dd>
                  </div>
                  <div>
                    <dt className="text-sm text-black/45">Source</dt>
                    <dd className="mt-1 text-sm font-medium">{request.source_platform}</dd>
                  </div>
                  <div>
                    <dt className="text-sm text-black/45">Workflow</dt>
                    <dd className="mt-1 break-all text-sm font-medium">{request.source_workflow_id}</dd>
                  </div>
                  <div>
                    <dt className="text-sm text-black/45">Execution</dt>
                    <dd className="mt-1 break-all text-sm font-medium">{request.source_execution_id}</dd>
                  </div>
                </dl>
              </section>

              <section className="rounded-lg border border-black/10 bg-white p-5 shadow-sm">
                <p className="text-xs font-semibold uppercase text-black/45">Context</p>
                <p className="mt-4 text-sm leading-6 text-black/75">
                  {request.context?.reason ?? "No reason provided."}
                </p>
                <p className="mt-3 text-sm text-black/55">
                  Reversible: {request.context?.reversible === undefined ? "Unknown" : request.context.reversible ? "Yes" : "No"}
                </p>
              </section>

              <section className="rounded-lg border border-black/10 bg-white p-5 shadow-sm">
                <p className="text-xs font-semibold uppercase text-black/45">Payload</p>
                <pre className="mt-4 max-h-[420px] overflow-auto rounded-md bg-[#15110d] p-4 text-xs leading-6 text-white">
                  {JSON.stringify(request.original_action, null, 2)}
                </pre>
              </section>

              <section className="rounded-lg border border-black/10 bg-white p-5 shadow-sm">
                <div className="flex items-center justify-between gap-4">
                  <p className="text-xs font-semibold uppercase text-black/45">Audit timeline</p>
                  <span className="text-xs text-black/45">{auditEvents.length} events</span>
                </div>
                {auditEvents.length > 0 ? (
                  <ol className="mt-5 space-y-5">
                    {auditEvents.map((event) => (
                      <li className="relative border-l border-black/10 pl-5" key={event.id}>
                        <span className="absolute -left-[5px] top-1.5 h-2.5 w-2.5 rounded-full bg-[#1f6f78]" />
                        <div className="flex flex-wrap items-start justify-between gap-2">
                          <div>
                            <p className="text-sm font-semibold">{formatEventType(event.event_type)}</p>
                            <p className="mt-1 text-xs text-black/50">
                              {event.actor_type}
                              {event.actor_id ? ` · ${event.actor_id}` : ""}
                            </p>
                          </div>
                          <time className="text-xs text-black/45" dateTime={event.created_at}>
                            {formatDate(event.created_at)}
                          </time>
                        </div>
                        {Object.keys(event.metadata ?? {}).length > 0 ? (
                          <dl className="mt-3 grid gap-2 rounded-md bg-[#f4f5f0] p-3 text-xs">
                            {Object.entries(event.metadata).map(([key, value]) => (
                              <div className="grid gap-1 sm:grid-cols-[150px_1fr]" key={key}>
                                <dt className="font-medium text-black/45">{formatEventType(key)}</dt>
                                <dd className="break-words text-black/70">{formatMetadataValue(value)}</dd>
                              </div>
                            ))}
                          </dl>
                        ) : null}
                      </li>
                    ))}
                  </ol>
                ) : (
                  <p className="mt-4 text-sm text-black/60">
                    No audit events have been recorded for this request yet.
                  </p>
                )}
              </section>
            </div>

            <aside className="space-y-4">
              <section className="rounded-lg border border-black/10 bg-white p-5 shadow-sm">
                <p className="text-xs font-semibold uppercase text-black/45">Matched policy</p>
                {request.matched_policy ? (
                  <dl className="mt-4 space-y-3 text-sm">
                    <div>
                      <dt className="text-black/45">Name</dt>
                      <dd className="mt-1 font-medium">{request.matched_policy.name}</dd>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <dt className="text-black/45">Effect</dt>
                      <dd className="rounded-md border border-black/10 bg-[#f4f5f0] px-2 py-1 font-medium">
                        {formatEventType(request.matched_policy.effect)}
                      </dd>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <dt className="text-black/45">Priority</dt>
                      <dd className="font-medium">{request.matched_policy.priority}</dd>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <dt className="text-black/45">Version</dt>
                      <dd className="font-medium">v{request.matched_policy.version_number}</dd>
                    </div>
                    {request.matched_policy.deadline_seconds > 0 ? (
                      <div className="flex items-center justify-between gap-4">
                        <dt className="text-black/45">Deadline</dt>
                        <dd className="font-medium">{formatDuration(request.matched_policy.deadline_seconds)}</dd>
                      </div>
                    ) : null}
                  </dl>
                ) : (
                  <p className="mt-4 text-sm text-black/60">
                    No policy matched. Workspace default behavior was used.
                  </p>
                )}
              </section>

              <section className="rounded-lg border border-black/10 bg-white p-5 shadow-sm">
                <p className="text-xs font-semibold uppercase text-black/45">Decision</p>
                <textarea
                  className="mt-4 min-h-28 w-full rounded-md border border-black/10 bg-white px-3 py-2 text-sm outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
                  disabled={request.status !== "pending"}
                  onChange={(event) => setComment(event.target.value)}
                  placeholder="Optional comment"
                  value={comment}
                />
                <div className="mt-4 grid gap-3">
                  <button
                    className="rounded-md bg-[#176a44] px-4 py-3 text-sm font-semibold text-white transition hover:bg-[#0f5636] disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={request.status !== "pending" || isSubmitting !== null}
                    onClick={() => void decide("approve")}
                    type="button"
                  >
                    {isSubmitting === "approve" ? "Approving..." : "Approve"}
                  </button>
                  <button
                    className="rounded-md border border-[#b74b2a]/30 bg-white px-4 py-3 text-sm font-semibold text-[#8d3419] transition hover:bg-[#fff2ec] disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={request.status !== "pending" || isSubmitting !== null}
                    onClick={() => void decide("reject")}
                    type="button"
                  >
                    {isSubmitting === "reject" ? "Rejecting..." : "Reject"}
                  </button>
                </div>
              </section>

              <section className="rounded-lg border border-black/10 bg-white p-5 shadow-sm">
                <p className="text-xs font-semibold uppercase text-black/45">Delivery</p>
                {request.status === "pending" ? (
                  <p className="mt-4 text-sm text-black/60">
                    Delivery will be scheduled after a decision is recorded.
                  </p>
                ) : delivery ? (
                  <dl className="mt-4 space-y-3 text-sm">
                    <div className="flex items-center justify-between gap-4">
                      <dt className="text-black/45">Status</dt>
                      <dd className="rounded-md border border-black/10 bg-[#f4f5f0] px-2 py-1 font-medium">
                        {deliveryStatusLabels[delivery.status] ?? delivery.status}
                      </dd>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <dt className="text-black/45">Attempts</dt>
                      <dd className="font-medium">{delivery.attempt_count}</dd>
                    </div>
                    {delivery.last_response_code ? (
                      <div className="flex items-center justify-between gap-4">
                        <dt className="text-black/45">Last response</dt>
                        <dd className="font-medium">{delivery.last_response_code}</dd>
                      </div>
                    ) : null}
                    {delivery.next_attempt_at ? (
                      <div>
                        <dt className="text-black/45">Next attempt</dt>
                        <dd className="mt-1 break-words font-medium">{formatDate(delivery.next_attempt_at)}</dd>
                      </div>
                    ) : null}
                    {delivery.delivered_at ? (
                      <div>
                        <dt className="text-black/45">Delivered</dt>
                        <dd className="mt-1 break-words font-medium">{formatDate(delivery.delivered_at)}</dd>
                      </div>
                    ) : null}
                    {delivery.last_error ? (
                      <div>
                        <dt className="text-black/45">Last error</dt>
                        <dd className="mt-1 break-words rounded-md bg-[#fff2ec] p-2 text-[#8d3419]">
                          {delivery.last_error}
                        </dd>
                      </div>
                    ) : null}
                  </dl>
                ) : (
                  <p className="mt-4 text-sm text-black/60">
                    No delivery has been scheduled for this request yet.
                  </p>
                )}
              </section>
            </aside>
          </>
        )}
      </section>
    </main>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatEventType(value: string) {
  return value
    .split(/[._-]/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatDuration(seconds: number) {
  if (seconds < 60) {
    return `${seconds}s`;
  }

  const minutes = Math.round(seconds / 60);
  if (minutes < 60) {
    return `${minutes}m`;
  }

  const hours = Math.round(minutes / 60);
  return `${hours}h`;
}

function formatMetadataValue(value: unknown) {
  if (value === null || value === undefined) {
    return "None";
  }

  if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }

  return JSON.stringify(value);
}
