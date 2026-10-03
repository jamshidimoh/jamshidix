`Jamshidix.exe` is the Windows desktop client.

On normal launch it opens the native GUI, automatically requests Administrator privileges for TUN operation, refreshes the free public node directory, and lets the user filter by region or sort by priority, speed, region, or freshness.

The client keeps the last valid directory in `%ProgramData%\\Jamshidix\\directory.json` and tries multiple remote mirrors before falling back to the local cache.

Advanced CLI commands remain available for diagnostics: `status`, `stop`, `import`, `config`, `check`, `run`, `autostart`, and `uninstall`.

Public nodes are untrusted infrastructure. A node passing TCP checks is not guaranteed to be reachable from every local network or to be safe to use for sensitive traffic.