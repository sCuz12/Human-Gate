import Link from "next/link";

import { CopySnippet } from "@/components/integrations/copy-snippet";

const requestUrl = "POST https://api.getdecree.com/api/v1/approval-requests";

const headersSnippet = `Authorization: Bearer hg_live_your_api_key
Idempotency-Key: n8n:{{$workflow.id}}:{{$execution.id}}:{{$json.action_id}}
Content-Type: application/json`;

const bodySnippet = `{
  "action": {
    "type": "customer.refund",
    "title": "Refund customer",
    "parameters": {
      "amount": 82,
      "currency": "GBP"
    }
  },
  "context": {
    "reason": "Customer supplied evidence of damaged goods",
    "reversible": false,
    "risk_level": "high"
  },
  "source": {
    "platform": "n8n",
    "workflow_id": "{{$workflow.id}}",
    "execution_id": "{{$execution.id}}"
  },
  "continuation": {
    "strategy": "webhook",
    "url": "{{$json.resumeUrl}}"
  }
}`;

const pendingResponseSnippet = `{
  "approval_request": {
    "id": "7d2f5e33-9fb4-4c3c-ae54-4d11f4d4d4f1",
    "status": "pending",
    "action_type": "customer.refund",
    "title": "Refund customer",
    "expires_at": "2026-08-05T18:30:00Z"
  },
  "decision": null
}`;

const callbackSnippet = `{
  "delivery_id": "6a5e7d82-28c2-4e85-8a3e-40b90f6aa7a5",
  "decision_id": "7f09c432-8b91-4f59-8d70-cb80ce5f9a6f",
  "approval_request_id": "7d2f5e33-9fb4-4c3c-ae54-4d11f4d4d4f1",
  "workspace_id": "18eb0a84-8c4e-42bb-a2f1-6e8e0ad3ad27",
  "decision": "approved",
  "action_type": "customer.refund",
  "original_action": {
    "type": "customer.refund",
    "title": "Refund customer",
    "parameters": {
      "amount": 82,
      "currency": "GBP"
    }
  },
  "original_action_hash": "sha256:...",
  "approved_action": {
    "type": "customer.refund",
    "title": "Refund customer",
    "parameters": {
      "amount": 82,
      "currency": "GBP"
    }
  },
  "approved_action_hash": "sha256:...",
  "source_platform": "n8n",
  "continuation_strategy": "webhook",
  "issued_at": "2026-08-05T16:12:30Z",
  "signature_algorithm": "HMAC-SHA256",
  "signature_header": "X-Decree-Signature"
}`;

const decisionExamplesSnippet = `[
  {
    "decision": "allowed",
    "case": "A policy allowed the action without human review.",
    "n8n_branch": "Continue to the business action.",
    "approved_action": {
      "type": "customer.refund",
      "title": "Refund customer",
      "parameters": {
        "amount": 32,
        "currency": "GBP"
      }
    },
    "approved_action_hash": "sha256:..."
  },
  {
    "decision": "approved",
    "case": "A human approved the original action without edits.",
    "n8n_branch": "Continue to the business action.",
    "approved_action": {
      "type": "customer.refund",
      "title": "Refund customer",
      "parameters": {
        "amount": 82,
        "currency": "GBP"
      }
    },
    "approved_action_hash": "sha256:..."
  },
  {
    "decision": "approved_with_changes",
    "case": "A human changed one or more action fields before approval.",
    "n8n_branch": "Use approved_action, not the original action.",
    "approved_action": {
      "type": "customer.refund",
      "title": "Refund customer",
      "parameters": {
        "amount": 50,
        "currency": "GBP"
      }
    },
    "approved_action_hash": "sha256:..."
  },
  {
    "decision": "rejected",
    "case": "A human rejected the pending request.",
    "n8n_branch": "Stop the business action and notify the workflow owner.",
    "approved_action": null,
    "approved_action_hash": null
  },
  {
    "decision": "blocked",
    "case": "A policy blocked the action before human review.",
    "n8n_branch": "Stop the business action and mark the execution blocked.",
    "approved_action": null,
    "approved_action_hash": null
  },
  {
    "decision": "expired",
    "case": "No valid approval was recorded before the deadline.",
    "n8n_branch": "Stop the business action or restart the approval request.",
    "approved_action": null,
    "approved_action_hash": null
  },
  {
    "decision": "cancelled",
    "case": "The source workflow or an administrator cancelled the request.",
    "n8n_branch": "Stop waiting and end the workflow cleanly.",
    "approved_action": null,
    "approved_action_hash": null
  }
]`;

