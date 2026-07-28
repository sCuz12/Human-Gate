# AGENTS.md

## Project overview

This repository contains the MVP for a centralized human-approval platform for AI agents and automated workflows.

The product receives proposed actions from systems such as:

* n8n;
* LangGraph;
* custom Go services;
* custom Python agents;
* other workflow and automation platforms.

The product evaluates configured policies and produces one of three outcomes:

* allow immediately;
* require human approval;
* block.

When human approval is required, the action appears in a centralized inbox. An authorized user can:

* approve;
* edit and approve;
* reject.

After the decision, the platform sends a signed decision back to the source workflow. The source workflow remains responsible for performing the final business action.

The platform does not directly:

* issue refunds;
* update CRM records;
* publish content;
* delete data;
* send quotations;
* change supplier details.

This is an approval orchestration product, not an execution engine.

---

# 1. Core architecture

The system uses a monorepo containing:

```text
/
├── apps/
│   ├── api/                   Go API application
│   ├── worker/                Go background worker
│   └── web/                   Next.js frontend
│
├── internal/
│   ├── approval/              Approval domain
│   ├── policy/                Policy evaluation
│   ├── integration/           n8n, LangGraph and webhook adapters
│   ├── notification/          Slack and email notifications
│   ├── identity/              Users, workspaces and permissions
│   ├── audit/                 Immutable audit events
│   ├── delivery/              Decision delivery and retry logic
│   └── platform/              Shared infrastructure
│
├── db/
│   ├── queries/               sqlc SQL query files
│   ├── generated/             sqlc-generated Go code
│   └── seeds/                 Development seed data
│
├── supabase/
│   ├── migrations/            Database migrations
│   ├── seed.sql               Local development seed
│   └── config.toml            Supabase local configuration
│
├── packages/
│   ├── contracts/             Shared API schemas where appropriate
│   └── ui/                    Optional shared UI components
│
├── docs/
│   ├── architecture/
│   ├── integrations/
│   └── decisions/
│
├── scripts/
├── Makefile
├── docker-compose.yml
├── sqlc.yaml
├── go.mod
├── go.sum
└── AGENTS.md
```

The exact structure may evolve, but dependencies must continue to point inward toward domain logic.

---

# 2. Technology stack

## Backend

Use:

* Go;
* `net/http`;
* `github.com/go-chi/chi`;
* `github.com/jackc/pgx/v5`;
* `github.com/jackc/pgx/v5/pgxpool`;
* `sqlc`;
* River for PostgreSQL-backed jobs;
* structured logging using `log/slog`;
* OpenTelemetry where observability is introduced.

Do not introduce a heavy Go web framework unless an explicit architecture decision approves it.

Chi is the HTTP router, not the application architecture.

Business logic must not live inside HTTP handlers.

## Frontend

Use:

* Next.js App Router;
* TypeScript;
* React Server Components by default;
* Client Components only when browser interaction is required;
* Tailwind CSS;
* shadcn/ui or equivalent accessible primitives;
* React Hook Form for complex forms;
* Zod for frontend validation;
* TanStack Query only where client-side server-state management is genuinely needed.

## Database and infrastructure

Use:

* Supabase PostgreSQL;
* Supabase Auth;
* Supabase Storage for approval evidence where needed;
* Supabase Realtime only where it adds clear value;
* Supabase CLI migrations;
* PostgreSQL Row Level Security where browser-accessible Supabase data is used;
* River tables and workers for asynchronous jobs.

---

# 3. Architectural boundaries

The system is divided into the following logical layers:

```text
Transport
   ↓
Application service
   ↓
Domain
   ↓
Repository interfaces
   ↓
PostgreSQL / external adapters
```

## Transport layer

Includes:

* HTTP handlers;
* webhook handlers;
* authentication middleware;
* request decoding;
* response encoding;
* transport-specific validation.

The transport layer must not contain business policy.

## Application layer

Coordinates use cases such as:

* submitting an approval request;
* evaluating policies;
* recording a decision;
* assigning approvers;
* scheduling decision delivery;
* expiring an approval request;
* acknowledging a workflow continuation.

Application services may start database transactions.

## Domain layer

Contains core concepts and invariants:

* approval request;
* proposed action;
* policy;
* policy condition;
* decision;
* approver assignment;
* continuation target;
* decision delivery;
* audit event.

