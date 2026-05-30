# Container runtime configuration

The runtime image expects configuration to be provided entirely by the
orchestrator at container start. The application now reads the CSRF key directly
from the Docker Swarm secret mounted at `/run/secrets/csrf_key` (or the path
specified by `CSRF_KEY_FILE`). Non-secret configuration must be supplied via
environment variables so the container can be configured completely through
Swarm managed inputs. Missing variables cause the application to exit with a
clear error message, preventing an unexpectedly insecure default.

## Local development

Local developers should continue using a `.env` file directly with the Go
application (for example via `go run ./cmd/bob`). Copy `.env.template` to `.env`
and provide explicit values for each variable before launching the app. The
container image relies exclusively on secrets managed by Docker Swarm and does
not read a bind-mounted `.env` file.

## Production secrets with Docker Swarm

Deployments must supply a Docker Swarm secret that contains the 32-byte CSRF
key. The application reads the secret file at startup and fails fast if the
value cannot be read. Example service definition:

```sh
docker service create \
  --secret source=csrf_key,target=csrf_key \
  registry.tail209cfc.ts.net/bob:latest
```

### Creating the CSRF secret

Generate a strong CSRF key and store it in Swarm (refer to the [official Docker
documentation](https://docs.docker.com/engine/swarm/secrets/) for additional
background):

```sh
openssl rand -base64 32 | docker secret create csrf_key -
```

### Updating the CSRF secret

Docker secrets are immutable, so rotating the CSRF key requires creating a new
secret, updating the service to consume it, and then removing the old value:

```sh
openssl rand -base64 32 > csrf.key
docker secret create csrf_key_v2 csrf.key
docker service update \
  --secret-add source=csrf_key_v2,target=csrf_key \
  your_service_name
docker service update --secret-rm csrf_key your_service_name
docker secret rm csrf_key
rm csrf.key
```

Wait for the service to report a healthy state before removing the original
secret. Using the same `target` path keeps the in-container location stable, so
no code changes are required during rotations.

## Additional configuration

The application requires the following environment variables at runtime:

* `SERVER_ADDRESS`
* `CSRF_SECURE`
* either `CSRF_KEY` or a readable secret file referenced by `CSRF_KEY_FILE`

Swarm services should set `SERVER_ADDRESS` and `CSRF_SECURE` through `--env`
flags (or their compose equivalents) alongside the mounted `csrf_key` secret.
The Makefile, Dockerfile, and GitHub Actions build workflow do not need to
generate a `.env` file because the service reads its configuration straight from
the runtime environment and exits cleanly when required values are absent.

## Build verification

The application still builds cleanly after these changes. Running `go test
./...` against the module (using Go 1.22.x in CI) succeeds without requiring a
`.env` file because the configuration loader only attempts to read one when it
is present on disk. This matches both the local development and container
runtime expectations.