const decisionCases = [
  {
    decision: "allowed",
    source: "Automatic policy",
    when: "Policy evaluation says the action can continue without human review.",
    n8n: "Continue to the final business action.",
  },
  {
    decision: "approved",
    source: "Human reviewer",
    when: "A reviewer approved the original action exactly as submitted.",
    n8n: "Continue using the original approved action.",
  },
  {
    decision: "approved_with_changes",
    source: "Human reviewer",
    when: "A reviewer edited fields such as amount, recipient, message, or timing.",
    n8n: "Continue using approved_action instead of original_action.",
  },
  {
    decision: "rejected",
    source: "Human reviewer",
    when: "A reviewer decided the action should not run.",
    n8n: "Stop the business action and notify the right channel.",
  },
  {
    decision: "blocked",
    source: "Automatic policy",
    when: "Policy evaluation says the action is not allowed.",
    n8n: "Stop the business action and record the blocked reason.",
  },
  {
    decision: "expired",
    source: "System deadline",
    when: "The request passed its approval deadline before a valid decision.",
    n8n: "Stop the action or submit a new approval request.",
  },
  {
    decision: "cancelled",
    source: "Workflow or admin",
    when: "The source execution no longer needs a decision, or an admin cancels it.",
    n8n: "End the wait path and avoid running the final action.",
  },
];

const setupSteps = [
  {
    title: "Create a workflow API key",
    body: "Generate a key from the dashboard with the approval_requests:create scope. Store it in n8n credentials or an environment variable.",
  },
  {
    title: "Add the approval request node",
    body: "Use an HTTP Request node to submit the proposed action before the business action runs.",
  },
  {
    title: "Pause with a Wait node",
    body: "Configure n8n to wait for a resume webhook. Send that webhook URL as the continuation target.",
  },
  {
    title: "Resume from the signed decision",
    body: "HumanGate sends the approval result back to n8n. Your workflow verifies it and then decides what to run next.",
  },
];