The domain layer must not import:

* Chi;
* pgx;
* sqlc-generated packages;
* River;
* Supabase SDKs;
* HTTP-specific packages.

## Infrastructure layer

Implements:

* PostgreSQL repositories;
* River jobs;
* Slack notifications;
* webhook delivery;
* signing and verification;
* Supabase identity adapters;
* object storage;
* external integration clients.

---

# 4. Domain terminology

Use these terms consistently.

## Proposed action

An action that an external workflow wants to perform.

Examples:

* refund a customer;
* publish an article;
* update supplier details;
* delete records;
* send a quotation.

## Approval request

The persisted review process created for a proposed action.

## Policy

A set of conditions and an outcome.

Example:

```text
WHEN action type is customer.refund
AND amount is greater than 50
THEN require approval from support_manager
```

## Decision

A final human or automatic result:

* allowed;
* approved;
* approved with changes;
* rejected;
* blocked;
* expired;
* cancelled.

## Continuation target

Instructions describing how to return the decision to the original workflow.

Examples:

* n8n resume webhook;
* generic webhook;
* LangGraph customer callback;
* polling.

## Delivery

An attempt to send a resolved decision to the continuation target.

## Acknowledgement

A message from the source workflow confirming that it received the decision or resumed successfully.

---

# 5. Backend package structure

Each domain should follow a structure similar to:

```text
internal/approval/
├── domain.go
├── errors.go
├── repository.go
├── service.go
├── commands.go
├── queries.go
└── validation.go
```

Infrastructure implementations belong outside the domain package:

```text
internal/platform/postgres/approval_repository.go
internal/platform/river/deliver_decision_job.go
internal/platform/webhook/client.go
```

HTTP handlers belong under the API application:

```text
apps/api/internal/httpapi/
├── router.go
├── middleware/
├── approvals/
├── policies/
├── integrations/
└── workspaces/
```

Do not create generic packages named:

* `utils`;
* `helpers`;
* `common`;
* `misc`.

Name packages after the capability they provide.

---

# 6. Go coding conventions

## General rules

All Go code must:

* pass `go test ./...`;
* pass `go vet ./...`;
* be formatted with `gofmt`;
* avoid global mutable state;
* propagate `context.Context`;
* wrap errors with relevant context;
* avoid unnecessary interfaces;
* use dependency injection through constructors;
* keep exported APIs small;
* use explicit types for important domain concepts.

Use:

```go
type ApprovalRequestID uuid.UUID
type WorkspaceID uuid.UUID
type DecisionID uuid.UUID
```

when stronger domain types prevent accidental misuse.

Do not create an interface merely to make a struct mockable.

Create an interface when:

* multiple implementations exist;
* the consumer needs only a narrow behaviour;
* the boundary represents infrastructure.

Interfaces should generally be defined by the consuming package.

## Constructors

Use explicit constructors:

```go
func NewService(
    requests RequestRepository,
    policies PolicyRepository,
    jobs JobClient,
    clock Clock,
) *Service {
    return &Service{
        requests: requests,
        policies: policies,
        jobs: jobs,
        clock: clock,
    }
}
```

Validate required dependencies in constructors where practical.

## Context

The first argument of I/O or request-scoped methods must be:

```go
ctx context.Context
```

Do not store request contexts inside long-lived structs.

Do not use `context.Background()` inside request handling or job execution when an existing context is available.

## Errors

Define domain errors using sentinel errors or typed errors where callers need to branch:

```go
var (
    ErrRequestNotFound      = errors.New("approval request not found")
    ErrRequestResolved      = errors.New("approval request already resolved")
    ErrRequestExpired       = errors.New("approval request expired")
    ErrUnauthorizedApprover = errors.New("user cannot approve this request")
)
```

Wrap infrastructure errors:

```go
return fmt.Errorf("load approval request %s: %w", requestID, err)
```

Do not expose internal error messages directly to API clients.

## Logging

Use structured logging:

```go
logger.InfoContext(
    ctx,
    "approval request created",
    "request_id", request.ID,
    "workspace_id", request.WorkspaceID,
    "action_type", request.Action.Type,
)
```

Never log:

* API keys;
* webhook secrets;
* JWTs;
* continuation URLs;
* full sensitive action payloads;
* customer evidence content.

---

# 7. HTTP API conventions

Base path:

```text
/api/v1
```

Examples:

```text
POST   /api/v1/approval-requests
GET    /api/v1/approval-requests
GET    /api/v1/approval-requests/{id}
POST   /api/v1/approval-requests/{id}/approve
POST   /api/v1/approval-requests/{id}/approve-with-changes
POST   /api/v1/approval-requests/{id}/reject
POST   /api/v1/approval-requests/{id}/acknowledgements

GET    /api/v1/policies
POST   /api/v1/policies
GET    /api/v1/policies/{id}
PATCH  /api/v1/policies/{id}
DELETE /api/v1/policies/{id}

GET    /api/v1/integrations
POST   /api/v1/api-keys
DELETE /api/v1/api-keys/{id}
```

## Request handling

Handlers should:

1. Authenticate.
2. Decode input.
3. Perform transport-level validation.
4. Call an application service.
5. Map domain errors to HTTP responses.
6. Encode a standard response.

Handlers must not:

* run SQL;
* evaluate policies;
* enqueue River jobs directly;
* build audit event sequences;
* implement authorization rules inline.

## Error response

Use a consistent shape:

```json
{
  "error": {
    "code": "approval_request_expired",
    "message": "This approval request has expired.",
    "request_id": "req_123"
  }
}
```

Do not return stack traces.

## Pagination

Use cursor-based pagination for large or time-ordered collections.

Initial MVP endpoints may use limit and offset only if the expected dataset is small, but new high-volume endpoints should prefer cursors.

## Idempotency

Creating an approval request requires:

```text
Idempotency-Key
```

The key is scoped to a workspace.

A repeated request with the same key must:

* return the original approval request;
* not create another request;
* not enqueue duplicate notifications;
* not evaluate the request as a new action.

---

# 8. Database access

## PostgreSQL driver

Use `pgxpool` for pooled PostgreSQL connections.

The backend connects directly to Supabase PostgreSQL.

Use the appropriate Supabase connection type:

* direct database connection for migrations;
* pooled connection for deployed API and worker services where required by the hosting environment.

## SQL access

Use `sqlc` for standard database operations.

Write SQL explicitly under:

```text
db/queries/
```

Generated code belongs under:

```text
db/generated/
```

Never manually edit generated files.

After changing query files, run:

```bash
make sqlc
```

or:

```bash
sqlc generate
```

## ORM policy

Do not introduce GORM, Ent or another ORM without an approved architecture decision.

The project intentionally uses:

```text
PostgreSQL + SQL + sqlc + pgx
```

because approval, policy and delivery logic benefit from explicit:

* transactions;
* locking;
* constraints;
* JSONB operations;
* PostgreSQL indexes;
* concurrent state transitions.

## Transactions

Application services own transaction boundaries.

Repositories must support transaction-scoped execution where needed.

Important state transitions must happen atomically.

Example approval transaction:

```text
BEGIN

1. Lock pending approval request.
2. Validate status and deadline.
3. Validate approver authorization.
4. Insert decision.
5. Update approval request status.
6. Insert audit event.
7. Insert River delivery job or transactional job record.

COMMIT
```

Use row locking when resolving concurrent approvals:

```sql
SELECT *
FROM approval_requests
WHERE id = $1
FOR UPDATE;
```

Only one transition from `pending` to a final state may succeed.

---

# 9. Database migrations

Supabase CLI migrations are the source of truth for the application schema.

Migration files belong under:

```text
supabase/migrations/
```

Use timestamped migration names:

```text
20260728120000_create_workspaces.sql
20260728121500_create_approval_requests.sql
20260728123000_create_policies.sql
20260728124500_create_decision_deliveries.sql
```

Create migrations with:

```bash
supabase migration new create_approval_requests
```

Apply locally with:

```bash
supabase db reset
```

Apply to a linked remote environment through the approved deployment process:

```bash
supabase db push
```

## Migration rules

Every schema change must be represented by a committed migration.

Never manually change production schema through the dashboard without creating the equivalent migration.

Migrations must:

* be deterministic;
* be reviewable;
* avoid destructive changes without an explicit migration strategy;
* include indexes required by new access paths;
* include constraints enforcing domain invariants;
* include RLS updates when exposed Supabase tables change;
* succeed against an empty local database;
* succeed against the previous production-compatible schema.

## Forward-only production policy

