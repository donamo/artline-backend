-- +goose Up
create table users (
  id uuid primary key default gen_random_uuid(),
  google_subject text not null unique,
  email text not null,
  display_name text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table creative_projects (
  id uuid primary key default gen_random_uuid(),
  owner_user_id uuid not null references users(id) on delete cascade,
  title text not null,
  description text,
  start_year int not null,
  start_month int not null check (start_month >= 1 and start_month <= 12),
  lyrics text,
  creation_method text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table creative_project_links (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references creative_projects(id) on delete cascade,
  platform text not null,
  url text not null,
  label text,
  sort_order int not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_creative_projects_owner_date
  on creative_projects(owner_user_id, start_year desc, start_month desc);

create index idx_creative_project_links_project
  on creative_project_links(project_id);

-- +goose Down
drop table if exists creative_project_links;
drop table if exists creative_projects;
drop table if exists users;
