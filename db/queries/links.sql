-- name: CreateProjectLink :one
insert into creative_project_links (project_id, platform, url, label, sort_order)
values ($1, $2, $3, $4, $5)
returning *;

-- name: ListProjectLinks :many
select * from creative_project_links
where project_id = $1
order by sort_order asc, created_at asc;

-- name: DeleteProjectLinks :exec
delete from creative_project_links where project_id = $1;