Prefer forward corrective migrations over editing migrations that have already been applied to shared environments.

A migration file may be changed only when it has not been applied outside the author's isolated local environment.

## River migrations

River owns its internal job-queue schema.

River migrations must be installed and managed using River's supported migration tooling.

Do not manually alter River's internal tables.

Application migrations and River migrations are separate concerns:

```text
supabase/migrations/
→ application-owned database schema

River migration tooling
→ River-owned job infrastructure
```

---

# 10. Core database design rules

All tenant-owned tables must contain:

```text
workspace_id
```

All tenant queries must scope by workspace.

Never load a resource by ID alone when the caller is workspace-scoped.

Incorrect:

```sql
SELECT * FROM approval_requests WHERE id = $1;
```

Correct:

```sql
SELECT *
FROM approval_requests
WHERE workspace_id = $1
  AND id = $2;
```

## Timestamps

Use:

```sql
timestamptz
```

Store timestamps in UTC.

Use clear names:

* `created_at`;
* `updated_at`;
* `resolved_at`;
* `expires_at`;
* `delivered_at`;
* `acknowledged_at`.

## Identifiers

Use UUIDs.

Prefer generated UUIDs in PostgreSQL or in the application consistently.

Do not expose sequential internal identifiers.

## JSONB

JSONB is appropriate for:

* original action payload;
* approved action payload;
* policy condition definitions;
* policy snapshots;
* integration metadata;
* evidence metadata.

Do not use JSONB as a replacement for every relational field.

Fields used for:

* filtering;
* sorting;
* joins;
* authorization;
* uniqueness;
* state transitions;

should generally be explicit columns.

## Constraints

Use database constraints for critical invariants.

Examples:

```sql
CHECK (status IN (...))
CHECK (resolved_at IS NOT NULL OR status = 'pending')
UNIQUE (workspace_id, idempotency_key)
UNIQUE (approval_request_id)
```

Application validation complements database constraints; it does not replace them.

---

# 11. Suggested database tables

Initial application-owned tables:

```text
users
workspaces
workspace_members
api_keys

approver_groups
approver_group_members

integrations
source_workflows

policies
policy_versions

approval_requests
approval_assignments
approval_evidence
approval_decisions
approval_comments

continuation_targets
decision_deliveries
workflow_acknowledgements

audit_events
notification_preferences
```

River creates and owns its separate job-related tables.

## Policy versions

Policies must be versioned or snapshotted.

Existing approval requests must preserve the exact policy configuration used when the request was evaluated.

Changing a policy must not retroactively alter an already-pending request unless an explicit reevaluation feature is implemented.

---

# 12. Background jobs

Use River for asynchronous and retryable work.

River is backed by PostgreSQL, allowing the project to avoid an additional queue dependency during the MVP.

## Initial job types

Create River jobs for:

```text
DeliverDecision
SendSlackNotification
SendEmailNotification
ExpireApprovalRequest
EscalateApprovalRequest
CleanupSensitiveContinuationData
RecordIntegrationHealth
```

## Job argument example

```go
type DeliverDecisionArgs struct {
    DeliveryID uuid.UUID `json:"delivery_id"`
}

func (DeliverDecisionArgs) Kind() string {
    return "deliver_decision"
}
```

Pass identifiers to jobs, not entire mutable domain objects.

The worker must reload the latest database state before acting.

## Job idempotency

Every job must be safe to retry.

A job may be executed more than once due to:

* worker crashes;
* network timeouts;
* deployment restarts;
* ambiguous external responses;
* manual retry.

Before performing an irreversible external action such as delivering a callback, check the current delivery state.

Example:

```text
If delivery is already marked delivered:
    return success without redelivery.
```

Where duplicate callback delivery is possible, use a stable decision ID so the receiving workflow can reject duplicates.

## Job transactions

When creating a decision and scheduling delivery, use transaction-safe job insertion where supported.

The decision and delivery job must not become inconsistent.

The desired invariant is:

```text
A committed approval decision always has a durable delivery job.
```

## Retry policy

Delivery jobs should use controlled exponential backoff.

Suggested MVP schedule:

```text
Attempt 1: immediately
Attempt 2: 5 seconds
Attempt 3: 30 seconds
Attempt 4: 2 minutes
Attempt 5: 10 minutes
Attempt 6: 1 hour
```

After final failure:

