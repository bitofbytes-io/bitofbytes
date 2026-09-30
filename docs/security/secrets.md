# Container runtime configuration

The runtime image expects configuration to be provided entirely by the
orchestrator at container start, through environment variables. The application
needs no secrets: CSRF protection checks the browser's `Sec-Fetch-Site` and
`Origin` headers (via `filippo.io/csrf/gorilla`) instead of signed tokens, so
there is no CSRF key to provision or rotate. Missing variables cause the
application to exit with a clear error message, preventing an unexpectedly
insecure default.

## Local development

Local developers should continue using a `.env` file directly with the Go
application (for example via `go run ./cmd/bob`). Copy `.env.template` to `.env`
and provide explicit values for each variable before launching the app. The
container image does not read a bind-mounted `.env` file.

## Required configuration

The application requires the following environment variables at runtime:

* `SERVER_ADDRESS`
* `CSRF_SECURE` (`true` behind production HTTPS, `false` for local HTTP); it
  controls the secure response headers

Swarm services should set these through `--env` flags or their compose
equivalents. The Dockerfile defaults `SERVER_ADDRESS=:3000` and
`CSRF_SECURE=true`.

## Upgrading from the CSRF key

Earlier versions required a `csrf_key` Docker Swarm secret (or `CSRF_KEY` /
`CSRF_KEY_FILE`). Those values are now ignored. Existing deployments keep working
with the secret still mounted; remove it from the stack and run
`docker secret rm csrf_key` whenever convenient.
