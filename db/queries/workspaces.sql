-- name: CreateWorkspace :one
insert into public.workspaces (
    name,
    slug,
    default_policy_effect,
    created_by
) values (
    $1,
    $2,
    $3,
    $4
)
returning *;

-- name: GetWorkspaceByID :one
select *
from public.workspaces
where id = $1;

-- name: GetWorkspaceBySlug :one
select *
from public.workspaces
where slug = $1;

-- name: GetWorkspaceMember :one
select *
from public.workspace_members
where workspace_id = $1
  and user_id = $2;

-- name: ListWorkspacesByUserID :many
select
    w.*,
    wm.role
from public.workspaces w
join public.workspace_members wm
  on wm.workspace_id = w.id
where wm.user_id = $1
order by w.created_at asc;

-- name: CreateWorkspaceMember :one
insert into public.workspace_members (
    workspace_id,
    user_id,
    role
) values (
    $1,
    $2,
    $3
)
returning *;