* mark delivery as permanently failed;
* create an audit event;
* notify a workspace administrator;
* expose manual retry in the UI.

---

# 13. Approval state machine

The approval request state machine is:

```text
received
    ↓
evaluating
    ├── allowed
    ├── blocked
    └── pending
          ├── approved
          ├── approved_with_changes
          ├── rejected
          ├── expired
          └── cancelled
```

Final states are immutable unless an explicit administrative correction mechanism is designed.

A resolved request must never return to `pending`.

## Delivery state machine

Decision delivery state is separate:

```text
not_required
pending
attempting
delivered
acknowledged
retrying
permanently_failed
```

Do not combine approval state and delivery state into a single field.

An approval may be valid even when workflow delivery fails.

---

# 14. Policy engine

The MVP policy engine supports:

## Conditions

* action type equals;
* source platform equals;
* source workflow equals;
* target system equals;
* numeric value greater than;
* numeric value greater than or equal;
* numeric value less than;
* numeric value less than or equal;
* string equals;
* boolean equals;
* risk level equals;
* reversible equals.

## Effects

* allow;
* require approval;
* block.

## Policy evaluation rules

* evaluate active policies only;
* evaluate by explicit priority;
* stop according to documented conflict resolution;
* record the matched policy version;
* return deterministic results;
* do not call an LLM to decide whether a policy matches;
* do not use model confidence as an authorization mechanism.

A policy engine must produce the same result for the same:

* action;
* workspace;
* active policy versions;
* evaluation time where deadlines matter.

## Policy conflict strategy

Initial recommendation:

1. `block` overrides every other result.
2. `require_approval` overrides `allow`.
3. Higher-priority policy wins within the same effect.
4. If no policy matches, use the workspace default.

Default MVP workspace behaviour:

```text
require_approval
```

This may later become configurable.

---

# 15. Action integrity and signing

An approval must be bound to the exact action reviewed.

## Canonical action

Create a deterministic canonical JSON representation.

Do not hash raw JSON input because:

* object key order may differ;
* whitespace may differ;
* numeric formatting may differ.

Generate:

```text
original_action_hash = SHA-256(canonical_original_action)
```

For edited approvals:

```text
approved_action_hash = SHA-256(canonical_approved_action)
```

## Signed decision

Decision signatures must cover:

* workspace ID;
* approval request ID;
* decision ID;
* decision type;
* original action hash;
* approved action hash;
* intended audience;
* issued time;
* expiry time.

Use a well-defined signing implementation.

Do not invent custom cryptographic algorithms.

Initial MVP may use HMAC-SHA256 with per-workspace integration secrets.

Design the signing abstraction so asymmetric signing can be added later.

## Replay prevention

Every decision has a globally unique decision ID.

Integration documentation must instruct source workflows to process each decision ID only once.

---

# 16. Authentication and authorization

Use Supabase Auth for user authentication.

The Next.js application obtains the user session through Supabase.

The Go API validates Supabase-issued access tokens.

The API must not trust user identity fields sent in request bodies.

## Authorization model

Workspace roles:

```text
owner
administrator
approver
viewer
```

Permissions should be explicit.

Examples:

| Action                   | Owner | Administrator | Approver | Viewer |
| ------------------------ | ----: | ------------: | -------: | -----: |
| View inbox               |   Yes |           Yes | Assigned |    Yes |
| Approve assigned request |   Yes |           Yes |      Yes |     No |
| Create policy            |   Yes |           Yes |       No |     No |
| Manage integrations      |   Yes |           Yes |       No |     No |
| Manage workspace members |   Yes |       Limited |       No |     No |
| View audit log           |   Yes |           Yes |  Limited |    Yes |

Do not rely only on frontend route protection.

Every protected action must be authorized in the Go API.

## Service authentication

External workflows authenticate using scoped API keys.

API keys must:

* be shown only once;
* be stored as hashes;
* have a display prefix;
* belong to one workspace;
* support revocation;
* support scopes;
* record last use.

Example scopes:

```text
approval_requests:create
approval_requests:read
acknowledgements:create
```

---

# 17. Supabase usage rules

Supabase is the managed infrastructure provider, but the Go backend remains the primary business API.

## Use Supabase for

* PostgreSQL;
* authentication;
* optional object storage;
* optional Realtime notifications;
* local database development;
* migrations and environment management.

