-- name: CreateAuditEvent :one
insert into public.audit_events (
    workspace_id,
    approval_request_id,
    decision_id,
    actor_type,
    actor_id,
    event_type,
    metadata
) values (
    sqlc.arg(workspace_id),
    sqlc.arg(approval_request_id),
    sqlc.arg(decision_id),
    sqlc.arg(actor_type),
    sqlc.arg(actor_id),
    sqlc.arg(event_type),
    sqlc.arg(metadata)::text::jsonb
)
returning *;

-- name: ListAuditEventsByRequestID :many
select *
from public.audit_events
where workspace_id = $1
  and approval_request_id = $2
order by created_at asc;
