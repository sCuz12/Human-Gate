create or replace function public.handle_auth_user_sync()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
    insert into public.users (
        id,
        email,
        display_name,
        created_at,
        updated_at
    ) values (
        new.id,
        coalesce(new.email, ''),
        coalesce(
            new.raw_user_meta_data ->> 'display_name',
            new.raw_user_meta_data ->> 'full_name',
            new.raw_user_meta_data ->> 'name'
        ),
        coalesce(new.created_at, now()),
        now()
    )
    on conflict (id) do update
    set email = excluded.email,
        display_name = excluded.display_name,
        updated_at = now();

    return new;
end;
$$;

drop trigger if exists on_auth_user_created_or_updated on auth.users;

create trigger on_auth_user_created_or_updated
after insert or update on auth.users
for each row
execute function public.handle_auth_user_sync();

insert into public.users (
    id,
    email,
    display_name,
    created_at,
    updated_at
)
select
    au.id,
    coalesce(au.email, ''),
    coalesce(
        au.raw_user_meta_data ->> 'display_name',
        au.raw_user_meta_data ->> 'full_name',
        au.raw_user_meta_data ->> 'name'
    ),
    coalesce(au.created_at, now()),
    now()
from auth.users au
on conflict (id) do update
set email = excluded.email,
    display_name = excluded.display_name,
    updated_at = now();