## Do not bypass the Go backend for

* creating approval requests;
* resolving approvals;
* evaluating policies;
* assigning approvers;
* creating decisions;
* scheduling workflow continuation;
* mutating audit-critical records.

Business mutations must pass through Go application services.

## Frontend direct access

The frontend may access Supabase directly for:

* authentication session handling;
* explicitly approved storage uploads;
* selected read-only Realtime subscriptions.

Do not let the browser write directly to approval or policy tables.

## Row Level Security

Enable RLS on any table accessible through Supabase's browser-facing APIs.

RLS is defence in depth and must not replace API-level authorization.

Service-role keys must never appear in browser code.

---

# 18. Next.js frontend architecture

Use the App Router.

Suggested structure:

```text
apps/web/
├── src/
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   └── callback/
│   │   │
│   │   ├── (dashboard)/
│   │   │   ├── layout.tsx
│   │   │   ├── inbox/
│   │   │   ├── approvals/[id]/
│   │   │   ├── policies/
│   │   │   ├── integrations/
│   │   │   ├── team/
│   │   │   └── audit/
│   │   │
│   │   └── api/
│   │
│   ├── components/
│   │   ├── approvals/
│   │   ├── policies/
│   │   ├── integrations/
│   │   └── ui/
│   │
│   ├── lib/
│   │   ├── api/
│   │   ├── auth/
│   │   ├── supabase/
│   │   └── validation/
│   │
│   └── types/
│
├── public/
├── package.json
└── tsconfig.json
```

## Server versus Client Components

Use Server Components by default for:

* page layout;
* initial approval list;
* request details;
* policies;
* audit data.

Use Client Components for:

* decision forms;
* interactive filters;
* policy condition builders;
* modals;
* live notification state;
* browser-only Supabase Auth operations.

Do not mark an entire route tree with `"use client"` merely because one child is interactive.

## API access

Create one typed frontend API client.

Do not scatter raw `fetch()` calls across components.

Example:

```text
src/lib/api/approvals.ts
src/lib/api/policies.ts
src/lib/api/integrations.ts
```

The frontend must call the Go API for business operations.

## Validation

Use Zod for frontend form validation.

The Go API remains authoritative and must independently validate every request.

Do not share generated validation logic blindly between TypeScript and Go if doing so weakens either side.

---

# 19. Frontend design principles

The approval detail page must prioritize decision quality.

Display:

* proposed action;
* reason;
* evidence;
* before-and-after values;
* affected systems;
* reversibility;
* policy matched;
* approver assignment;
* deadline;
* source workflow;
* audit timeline.

Do not make model confidence the main visual signal.

High-risk actions must not be approved accidentally.

Use:

* explicit confirmation;
* clear primary and destructive actions;
* visible edited values;
* warnings for irreversible actions;
* disabled actions after expiration;
* authenticated web confirmation for sensitive requests.

Accessibility requirements:

* keyboard navigation;
* visible focus states;
* semantic controls;
* sufficient contrast;
* screen-reader labels;
* no colour-only status meaning.

---

# 20. Integration architecture

All source platforms use the same core approval API.

Platform-specific behaviour belongs in adapters.

```text
n8n adapter
LangGraph adapter
Generic webhook adapter
Polling adapter
        ↓
Common approval API
        ↓
Policy engine
        ↓
Approval inbox
        ↓
Decision delivery adapters
```

## Generic submission contract

A submission includes:

```json
{
  "action": {
    "type": "customer.refund",
    "title": "Refund £82 to customer",
    "parameters": {
      "amount": 82,
      "currency": "GBP"
    }
  },
  "context": {
    "reason": "Customer supplied evidence of damaged goods",
    "reversible": false
  },
  "source": {
    "platform": "n8n",
    "workflow_id": "workflow_123",
    "execution_id": "execution_456"
  },
  "continuation": {
    "strategy": "webhook",
    "url": "https://..."
  }
}
```

## Integration boundary

Adapters translate platform-specific continuation information into a common internal model.

The approval domain must not contain n8n- or LangGraph-specific logic.

---

# 21. n8n integration rules

The initial n8n integration uses:

* HTTP Request node;
* Wait node;
* unique resume webhook;
* signed callback payload;
* decision switch.

The source n8n workflow owns:

* paused execution state;
* destination credentials;
* final execution;
* business-action retries.

