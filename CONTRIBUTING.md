# Contributing to g-brook/brook

Thank you for helping improve Brook. Small, reproducible, well-tested changes are easier to review and safer to ship in a network-facing project.

## Before opening an Issue

- Search existing Issues and Discussions.
- Use Discussions for deployment help and open-ended design questions.
- Use the structured Issue forms for reproducible bugs, feature use cases, and documentation problems.
- Never publish credentials, tunnel tokens, Web sessions, certificate private keys, production IP addresses, or personal data.
- Do not open a public Issue for a suspected vulnerability. Follow `SECURITY.md` after the maintainer publishes it and enables a private reporting route.

## Before writing code

For a small bug fix, open or link a focused Issue. For protocol changes, new tunnel types, database migrations, public API changes, or substantial UI work, start a Discussion and agree on compatibility and rollback before implementation.

Avoid large unsolicited pull requests. A short design note can prevent duplicated effort and incompatible implementations.

## Development areas

The repository uses a Go workspace with separate modules for the common library, server, client, and command packages. The Web management UI lives under `portal/server`.

Use the build instructions in the root README as the current source of truth. When a change touches multiple modules, test every affected module rather than only the package where the edit was made.

## Testing expectations

- Add focused unit tests for new parsing, routing, configuration, and authentication behavior.
- Add integration coverage for control/data-plane changes and client/server compatibility.
- Exercise timeouts, reconnects, half-closed connections, server restarts, invalid frames, and resource exhaustion where relevant.
- For performance claims, include a reproducible benchmark, workload, hardware description, comparison method, and raw results.
- Do not weaken validation or skip a failing test to make a change pass.

## Protocol and configuration compatibility

Document framing or command changes, version negotiation, mixed-version behavior, migration, and rollback. New configuration fields should have safe defaults and preserve existing deployments whenever possible.

## Pull requests

Keep each pull request focused. Explain the user problem, implementation, compatibility impact, security impact, and exact verification steps. Update both English and Chinese documentation when the same behavior is described in both places.

Maintainers may ask for a smaller change, additional failure tests, or a design Discussion before review continues.
