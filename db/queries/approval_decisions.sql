-- name: CreateApprovalDecision :one
insert into public.approval_decisions (
    workspace_id,
    approval_request_id,
    decision,
    original_action_hash,
    approved_action,
    approved_action_hash,
    changed_fields,
    comment,
    decided_by,
    issued_at,
    expires_at
) values (
    sqlc.arg(workspace_id),
    sqlc.arg(approval_request_id),
    sqlc.arg(decision),
    sqlc.arg(original_action_hash),
    sqlc.arg(approved_action)::text::jsonb,
    sqlc.arg(approved_action_hash),
    sqlc.arg(changed_fields)::text::jsonb,
    sqlc.arg(comment),
    sqlc.arg(decided_by),
    sqlc.arg(issued_at),
    sqlc.arg(expires_at)
)
returning *;

-- name: GetApprovalDecisionByRequestID :one
select *
from public.approval_decisions
where workspace_id = $1
  and approval_request_id = $2;
