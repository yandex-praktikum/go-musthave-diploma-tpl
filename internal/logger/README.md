# logger

Global zap logger setup.

- `Logger` — package-level logger, no-op until `InitLogger` is called.
- `InitLogger(level)` — builds a production zap logger at the given level.
