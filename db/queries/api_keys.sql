-- name: CreateAPIKey :one
insert into public.api_keys (
    workspace_id,
    name,
    key_prefix,
    key_hash,
    scopes,
    created_by
) values (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
returning *;

-- name: GetActiveAPIKeyByPrefix :one
select *
from public.api_keys
where workspace_id = $1
  and key_prefix = $2
  and revoked_at is null;

-- name: GetActiveAPIKeyByPrefixGlobal :many
select *
from public.api_keys
where key_prefix = $1
  and revoked_at is null;

-- name: ListAPIKeysByWorkspace :many
select *
from public.api_keys
where workspace_id = $1
order by created_at desc;

-- name: UpdateAPIKeyLastUsedAt :exec
update public.api_keys
set last_used_at = $3
where workspace_id = $1
  and id = $2;

-- name: RevokeAPIKey :exec
update public.api_keys
set revoked_at = $3
where workspace_id = $1
  and id = $2
  and revoked_at is null;
