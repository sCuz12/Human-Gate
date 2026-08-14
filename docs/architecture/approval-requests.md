# Approval Requests

This document explains how approval requests work in the current API and what the intended architecture is for the next iterations.

An approval request is the persisted review process for a proposed action from an external workflow. The source workflow still owns the final business action. HumanGate only decides whether the proposed action is allowed, blocked, approved, or rejected, then delivers that decision back to the workflow.

## Core Concepts

An approval request binds together:

- the original proposed action;
- the source workflow that asked for approval;
- the policy result used at creation time;
- the human or automatic decision;
- the continuation target used to return the decision;
- delivery state;
- immutable audit events.

The important tables are:

```text
approval_requests
approval_decisions
continuation_targets
decision_deliveries
workflow_acknowledgements
audit_events
```

Every tenant-owned record is scoped by `workspace_id`.

## Public API Surface

External workflow endpoint:

```text
POST /api/v1/approval-requests
```

Human dashboard endpoints:

```text
GET  /api/v1/approval-requests
GET  /api/v1/approval-requests/{id}
GET  /api/v1/approval-requests/{id}/audit-events
GET  /api/v1/approval-requests/{id}/delivery
POST /api/v1/approval-requests/{id}/approve
POST /api/v1/approval-requests/{id}/reject
```

Planned but not implemented yet:

```text
POST /api/v1/approval-requests/{id}/approve-with-changes
POST /api/v1/approval-requests/{id}/acknowledgements
```

## Creation Flow

The source workflow submits a proposed action with an API key and an `Idempotency-Key`.

```text
workflow
  -> POST /api/v1/approval-requests
  -> API key authentication
  -> workspace derived from API key
  -> request decoded and validated
  -> approval.SubmitApprovalRequest
```

The `Idempotency-Key` is required and scoped to the workspace. If the same workspace submits the same key again, the API returns the existing approval request instead of creating a duplicate.

The current submit command contains:

```text
workspace id
idempotency key
action
context
source workflow identity
continuation target
authenticated actor
```

The request body cannot choose the workspace. The workspace comes from the authenticated API key.

## Action Integrity

On creation, the service builds a canonical JSON representation of the proposed action:

```json
{
  "type": "customer.refund",
  "title": "Refund customer",
  "description": "...",
  "parameters": {}
}
```

It then stores:

- `original_action`;
- `original_action_hash`.

The hash is SHA-256 over the canonical action JSON. This prevents later mutation of what was reviewed and gives decision delivery a stable action identity.

For an approval without edits, the approved action is the original action and the approved action hash matches the original action hash.

For a rejection or block, there is no approved action.

## Policy Evaluation

Policy evaluation currently happens synchronously during request creation.

```text
load active policy versions for workspace
  -> evaluate conditions in priority order
  -> use first matching policy
  -> fall back to workspace default effect
  -> snapshot matched policy onto approval request
```

Supported condition checks in the current MVP code:

- `action.type equals`;
- `source.platform equals`;
- `context.reversible equals`.

The database and product model are broader than the current evaluator. More condition types can be added behind the same creation flow.

Policy effects map to approval request status:

```text
allow            -> allowed
block            -> blocked
require_approval -> pending
```

When a policy matches, the request stores:

- `matched_policy_id`;
- `matched_policy_version_id`;
- `matched_policy_snapshot`.

The snapshot preserves the policy version that was used for the decision. Later policy edits must not change the meaning of existing approval requests.

## State Model

The database supports the full approval status enum:

```text
received
evaluating
allowed
blocked
pending
approved
approved_with_changes
rejected
expired
cancelled
```

The intended product state machine is:

```text
received
  -> evaluating
     -> allowed
     -> blocked
     -> pending
        -> approved
        -> approved_with_changes
        -> rejected
        -> expired
        -> cancelled
```

The current implementation performs creation and evaluation in one synchronous transaction. In practice, new requests currently go directly to one of:

```text
allowed
blocked
pending
```

Current manual resolution supports:

```text
pending -> approved
pending -> rejected
```

`approved_with_changes`, `expired`, and `cancelled` exist in the model but are not fully exposed through the API yet.

Final states are immutable during normal application flows.

## Automatic Outcomes

If policy evaluation returns `allow`, the request is immediately resolved as `allowed`.

If policy evaluation returns `block`, the request is immediately resolved as `blocked`.

For both outcomes, the service creates:

- an `approval_requests` row;
- a `continuation_targets` row;
- an `approval_decisions` row;
- a `decision_deliveries` row;
- audit events for receipt, decision, and delivery scheduling.

The request is still persisted even when no human is required. That preserves auditability, idempotency, policy snapshots, and delivery tracking.

