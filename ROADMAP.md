# Brook Roadmap

This roadmap describes priorities, not promises or release dates. Work should be linked to an Issue or Discussion with acceptance criteria and a maintainer.

## Now: production trust baseline

- [ ] Add authenticated encryption for the manager and smux data connections.
- [ ] Bind tunnel streams to an authenticated client session.
- [ ] Replace legacy Web password hashing and use cryptographically secure session tokens.
- [ ] Add a configurable management bind address and a documented TLS deployment path.
- [ ] Publish a threat model, production hardening guide, and `SECURITY.md` with private reporting.
- [ ] Add integration tests for authentication failures, reconnects, stale workers, and server restarts.

## Next: reproducibility and operations

- [ ] Publish a versioned configuration reference with validation rules and examples.
- [ ] Document metrics, logs, health checks, backup, restore, upgrade, and rollback.
- [ ] Build reproducible throughput and latency benchmarks with raw results.
- [ ] Add tested container deployment guidance without making containers the only supported path.
- [ ] Define the compatibility policy for clients, servers, configuration, and the tunnel frame format.

## Later: evaluated extensions

- [ ] Evaluate QUIC only after an end-to-end implementation, compatibility plan, and failure tests exist.
- [ ] Evaluate additional authentication providers against the published threat model.
- [ ] Improve multi-instance operations after baseline security and observability are complete.

## How to help

Start with an Issue labeled `good first issue` or `help wanted`, or open a Discussion with a reproducible use case. Security details belong in the private reporting channel, not a public Issue.
