# apps/api Architecture

`apps/api` is the Go HTTP API process for HumanGate. It exposes workflow-facing and human-facing endpoints, authenticates callers, translates HTTP requests into application commands, and returns stable API responses.

It is not the business logic layer. Approval orchestration, policy evaluation, identity rules, delivery scheduling, audit creation, and signing belong in internal application/domain packages and infrastructure adapters.

## Responsibilities

`apps/api` owns:

- process startup and shutdown;
- configuration loading;
- database pool creation;
- HTTP routing;
- HTTP middleware;
- request decoding and response encoding;
- transport-level validation;
- authentication middleware and HTTP auth challenge behavior;
- mapping domain/application errors to public API errors.

`apps/api` must not own:

- policy evaluation;
- approval state transitions;
- authorization rules beyond attaching authenticated identity to the request context;
- SQL queries;
- River job execution;
- decision signing internals;
- notification delivery;
- webhook callback delivery;
- audit event sequencing.

## Current Package Shape

```text
apps/api/
├── main.go
└── internal/httpapi/
    ├── router.go
    ├── middleware/
    ├── health/
    ├── approvals/
    ├── apikeys/
    ├── policies/
    └── workspaces/
```

`main.go` is the process entrypoint. It loads configuration, creates infrastructure clients, builds the router, starts `net/http`, and handles graceful shutdown.

`internal/httpapi/router.go` owns the HTTP surface. It wires middleware, handlers, and route groups.

Feature handler packages under `internal/httpapi/*` are transport adapters. They decode HTTP input, call application services, and encode responses.

## Dependency Direction

The intended direction is:

```text
apps/api
  -> internal/httpapi
  -> internal/{approval,policy,identity,delivery}
  -> internal/platform/*
  -> db/generated
```

Domain model code must not import `apps/api`, `chi`, `net/http`, River, Supabase SDKs, or sqlc-generated packages.

Application services may coordinate use cases and transactions. Infrastructure adapters provide PostgreSQL, River, webhook, signing, and notification implementations.

## Request Flow

### External Workflow Submission

```text
POST /api/v1/approval-requests
  -> API key authentication
  -> decode submission payload
  -> require Idempotency-Key
  -> submit approval request use case
  -> evaluate policies
  -> persist request, continuation target, audit event, and automatic decision when applicable
  -> return request id and status
```

The workspace is derived from the authenticated API key. The request body must not be trusted as the source of workspace authority.

### Human Dashboard Actions

```text
GET/POST /api/v1/*
  -> Supabase JWT authentication
  -> attach authenticated user to context
  -> decode request
  -> call use case with user id and workspace id
  -> service authorizes access inside the workspace
  -> return response
```

Frontend route protection is not sufficient. Every protected mutation must be authorized by the Go API.

## Authentication Boundaries

The API has two caller classes:

- external workflows, authenticated by scoped workspace API keys;
- human users, authenticated by Supabase-issued JWTs.

Workflow endpoints should use API-key scopes such as `approval_requests:create` and `acknowledgements:create`.

Human endpoints should use Supabase auth middleware and then application-level workspace authorization.

These auth paths should stay separate. A workflow credential should not act like a human user, and a human session should not bypass workflow idempotency or integration rules.

## Error Responses

Handlers should return one public error shape:

```json
{
  "error": {
    "code": "approval_request_expired",
    "message": "This approval request has expired.",
    "request_id": "req_123"
  }
}
```

Handlers may log internal errors with safe structured metadata, but public responses must not expose SQL errors, stack traces, secrets, continuation URLs, JWTs, API keys, or sensitive action payloads.

## Service Boundary

The preferred MVP shape is feature-specific services:

```text
internal/approval
internal/policy
internal/identity/apikeys
internal/identity/workspaces
internal/delivery
```

These packages own application use cases for their capability. Cross-domain orchestration should stay in the use case that owns the business event. For example, resolving an approval request may create an approval decision, audit event, delivery record, and River job in one transaction.

Do not introduce a broad `internal/app` package until there is repeated cross-domain orchestration that cannot be cleanly owned by a feature service.

## PostgreSQL Boundary

The current MVP code uses `pgxpool` and sqlc-generated queries directly from some services. That is acceptable only for application-service code that is already infrastructure-aware.

The target shape for complex behavior is:

```text
internal/approval
  -> repository interfaces owned by approval use cases

internal/platform/postgres
  -> sqlc-backed repository implementations

db/generated
  -> generated query code only
```

Move toward repositories when:

- transaction logic becomes hard to read inside a service;
- tests need to exercise domain behavior without PostgreSQL setup;
- multiple storage implementations exist;
- a use case needs only a narrow persistence behavior;
- SQL details are leaking into domain concepts.

Do not create repository interfaces only for mocking. Define them when they express a real boundary.

## Transaction Ownership

Application services own transaction boundaries. Important state transitions must be atomic.

Approval resolution should happen in one transaction:

```text
1. Lock pending approval request by workspace and request id.
2. Validate status, expiry, and approver authorization.
3. Insert decision.
4. Update approval request status.
5. Insert audit event.
6. Insert delivery record or transaction-safe River job.
7. Commit.
```

Handlers must not begin transactions.

## Async Work

The API may schedule asynchronous work, but it must not execute it inline.

Use cases should schedule durable jobs through a narrow job client or delivery service. The worker process should execute River jobs and reload current database state before side effects.

Initial async boundaries:

- decision delivery;
- Slack notification;
- email notification;
- approval expiry;
- escalation;
- cleanup of sensitive continuation data.

Job arguments should contain stable identifiers, not full mutable domain objects.

## Router Composition

For now, `httpapi.NewRouter` may construct handlers and simple services.

Extract a dedicated composition package when wiring becomes noisy:

```text
apps/api/internal/runtime
  ├── container.go
  └── server.go
```

That package would assemble repositories, services, job clients, signers, notifier clients, and handlers. `main.go` should remain small and process-focused.

## Versioning

The public API base path is:

```text
/api/v1
```

Route handlers should preserve backwards-compatible response shapes within a version. Breaking contract changes require a new version or an explicit migration plan.

## Security Rules

Every API change must preserve these rules:

- workspace scope is derived from authenticated identity or verified membership;
- tenant-owned resources are never loaded by ID alone;
- API keys are stored only as hashes and shown only once;
- continuation URLs are never returned to browser clients or logs;
- approval handlers never mutate immutable original action data;
- expired requests cannot be approved or rejected;
- duplicate submissions with the same workspace-scoped idempotency key do not create duplicate requests or jobs;
- the API never executes the customer business action itself.

## Near-Term Target

The next iteration of `apps/api` should aim for this shape:

```text
apps/api/main.go
  process lifecycle only

apps/api/internal/httpapi
  router, middleware, handlers, response helpers

internal/{approval,policy,identity,delivery}
  application services and domain rules

internal/platform/{postgres,river,webhook,signing,notification}
  infrastructure implementations

db/queries and db/generated
  explicit SQL and generated access code
```

This keeps the API process thin while allowing the product behavior to deepen behind stable application use cases.
