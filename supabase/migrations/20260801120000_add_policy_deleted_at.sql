alter table public.policies
    add column deleted_at timestamptz;

create index idx_policies_workspace_id_deleted_at
    on public.policies (workspace_id, deleted_at);
