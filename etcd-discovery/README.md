# Etcd Discovery Service

This provides a simple discovery service for etcd. It is used during
intial Deis bootstrap when a new etcd cluster comes online.
Periodically, it may be used when an etcd node is recovering.

## How It Works

The `boot.go` code bootstraps a new etcd server running as a single
node. This server is used _only_ for discovery, and never joins a
cluster. It then reads its discovery token out of its secrets file, and
sets up the initial cluster expectations -- notably the size of the
cluster.

When the etcd cluster spins up, each node should...

- Get the discovery token from the secrets volume
- Find this server's discovery service from the environment
- Connect to this discovery service and walk through the discovery
  process

Initial size of the cluster is determined by etcd-discovery's
environment variable `$DEIS_ETCD_CLUSTER_SIZE`.