export default function N8NSetupGuidePage() {
  return (
    <main className="min-h-screen bg-[linear-gradient(180deg,#f8f3eb_0%,#efe7dc_100%)] px-6 py-8 text-[#15110d] md:px-10">
      <div className="mx-auto max-w-6xl">
        <header className="rounded-[2rem] border border-black/10 bg-[#16302b] px-6 py-6 text-[#f5ecde] shadow-[0_24px_60px_rgba(0,0,0,0.12)] md:px-8">
          <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
            <div className="max-w-3xl">
              <p className="text-sm font-semibold uppercase tracking-[0.24em] text-[#9bc5bb]">
                n8n integration
              </p>
              <h1 className="mt-4 text-3xl font-semibold leading-tight md:text-5xl">
                Connect n8n to HumanGate approvals.
              </h1>
              <p className="mt-4 max-w-2xl text-sm leading-6 text-[#dbcdbb] md:text-base">
                Pause a workflow, review the proposed action, then resume n8n
                with a signed decision payload.
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row lg:flex-col">
              <Link
                className="rounded-md bg-[#f5ecde] px-4 py-3 text-center text-sm font-semibold text-[#15110d] transition hover:bg-white active:translate-y-px"
                href="/dashboard"
              >
                Create API key
              </Link>
              <Link
                className="rounded-md border border-white/15 px-4 py-3 text-center text-sm font-semibold text-[#f5ecde] transition hover:bg-white/10 active:translate-y-px"
                href="/inbox"
              >
                Open inbox
              </Link>
            </div>
          </div>
        </header>

        <section className="mt-8 grid gap-6 lg:grid-cols-[0.92fr_1.08fr]">
          <aside className="space-y-6">
            <div className="rounded-lg border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <h2 className="text-xl font-semibold">Setup path</h2>
              <div className="mt-5 space-y-4">
                {setupSteps.map((step) => (
                  <div key={step.title} className="rounded-lg border border-black/10 bg-[#fbf8f2] p-4">
                    <h3 className="text-sm font-semibold text-[#16302b]">
                      {step.title}
                    </h3>
                    <p className="mt-2 text-sm leading-6 text-black/65">
                      {step.body}
                    </p>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-lg border border-[#1f6f78]/20 bg-[#eef8f9] p-6 text-[#13505a] shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <h2 className="text-lg font-semibold">Boundary to keep clear</h2>
              <p className="mt-3 text-sm leading-6">
                HumanGate approves or rejects the proposed action. n8n remains
                responsible for issuing refunds, updating systems, publishing
                content, and retrying the final business operation.
              </p>
            </div>
          </aside>

          <div className="space-y-6">
            <section className="rounded-lg border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <h2 className="text-xl font-semibold">HTTP Request node</h2>
                  <p className="mt-2 text-sm leading-6 text-black/65">
                    Configure the node to send JSON and continue regardless of
                    whether the first response is allowed, pending, or blocked.
                  </p>
                </div>
                <code className="rounded-md bg-[#fbf8f2] px-3 py-2 text-xs font-semibold text-[#8d3419]">
                  {requestUrl}
                </code>
              </div>
              <div className="mt-5 grid gap-5">
                <CopySnippet label="Headers" language="http" value={headersSnippet} />
                <CopySnippet label="Body" language="json" value={bodySnippet} />
              </div>
            </section>

            <section className="rounded-lg border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <h2 className="text-xl font-semibold">Response handling</h2>
              <p className="mt-2 text-sm leading-6 text-black/65">
                If the response is pending, route the execution into the Wait
                node. If it is allowed or blocked, branch immediately.
              </p>
              <div className="mt-5">
                <CopySnippet
                  label="Pending response"
                  language="json"
                  value={pendingResponseSnippet}
                />
              </div>
            </section>

            <section className="rounded-lg border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <h2 className="text-xl font-semibold">Decision branches</h2>
              <p className="mt-2 text-sm leading-6 text-black/65">
                Use a Switch node after the callback. Branch on the decision
                field and keep rejected, blocked, expired, and cancelled paths
                away from the final business action.
              </p>
              <div className="mt-5 grid gap-3">
                {decisionCases.map((item) => (
                  <div
                    className="rounded-lg border border-black/10 bg-[#fbf8f2] p-4"
                    key={item.decision}
                  >
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                      <div>
                        <h3 className="font-mono text-sm font-semibold text-[#16302b]">
                          {item.decision}
                        </h3>
                        <p className="mt-2 text-sm leading-6 text-black/65">
                          {item.when}
                        </p>
                      </div>
                      <span className="w-fit rounded-md border border-[#1f6f78]/20 bg-[#eef8f9] px-3 py-1 text-xs font-semibold text-[#13505a]">
                        {item.source}
                      </span>
                    </div>
                    <p className="mt-3 rounded-md bg-white px-3 py-2 text-sm text-black/70">
                      n8n path: {item.n8n}
                    </p>
                  </div>
                ))}
              </div>
              <div className="mt-5">
                <CopySnippet
                  label="Decision examples"
                  language="json"
                  value={decisionExamplesSnippet}
                />
              </div>
            </section>

            <section className="rounded-lg border border-black/10 bg-white p-6 shadow-[0_20px_55px_rgba(0,0,0,0.08)]">
              <h2 className="text-xl font-semibold">Resume callback</h2>
              <p className="mt-2 text-sm leading-6 text-black/65">
                HumanGate sends a signed decision to the n8n resume webhook.
                Verify the signature before running the final action.
              </p>
              <div className="mt-5 grid gap-5">
                <CopySnippet
                  label="Decision payload"
                  language="json"
                  value={callbackSnippet}
                />
               
              </div>
            </section>
          </div>
        </section>
      </div>
    </main>
  );
}
