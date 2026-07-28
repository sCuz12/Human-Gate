-- name: CreateContinuationTarget :one
insert into public.continuation_targets (
    workspace_id,
    approval_request_id,
    strategy,
    platform,
    destination,
    encrypted_configuration
) values (
    sqlc.arg(workspace_id),
    sqlc.arg(approval_request_id),
    sqlc.arg(strategy),
    sqlc.arg(platform),
    sqlc.arg(destination),
    sqlc.arg(encrypted_configuration)::text::jsonb
)
returning *;

-- name: GetContinuationTargetByRequestID :one
select *
from public.continuation_targets
where workspace_id = $1
  and approval_request_id = $2;
