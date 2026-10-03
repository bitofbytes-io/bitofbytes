# HTTP Server Timeout Defaults

The HTTP server timeouts configured in `cmd/bob/bob.go` follow the guidance in
[Cloudflare's "The complete guide to Go net/http timeouts"](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/).

The article recommends the following production-friendly defaults:

| Timeout              | Value          | Rationale |
| -------------------- | -------------- | --------- |
| `ReadTimeout`        | 5 seconds      | Limits how long a client can take to finish sending the request body, mitigating slowloris-style attacks. |
| `ReadHeaderTimeout`  | 2 seconds      | Prevents clients from stalling when sending HTTP headers. |
| `WriteTimeout`       | 10 seconds     | Bounds the amount of time we spend writing the response, protecting server resources when clients read slowly. |
| `IdleTimeout`        | 120 seconds    | Keeps idle keep-alive connections from lingering forever while still being generous for typical browsers. |

On SIGINT or SIGTERM the server stops accepting connections and gives in-flight
requests up to 5 seconds (`shutdownTimeout`) to finish, inside Docker's default
10-second stop grace period.

These values provide hardened defaults while remaining practical for the
application's use case. Adjust them if future functional requirements demand
more permissive thresholds. 
