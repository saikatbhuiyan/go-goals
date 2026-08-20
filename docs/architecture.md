# Architecture

This repository is being shaped as a production-grade 3-tier modular monolith.
The tiers are intentionally separated so the Go application can later split into
microservices without forcing an early distributed-system rewrite.

## Tiers

1. UI tier: `apps/web`
   - Next.js App Router frontend.
   - Owns browser rendering, route-level UX, forms, and frontend validation.
   - Talks to the backend through HTTP APIs.

2. Application tier: `apps/api`
   - Go HTTP API and server-side application logic.
   - Organized by modules/domains, not by technical file type alone.
   - Owns authentication, authorization, business rules, and integration with the data tier.

3. Data tier: PostgreSQL
   - Local development is managed by `docker-compose.yml`.
   - Schema changes live in `migrations/`.
   - The application tier is the only tier that should talk directly to PostgreSQL.

## Backend Module Boundaries

The Go backend should evolve toward this structure:

```text
apps/api/
  cmd/server/          # executable entrypoint
  internal/platform/   # shared server, config, database, sessions
  internal/modules/
    auth/              # signup, signin, password hashing, sessions
    users/             # profiles, follows, administrators
    updates/           # aspiration updates, likes, comments
    pages/             # static/document pages while templates remain in Go
```

Modules may share platform primitives, but they should not reach into each
other's private implementation details. When a module needs data owned by
another module, introduce a small interface at the consuming boundary.

## Microservice Readiness

A module is ready to become a microservice when:

- Its data ownership is clear.
- Its public operations are already expressed as handlers/services.
- Cross-module calls happen through interfaces or HTTP/RPC clients.
- Migrations for its tables are isolated enough to move.
- It can be tested without booting the full application.

Good first service candidates, in order:

1. Auth and identity.
2. Notification delivery.
3. Feed/updates.
4. Payments or other external integrations if they are added later.

## Frontend Stack

The Next.js UI should use:

- TypeScript.
- App Router.
- ESLint.
- Tailwind CSS.
- Typed environment variables.
- A generated or hand-maintained API client around the Go HTTP API.

The first migration keeps the existing Go templates running while `apps/web`
is introduced. After feature parity exists in Next.js, Go template rendering can
be removed and the Go tier can become API-only.
