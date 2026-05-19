-- name: UpsertUser :one
insert into users (google_subject, email, display_name, updated_at)
values ($1, $2, $3, now())
on conflict (google_subject)
do update set
  email = excluded.email,
  display_name = excluded.display_name,
  updated_at = now()
returning *;

-- name: GetUserByID :one
select * from users where id = $1;

-- name: GetUserByGoogleSubject :one
select * from users where google_subject = $1;
