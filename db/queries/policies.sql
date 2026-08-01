-- name: CreatePolicy :one
insert into public.policies (
    workspace_id,
    name,
    description,
    priority,
    is_active,
    created_by
) values (
    sqlc.arg(workspace_id),
    sqlc.arg(name),
    sqlc.arg(description),
    sqlc.arg(priority),
    sqlc.arg(is_active),
    sqlc.arg(created_by)
)
returning *;

-- name: CreatePolicyVersion :one
insert into public.policy_versions (
    workspace_id,
    policy_id,
    version_number,
    conditions,
    effect,
    approval_settings,
    created_by
) values (
    sqlc.arg(workspace_id),
    sqlc.arg(policy_id),
    sqlc.arg(version_number),
    sqlc.arg(conditions)::text::jsonb,
    sqlc.arg(effect),
    sqlc.arg(approval_settings)::text::jsonb,
    sqlc.arg(created_by)
)
returning *;

-- name: ListPolicySummariesForWorkspace :many
select
    p.id as policy_id,
    p.workspace_id,
    p.name,
    p.description,
    p.priority,
    p.is_active,
    p.created_at,
    p.updated_at,
    pv.id as policy_version_id,
    pv.version_number,
    pv.conditions,
    pv.effect,
    pv.approval_settings,
    pv.created_at as version_created_at
from public.policies p
join public.policy_versions pv
  on pv.policy_id = p.id
 and pv.workspace_id = p.workspace_id
where p.workspace_id = $1
  and pv.version_number = (
      select max(pv2.version_number)
      from public.policy_versions pv2
      where pv2.policy_id = p.id
        and pv2.workspace_id = p.workspace_id
  )
order by p.priority asc, p.created_at asc;

-- name: ListActivePolicyVersionsForWorkspace :many
select
    p.id as policy_id,
    p.workspace_id,
    p.name,
    p.description,
    p.priority,
    pv.id as policy_version_id,
    pv.version_number,
    pv.conditions,
    pv.effect,
    pv.approval_settings,
    pv.created_at as version_created_at
from public.policies p
join public.policy_versions pv
  on pv.policy_id = p.id
 and pv.workspace_id = p.workspace_id
where p.workspace_id = $1
  and p.is_active = true
  and pv.version_number = (
      select max(pv2.version_number)
      from public.policy_versions pv2
      where pv2.policy_id = p.id
        and pv2.workspace_id = p.workspace_id
  )
order by p.priority asc, p.created_at asc;
