-- name: CreateCreativeProject :one
insert into creative_projects (owner_user_id, title, description, start_year, start_month, lyrics, creation_method)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: GetCreativeProject :one
select * from creative_projects where id = $1;

-- name: ListCreativeProjectsByOwner :many
select * from creative_projects
where owner_user_id = $1
order by start_year desc, start_month desc, created_at desc;

-- name: UpdateCreativeProject :one
update creative_projects set
  title = $2,
  description = $3,
  start_year = $4,
  start_month = $5,
  lyrics = $6,
  creation_method = $7,
  updated_at = now()
where id = $1
returning *;

-- name: DeleteCreativeProject :exec
delete from creative_projects where id = $1;
