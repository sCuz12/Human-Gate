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
