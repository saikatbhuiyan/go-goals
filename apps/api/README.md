# API Architecture

The API tier follows a DDD-friendly modular monolith layout.

```text
cmd/server
  process entrypoint

internal/domain
  shared enterprise concepts and domain errors

internal/app/server
  HTTP application composition, route registration, middleware, and startup

internal/modules
  auth      signup, signin, password verification, session establishment
  users     profiles, follows, administrator checks, moderation
  updates   aspiration updates, likes, comments, feeds
  pages     legacy server-rendered pages while Next.js takes over the UI

internal/platform
  config    environment configuration
  httpx     transport-neutral HTTP helpers
  postgres  database connection and adapter helpers
  sessions  cookie session storage adapters

migrations
  PostgreSQL schema migrations owned by the API tier

templates
  legacy Go-rendered pages during the UI migration

static
  legacy Go-rendered assets during the UI migration
```

Rules:

- Domain packages do not import transport or infrastructure packages.
- Modules expose small services and HTTP adapters rather than sharing global state.
- Platform packages know about frameworks and infrastructure, but not business workflows.
- `internal/app/server` composes modules and platform adapters.
