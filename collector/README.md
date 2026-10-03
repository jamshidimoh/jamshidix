# Jamshidix public-node collector

The scheduled collector downloads public VLESS sources, keeps only VLESS/REALITY/TCP entries compatible with the Windows client, de-duplicates them, performs a remote TCP reachability check, and publishes `directory/nodes.json`.

Public nodes are untrusted infrastructure. The collector does not claim that a TCP-open node is safe or reachable from every local network.
