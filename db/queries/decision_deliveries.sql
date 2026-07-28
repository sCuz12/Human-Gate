-- name: CreateDecisionDelivery :one
insert into public.decision_deliveries (
    workspace_id,
    decision_id,
    continuation_target_id,
    status,
    next_attempt_at
) values (
    $1,
    $2,
    $3,
    $4,
    $5
)
returning *;

-- name: GetDecisionDeliveryByDecisionID :one
select *
from public.decision_deliveries
where workspace_id = $1
  and decision_id = $2;

-- name: GetDecisionDeliveryByApprovalRequestID :one
select
    dd.id,
    dd.workspace_id,
    dd.decision_id,
    dd.continuation_target_id,
    dd.status,
    dd.attempt_count,
    dd.next_attempt_at,
    dd.last_attempt_at,
    dd.last_response_code,
    dd.last_error,
    dd.delivered_at,
    dd.acknowledged_at,
    dd.created_at,
    dd.updated_at
from public.decision_deliveries dd
join public.approval_decisions ad
  on ad.id = dd.decision_id
 and ad.workspace_id = dd.workspace_id
where dd.workspace_id = $1
  and ad.approval_request_id = $2;

-- name: ListDueDecisionDeliveries :many
select
    dd.id as delivery_id,
    dd.workspace_id,
    dd.decision_id,
    dd.continuation_target_id,
    dd.status,
    dd.attempt_count,
    dd.next_attempt_at,
    ad.approval_request_id,
    ad.decision,
    ad.original_action_hash,
    ad.approved_action,
    ad.approved_action_hash,
    ad.issued_at,
    ar.original_action,
    ar.action_type,
    ct.strategy,
    ct.platform,
    ct.destination
from public.decision_deliveries dd
join public.approval_decisions ad
  on ad.id = dd.decision_id
 and ad.workspace_id = dd.workspace_id
join public.approval_requests ar
  on ar.id = ad.approval_request_id
 and ar.workspace_id = dd.workspace_id
join public.continuation_targets ct
  on ct.id = dd.continuation_target_id
 and ct.workspace_id = dd.workspace_id
where dd.status in ('pending', 'retrying')
  and (dd.next_attempt_at is null or dd.next_attempt_at <= sqlc.arg(now_at))
order by coalesce(dd.next_attempt_at, dd.created_at), dd.created_at
limit sqlc.arg(limit_count);

-- name: UpdateDecisionDeliveryAttempt :one
update public.decision_deliveries
set status = $3,
    attempt_count = $4,
    next_attempt_at = $5,
    last_attempt_at = $6,
    last_response_code = $7,
    last_error = $8,
    delivered_at = $9,
    acknowledged_at = $10,
    updated_at = $11
where workspace_id = $1
  and id = $2
returning *;
