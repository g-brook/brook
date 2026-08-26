## What changed?

Describe the user-visible behavior and why this change is needed. Link the Issue or Discussion that established the scope.

## How was it verified?

List exact commands, environments, and relevant output. Include failure-path testing for networking or protocol changes.

## Compatibility

- [ ] No protocol, configuration, database, or CLI compatibility impact.
- [ ] Backward-compatible change; the compatibility path is tested.
- [ ] Breaking change; migration and rollback instructions are included.

## Security and operations

- [ ] Authentication, authorization, secret handling, and network exposure were considered.
- [ ] Resource limits, timeouts, reconnect behavior, and shutdown behavior were considered.
- [ ] Logs and examples contain no credentials, tokens, private keys, public IP addresses, or personal data.
- [ ] Documentation and release notes describe any new operational requirement.

## Checklist

- [ ] Tests cover the change and relevant failure paths.
- [ ] English and Chinese documentation remain consistent where applicable.
- [ ] Generated files and binaries are not included unless the build process requires them.
- [ ] The change is focused and does not contain unrelated formatting or refactoring.
