-- name: CreateApprovalRequest :one
insert into public.approval_requests (
    workspace_id,
    idempotency_key,
    action_type,
    title,
    description,
    original_action,
    original_action_hash,
    source_platform,
    source_workflow_id,
    source_execution_id,
    context,
    affected_systems,
    metadata,
    matched_policy_id,
    matched_policy_version_id,
    matched_policy_snapshot,
    assigned_user_id,
    assigned_group_id,
    status,
    decision_required,
    expires_at,
    resolved_at
) values (
    sqlc.arg(workspace_id),
    sqlc.arg(idempotency_key),
    sqlc.arg(action_type),
    sqlc.arg(title),
    sqlc.arg(description),
    sqlc.arg(original_action)::text::jsonb,
    sqlc.arg(original_action_hash),
    sqlc.arg(source_platform),
    sqlc.arg(source_workflow_id),
    sqlc.arg(source_execution_id),
    sqlc.arg(context)::text::jsonb,
    sqlc.arg(affected_systems)::text::jsonb,
    sqlc.arg(metadata)::text::jsonb,
    sqlc.arg(matched_policy_id),
    sqlc.arg(matched_policy_version_id),
    sqlc.arg(matched_policy_snapshot)::text::jsonb,
    sqlc.arg(assigned_user_id),
    sqlc.arg(assigned_group_id),
    sqlc.arg(status),
    sqlc.arg(decision_required),
    sqlc.arg(expires_at),
    sqlc.arg(resolved_at)
)
returning *;

-- name: GetApprovalRequestByID :one
select *
from public.approval_requests
where workspace_id = $1
  and id = $2;

-- name: GetApprovalRequestByIdempotencyKey :one
select *
from public.approval_requests
where workspace_id = $1
  and idempotency_key = $2;

-- name: ListApprovalRequestsByWorkspace :many
select *
from public.approval_requests
where workspace_id = $1
order by created_at desc
limit $2
offset $3;

-- name: ListPendingApprovalRequestsByAssignee :many
select *
from public.approval_requests
where workspace_id = $1
  and assigned_user_id = $2
  and status = 'pending'
order by created_at desc
limit $3
offset $4;

-- name: LockApprovalRequestForDecision :one
select *
from public.approval_requests
where workspace_id = $1
  and id = $2
for update;

-- name: ResolveApprovalRequest :one
update public.approval_requests
set status = $3,
    resolved_at = $4,
    updated_at = $4
where workspace_id = $1
  and id = $2
returning *;

-- name: MarkApprovalRequestEvaluating :one
update public.approval_requests
set status = 'evaluating',
    updated_at = $3
where workspace_id = $1
  and id = $2
returning *;

-- name: UpdateApprovalRequestAssignment :one
update public.approval_requests
set assigned_user_id = $3,
    assigned_group_id = $4,
    expires_at = $5,
    updated_at = $6
where workspace_id = $1
  and id = $2
returning *;