The platform stores the n8n resume URL encrypted.

Never:

* return the resume URL to frontend clients;
* log the resume URL;
* expose it in audit events;
* treat possession of the URL alone as sufficient authorization.

The callback also requires a valid signed decision.

---

# 22. LangGraph integration rules

LangGraph integrations use:

* stable thread ID;
* durable checkpointer;
* `interrupt()`;
* customer-owned resume endpoint initially;
* `Command(resume=...)`.

The LangGraph application owns:

* graph state;
* thread persistence;
* model and tool credentials;
* tool execution.

Approval request creation must be idempotent because interrupt nodes may replay code during resumption.

Use a stable key such as:

```text
langgraph:{thread_id}:{tool_call_id}
```

Do not create an approval request with a random idempotency key from inside a replayable node.

---

# 23. Audit requirements

Audit events are immutable.

Every important transition creates an audit event.

Examples:

```text
approval_request.received
approval_request.validated
policy.evaluated
policy.matched
approval_request.assigned
approval_request.viewed
approval_request.approved
approval_request.approved_with_changes
approval_request.rejected
approval_request.expired
decision.delivery_scheduled
decision.delivery_attempted
decision.delivered
decision.delivery_failed
workflow.acknowledged
```

An audit event should contain:

* event ID;
* workspace ID;
* request ID;
* event type;
* actor type;
* actor ID where applicable;
* timestamp;
* safe structured metadata.

Do not store secrets inside audit metadata.

Do not update or delete audit events through normal application flows.

---

# 24. Testing strategy

## Unit tests

Write unit tests for:

* policy matching;
* policy priority;
* state transitions;
* authorization;
* action canonicalization;
* signing;
* signature validation;
* expiration;
* approver assignment;
* edit validation.

Use table-driven tests.

## Repository tests

Test PostgreSQL repositories against a real PostgreSQL instance.

Do not mock SQL generated by sqlc for repository correctness tests.

Repository tests must verify:

* constraints;
* transactions;
* concurrent approval attempts;
* idempotency;
* workspace isolation;
* JSONB persistence;
* row locking.

## Integration tests

Write tests for:

* submit → pending → approve → delivery;
* submit → allow;
* submit → block;
* edit and approve;
* expired request;
* failed webhook retry;
* duplicate callback protection;
* duplicate submission;
* unauthorized approval;
* n8n-style resume callback;
* LangGraph-style callback payload.

## End-to-end tests

Use Playwright for core user journeys:

1. Sign in.
2. Open inbox.
3. Review request.
4. Approve request.
5. Confirm status.
6. Create a policy.
7. Test integration.

Do not depend only on snapshots.

Assert meaningful user-visible behaviour.

---

# 25. Security requirements

Every change must consider:

* tenant isolation;
* broken object-level authorization;
* API key leakage;
* webhook forgery;
* replay attacks;
* duplicate delivery;
* action payload mutation;
* self-approval;
* expired approvals;
* sensitive evidence exposure;
* continuation URL leakage;
* audit tampering.

## Mandatory security rules

* Never trust workspace IDs supplied by clients without authorization.
* Never expose Supabase service-role credentials to the frontend.
* Never store raw API keys.
* Never log access tokens.
* Never let an approval modify immutable original action data.
* Never allow expired requests to resolve.
* Never enqueue duplicate business decisions for the same request.
* Never execute customer business actions from the MVP platform.

---

# 26. Observability

Use structured logs with stable fields:

```text
request_id
trace_id
workspace_id
approval_request_id
decision_id
delivery_id
job_kind
integration_type
```

Track metrics for:

* approval requests created;
* policy outcomes;
* pending approvals;
* approval latency;
* rejection rate;
* edit rate;
* expired requests;
* delivery success;
* delivery retries;
* permanent delivery failures;
* worker job latency;
* API error rate.

Do not include high-cardinality raw payloads as metric labels.

---

# 27. Local development

Expected commands:

```bash
make dev
make api
make worker
make web
make test
make lint
make sqlc
make db-reset
make db-migration name=create_approval_requests
```

Suggested `make dev` behaviour:

1. Start local Supabase.
2. Apply migrations.
3. Start the Go API.
4. Start the River worker.
5. Start the Next.js frontend.

The repository must include:

```text
.env.example
```

Never commit real secrets.

