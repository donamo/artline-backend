# Artline Backend

Go backend az Alkotói útvonal alkalmazáshoz. Privát alkotói idővonal, ahol a felhasználó Google-fiókkal bejelentkezve saját zenei, videós és kreatív projektjeit rögzítheti.

## Stack

- **Go** + **chi** — HTTP router
- **gqlgen** — GraphQL (schema-first)
- **sqlc** — típusos SQL lekérdezések
- **goose** — DB migráció (embed FS, auto-futás induláskor)
- **PostgreSQL**
- **Google OpenID Connect** — authentikáció, httpOnly cookie session

## Könyvtárstruktúra

```
cmd/server/        – belépési pont (main.go)
internal/
  auth/            – Google OAuth, session middleware
  graph/           – gqlgen resolverek
  db/              – sqlc generált kód (ne szerkeszd kézzel)
db/
  migrations/      – goose SQL migrációk
  queries/         – sqlc SQL lekérdezések
```

## Helyi fejlesztés

### Előfeltételek

- Go 1.26+
- Docker (lokális Postgres fallbackhez) vagy elérhető Postgres szerver
- `sqlc` és `goose` CLI

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Indítás

```bash
cp .env.example .env
# töltsd ki a Google OAuth értékeket .env-ben

docker compose up -d   # ha lokális Postgres kell

go run ./cmd/server/
```

A szerver elindul és automatikusan lefuttatja a DB migrációkat.

### Környezeti változók

| Változó               | Leírás                                      |
|-----------------------|---------------------------------------------|
| `PORT`                | HTTP port (alapértelmezett: 8080)           |
| `DATABASE_URL`        | PostgreSQL connection string                |
| `TEST_DATABASE_URL`   | E2E tesztekhez használt DB                  |
| `GOOGLE_CLIENT_ID`    | Google OAuth client ID                      |
| `GOOGLE_CLIENT_SECRET`| Google OAuth client secret                  |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL                          |
| `SESSION_SECRET`      | httpOnly cookie session titkosítási kulcs   |

### Végpontok

| Metódus | Útvonal                   | Leírás                        |
|---------|---------------------------|-------------------------------|
| GET     | `/health`                 | Szerver állapot               |
| GET     | `/auth/google/login`      | Google login indítása         |
| GET     | `/auth/google/callback`   | Google OAuth callback         |
| POST    | `/auth/logout`            | Kijelentkezés                 |
| POST    | `/graphql`                | GraphQL endpoint              |

## GraphQL

A fő adatkezelés GraphQL-en keresztül történik. A séma a `internal/graph/schema.graphqls` fájlban van.

Főbb műveletek:

```graphql
query { me { id email displayName } }
query { myCreativeProjects { id title startYear startMonth links { platform url } } }
mutation { createCreativeProject(input: { ... }) { id } }
mutation { updateCreativeProject(id: "...", input: { ... }) { id } }
mutation { deleteCreativeProject(id: "...") }
```

## DB migráció

```bash
# új migráció létrehozása
goose -dir db/migrations create <név> sql

# manuális futtatás
goose -dir db/migrations postgres "$DATABASE_URL" up

# visszaállítás
goose -dir db/migrations postgres "$DATABASE_URL" down
```

## API leírók generálása (frontend számára)

```bash
make api
```

Kimenet:
- `api/schema.graphql` — GraphQL séma
- `api/openapi.yaml` — OpenAPI 3.0 spec a REST végpontokhoz (`/auth/*`, `/health`, `/graphql`)

## sqlc kód újragenerálása

Ha módosítod a `db/queries/` alatt lévő SQL fájlokat:

```bash
sqlc generate
```

## Tesztek

```bash
go test ./...
```

Az e2e tesztek a `TEST_DATABASE_URL` környezeti változót használják.
