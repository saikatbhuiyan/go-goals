# API Architecture

The API tier follows a DDD-friendly modular monolith layout.

```text
cmd/server
  process entrypoint and final HTTP wiring

internal/domain
  shared enterprise concepts and domain errors

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
```

Rules:

- Domain packages do not import transport or infrastructure packages.
- Modules expose small services and HTTP adapters rather than sharing global state.
- Platform packages know about frameworks and infrastructure, but not business workflows.
- `cmd/server` composes modules and platform adapters.