---

# 28. Environment variables

Backend:

```text
APP_ENV
HTTP_ADDR
DATABASE_URL
SUPABASE_URL
SUPABASE_JWT_ISSUER
SUPABASE_JWKS_URL
APP_ENCRYPTION_KEY
DECISION_SIGNING_KEY
SLACK_CLIENT_ID
SLACK_CLIENT_SECRET
SLACK_SIGNING_SECRET
PUBLIC_APP_URL
```

Frontend:

```text
NEXT_PUBLIC_SUPABASE_URL
NEXT_PUBLIC_SUPABASE_ANON_KEY
NEXT_PUBLIC_API_URL
```

Worker:

```text
DATABASE_URL
APP_ENCRYPTION_KEY
DECISION_SIGNING_KEY
WORKER_QUEUES
```

Validate environment configuration at startup.

Fail fast when required configuration is missing.

---

# 29. Pull request rules for coding agents

Before changing code:

1. Read this file.
2. Inspect the relevant domain package.
3. Inspect existing tests.
4. Inspect database migrations and sqlc queries.
5. Identify security and tenancy implications.
6. Avoid unrelated refactoring.

Every pull request must:

* explain the user-facing behaviour changed;
* list database migrations;
* list new jobs;
* list new environment variables;
* include or update tests;
* preserve workspace isolation;
* preserve idempotency;
* preserve auditability;
* document integration contract changes.

Do not:

* change public API contracts silently;
* edit applied migration files;
* modify generated sqlc files manually;
* bypass application services from handlers;
* add a dependency without justification;
* add a new infrastructure service when PostgreSQL can safely satisfy the MVP requirement;
* introduce an ORM;
* introduce a generic abstraction before two concrete use cases require it.

---

# 30. Definition of done

A backend feature is complete when:

* domain invariants are defined;
* authorization is implemented;
* workspace isolation is enforced;
* application logic is tested;
* repository queries are implemented through sqlc;
* migrations are included where required;
* audit events are created;
* idempotency implications are handled;
* jobs are retry-safe;
* API errors are documented;
* logs avoid sensitive information.

A frontend feature is complete when:

* loading state exists;
* empty state exists;
* error state exists;
* authorization failure is handled;
* keyboard interaction works;
* forms have validation;
* actions cannot be submitted twice accidentally;
* resolved approvals cannot be edited;
* tests cover critical behaviour.

An integration feature is complete when:

* request submission is authenticated;
* idempotency is supported;
* action hashing is verified;
* callback signing is verified;
* retries are safe;
* duplicate decisions are safe;
* secrets are not exposed;
* failure status is visible;
* integration documentation exists.

---

# 31. Initial implementation order

Build in this order:

## Phase 1: Foundation

* Supabase local environment;
* base migrations;
* Go API;
* pgx pool;
* sqlc;
* authentication;
* workspace model;
* API keys.

## Phase 2: Approval core

* action submission;
* idempotency;
* action canonicalization;
* policy evaluation;
* approval request state machine;
* audit events.

## Phase 3: Inbox

* approval list;
* approval detail;
* approve;
* reject;
* edit and approve;
* authorization.

## Phase 4: Delivery

* River;
* decision-delivery jobs;
* HMAC signing;
* webhook callbacks;
* retries;
* delivery dashboard;
* acknowledgements.

## Phase 5: n8n

* importable workflow template;
* Wait node integration;
* signature verification;
* refund demo;
* setup documentation.

## Phase 6: Policies

* policy CRUD;
* condition builder;
* priority;
* version snapshots;
* workspace default behaviour.

## Phase 7: Notifications

* Slack OAuth;
* approver notifications;
* expiration jobs;
* escalation.

## Phase 8: LangGraph

* Python integration helper;
* interrupt example;
* customer resume endpoint;
* stable idempotency pattern;
* documentation.

---

# 32. Product boundary

The MVP platform owns:

```text
Approval request ingestion
Policy evaluation
Approver routing
Central inbox
Human decisions
Decision signing
Decision delivery
Audit history
```

The connected workflow owns:

```text
Workflow state
Destination credentials
Business logic
Final action execution
Execution retries
Compensation or rollback
```

Do not cross this boundary without an explicit future product decision.

The core promise is:

> The platform determines whether a workflow may continue. The workflow remains responsible for what happens after that decision.