## Pending Requests

If policy evaluation returns `require_approval`, the request is created as `pending` with `decision_required = true`.

Policy approval settings may set:

- `assigned_user_id`;
- `assigned_group_id`;
- `expires_at`.

If no assignee is set, the request is still pending. The intended MVP behavior is that workspace owners and administrators can see and resolve unassigned pending requests. Current authorization is less strict and only checks workspace membership.

## Human Review

Human dashboard requests use Supabase JWT authentication. The API attaches the authenticated user to the request context.

For list and read operations, the service:

```text
validates workspace id, request id, and user id
checks workspace membership
loads approval requests scoped by workspace
returns summaries for the UI
```

Current list pagination uses `limit` and `offset`, capped to a maximum limit of 100. Cursor pagination should replace this when inbox volume grows.

## Approve Or Reject Flow

The approve and reject endpoints call the same decision use case with a different decision value.

```text
POST /api/v1/approval-requests/{id}/approve
POST /api/v1/approval-requests/{id}/reject
```

The current transaction is:

```text
begin transaction
check workspace membership
lock approval request by workspace and id for update
ensure request is pending
ensure request is not expired
create approval decision
load continuation target
create decision delivery
resolve approval request
create decision audit event
create delivery scheduled audit event
commit
```

The row lock ensures only one concurrent transition from `pending` to a final state can succeed.

Approve creates:

```text
decision = approved
approval_request.status = approved
approved_action = original_action
approved_action_hash = original_action_hash
```

Reject creates:

```text
decision = rejected
approval_request.status = rejected
approved_action = null
approved_action_hash = null
```

## Authorization

Current implementation:

- external creation requires an API key with `approval_requests:create`;
- human list/read/approve/reject requires Supabase auth;
- human service methods check workspace membership.

Intended product behavior:

- viewers can read where policy allows, but cannot approve;
- approvers can approve assigned requests;
- owners and administrators can approve workspace requests;
- plain workspace membership is not enough to resolve a request.

This is an important gap. Before production use, approval resolution should enforce role and assignment rules, not membership alone.

## Continuation And Delivery

Every submitted approval request stores a continuation target. The continuation target describes how to return the decision to the original workflow.

Current fields include:

- strategy;
- platform;
- destination;
- encrypted configuration.

Decision delivery is tracked separately from approval request status.

```text
approval status  -> what was decided
delivery status  -> whether the workflow received it
```

Delivery statuses include:

```text
not_required
pending
attempting
delivered
acknowledged
retrying
permanently_failed
```

The API currently creates `decision_deliveries` rows. The worker is responsible for executing delivery attempts and updating delivery state.

Approval can be valid even if delivery fails.

## Audit Events

Audit events are append-only records of important transitions. Approval request flows currently write events such as:

```text
approval_request.received
approval_request.allowed
approval_request.blocked
approval_request.approved
approval_request.rejected
decision.delivery_scheduled
```

Audit metadata must stay safe. It should not contain:

- API keys;
- JWTs;
- continuation URLs;
- webhook secrets;
- full sensitive payloads;
- customer evidence content.

## Error Semantics

Common public error mappings:

```text
invalid request     -> 400
unauthenticated     -> 401
forbidden           -> 403
not found           -> 404
already resolved    -> 409
expired             -> 409
internal failure    -> 500
```

Public errors use the standard API shape:

```json
{
  "error": {
    "code": "request_resolved",
    "message": "This approval request has already been resolved."
  }
}
```

Internal errors may be logged with safe structured metadata.

## Invariants

The approval request system should preserve these invariants:

- every approval request belongs to exactly one workspace;
- tenant queries are scoped by workspace;
- `Idempotency-Key` is unique per workspace;
- original action data is immutable after creation;
- every created request has a continuation target;
- every resolved request has one decision;
- every decision has at most one delivery record;
- only `pending` requests can be manually approved or rejected;
- expired requests cannot be approved or rejected;
- final approval states do not return to `pending`;
- delivery state is separate from approval state;
- source workflows execute business actions, not HumanGate.

## Current Gaps

The current MVP implementation is intentionally narrow. Known gaps:

- `approved_with_changes` is modeled but not implemented;
- `cancelled` is modeled but not implemented;
- expiry scanning exists at the query/model level but is not exposed as a complete workflow here;
- approval authorization checks membership, not role plus assignment;
- policy condition support is smaller than the product policy model;
- delivery job execution is separate from the API and not described by the request handler itself;
- continuation destination storage is present, but production-grade encryption and logging controls must be verified before real customer data;
- list pagination is offset-based rather than cursor-based.

These gaps should be treated as product and security work before the approval system is considered production-complete.
