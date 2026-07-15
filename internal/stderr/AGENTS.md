# internal/stderr

Central error presentation and exit behavior.

- Keep stderr output consistent and separate from scriptable stdout.
- Log errors before presenting them.
- Preserve configurable exit behavior used by tests.
- Do not introduce business decisions here.

Any output wording change is user-visible: test failure and non-exiting test paths and inspect affected snapshots.
