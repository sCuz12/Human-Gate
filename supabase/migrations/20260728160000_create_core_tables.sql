create type workspace_role as enum (
    'owner',
    'administrator',
    'approver',
    'viewer'
);

create type policy_effect as enum (
    'allow',
    'require_approval',
    'block'
);

create type approval_status as enum (
    'received',
    'evaluating',
    'allowed',
    'blocked',
    'pending',
    'approved',
    'approved_with_changes',
    'rejected',
    'expired',
    'cancelled'
);

create type decision_type as enum (
    'allowed',
    'approved',
    'approved_with_changes',
    'rejected',
    'blocked',
    'expired',
    'cancelled'
);

create type continuation_strategy as enum (
    'webhook',
    'polling'
);

create type delivery_status as enum (
    'not_required',
    'pending',
    'attempting',
    'delivered',
    'acknowledged',
    'retrying',
    'permanently_failed'
);

create type acknowledgement_status as enum (
    'decision_received',
    'workflow_resumed',
    'workflow_completed',
    'workflow_failed'
);

create table public.users (
    id uuid primary key references auth.users(id) on delete cascade,
    email text not null,
    display_name text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table public.workspaces (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    slug text not null,
    default_policy_effect policy_effect not null default 'require_approval',
    created_by uuid references public.users(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (slug)
);

create table public.workspace_members (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    user_id uuid not null references public.users(id) on delete cascade,
    role workspace_role not null,
    created_at timestamptz not null default now(),
    unique (workspace_id, user_id)
);

create table public.api_keys (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    name text not null,
    key_prefix text not null,
    key_hash text not null,
    scopes text[] not null default '{}',
    last_used_at timestamptz,
    revoked_at timestamptz,
    created_by uuid references public.users(id),
    created_at timestamptz not null default now(),
    unique (workspace_id, key_prefix)
);

create table public.approver_groups (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    name text not null,
    description text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (workspace_id, name)
);

create table public.approver_group_members (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    approver_group_id uuid not null references public.approver_groups(id) on delete cascade,
    user_id uuid not null references public.users(id) on delete cascade,
    created_at timestamptz not null default now(),
    unique (approver_group_id, user_id)
);

create table public.policies (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    name text not null,
    description text,
    priority integer not null default 100,
    is_active boolean not null default false,
    created_by uuid references public.users(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table public.policy_versions (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    policy_id uuid not null references public.policies(id) on delete cascade,
    version_number integer not null,
    conditions jsonb not null default '[]'::jsonb,
    effect policy_effect not null,
    approval_settings jsonb not null default '{}'::jsonb,
    created_by uuid references public.users(id),
    created_at timestamptz not null default now(),
    unique (policy_id, version_number)
);

create table public.approval_requests (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    idempotency_key text not null,
    action_type text not null,
    title text not null,
    description text,
    original_action jsonb not null,
    original_action_hash text not null,
    source_platform text not null,
    source_workflow_id text not null,
    source_execution_id text not null,
    context jsonb not null default '{}'::jsonb,
    affected_systems jsonb not null default '[]'::jsonb,
    metadata jsonb not null default '{}'::jsonb,
    matched_policy_id uuid references public.policies(id),
    matched_policy_version_id uuid references public.policy_versions(id),
    matched_policy_snapshot jsonb,
    assigned_user_id uuid references public.users(id),
    assigned_group_id uuid references public.approver_groups(id),
    status approval_status not null,
    decision_required boolean not null default false,
    expires_at timestamptz,
    resolved_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (workspace_id, idempotency_key),
    check (
        (status = 'pending' and resolved_at is null)
        or (status <> 'pending' and resolved_at is not null)
        or status in ('received', 'evaluating')
    )
);

create table public.approval_decisions (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    approval_request_id uuid not null references public.approval_requests(id) on delete cascade,
    decision decision_type not null,
    original_action_hash text not null,
    approved_action jsonb,
    approved_action_hash text,
    changed_fields jsonb not null default '[]'::jsonb,
    comment text,
    decided_by uuid references public.users(id),
    issued_at timestamptz not null default now(),
    expires_at timestamptz,
    created_at timestamptz not null default now(),
    unique (approval_request_id)
);

create table public.approval_comments (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    approval_request_id uuid not null references public.approval_requests(id) on delete cascade,
    author_id uuid not null references public.users(id),
    body text not null,
    created_at timestamptz not null default now()
);

create table public.continuation_targets (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    approval_request_id uuid not null references public.approval_requests(id) on delete cascade,
    strategy continuation_strategy not null,
    platform text not null,
    destination text,
    encrypted_configuration jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    unique (approval_request_id)
);

create table public.decision_deliveries (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    decision_id uuid not null references public.approval_decisions(id) on delete cascade,
    continuation_target_id uuid not null references public.continuation_targets(id) on delete cascade,
    status delivery_status not null,
    attempt_count integer not null default 0,
    next_attempt_at timestamptz,
    last_attempt_at timestamptz,
    last_response_code integer,
    last_error text,
    delivered_at timestamptz,
    acknowledged_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (decision_id)
);

create table public.workflow_acknowledgements (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    approval_request_id uuid not null references public.approval_requests(id) on delete cascade,
    decision_id uuid references public.approval_decisions(id) on delete set null,
    status acknowledgement_status not null,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create table public.audit_events (
    id uuid primary key default gen_random_uuid(),
    workspace_id uuid not null references public.workspaces(id) on delete cascade,
    approval_request_id uuid references public.approval_requests(id) on delete cascade,
    decision_id uuid references public.approval_decisions(id) on delete cascade,
    actor_type text not null,
    actor_id uuid,
    event_type text not null,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index idx_workspace_members_workspace_id
    on public.workspace_members (workspace_id);

create index idx_api_keys_workspace_id
    on public.api_keys (workspace_id);

create index idx_approver_groups_workspace_id
    on public.approver_groups (workspace_id);

create index idx_approver_group_members_workspace_id_group_id
    on public.approver_group_members (workspace_id, approver_group_id);

create index idx_policies_workspace_id_priority
    on public.policies (workspace_id, priority);

create index idx_policy_versions_workspace_id_policy_id
    on public.policy_versions (workspace_id, policy_id);

create index idx_approval_requests_workspace_id_status_created_at
    on public.approval_requests (workspace_id, status, created_at desc);

create index idx_approval_requests_workspace_id_assigned_user_id_status
    on public.approval_requests (workspace_id, assigned_user_id, status);

create index idx_approval_requests_workspace_id_source
    on public.approval_requests (workspace_id, source_platform, source_workflow_id);

create index idx_approval_decisions_workspace_id
    on public.approval_decisions (workspace_id);

create index idx_approval_comments_workspace_id_request_id
    on public.approval_comments (workspace_id, approval_request_id, created_at);

create index idx_continuation_targets_workspace_id
    on public.continuation_targets (workspace_id);

create index idx_decision_deliveries_workspace_id_status_next_attempt_at
    on public.decision_deliveries (workspace_id, status, next_attempt_at);

create index idx_workflow_acknowledgements_workspace_id_request_id
    on public.workflow_acknowledgements (workspace_id, approval_request_id, created_at);

create index idx_audit_events_workspace_id_request_id_created_at
    on public.audit_events (workspace_id, approval_request_id, created_at desc);
