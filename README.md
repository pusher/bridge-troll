# Bridge Troll
Bridge Troll is an in-cluster watchdog for file changes.

[![Docker Repository on Quay](https://quay.io/repository/pusher/bridge-troll/status "Docker Repository on Quay")](https://quay.io/repository/pusher/bridge-troll)

## Use Case

Some processes that consume configuration from files do not support hot-reloading and
do not detect changes to the underlying files.
When running inside Kubernetes, such config files are often mounted into Pods from e.g.
a `ConfigMap`.

If the `ConfigMap` is updated in Kubernetes, the underlying files inside existing `Pods` will change,
but the processes may not become aware of this change until they (often accidentally) restart.

## Operation

Bridge Troll is meant to run as a "side-car" container inside a `Pod`, with the name and namespace
passed in as environment variables.
It will then calculate a hash on the contents of the provided set of files and store that hash as an
`Annotation` on the `Pod` it is running in.

It then commences to periodically compare the current hash of those files with the one calculated at `Pod` creation
time.

Bridge Troll emits a Prometheus metric `bridge-troll.monitoring.pusher.com/original-config-hash` which indicates
whether any files have changed since the `Pod` was created.

## Usage

Add a Bridge Troll container to your `Deployment` and pass in the `Pod`'s Name and Namespace via the downwards API:

```
apiVersion: apps/v1
Kind: Deployment
[...]
spec:
  [...]
  template:
    metadata:
      [...]
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/scheme: http
        prometheus.io/port: multiple
        prometheus.io/path: /metrics
  spec:
    serviceAccount: <service_account_name>
    containers:
    [...]
    - name: bridge-troll
      image: quay.io/pusher/bridge-troll:v0.1
      args: ["-f", "<some_file>", "-f", "<some_other_file>"]
      env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      ports:
      - name: metrics
        containerPort: <metrics-port>
      volumeMounts:
      - mountPath: <some_path>
        name: <shared_volume_with_watchfiles>
      resources:
        limits:
          cpu: 100m
          memory: 100Mi
        requests:
          cpu: 100mm
          memory: 100Mi

[...]
```

## Commandline Options

Bridge Troll supports the following options:
```
-i, --check-interval int      The numer of seconds between checks of the watchfile contents (default 10)
-h, --help                    This help
-m, --metrics-path string     The path for the metrics endpoint (default "/metrics")
-p, --metrics-port int        The metrics port to use (default 2112)
-f, --watchfile stringArray   The file to watch. Can be used multiple times
```

## RBAC

If you are using RBAC, you will need to provide a service account for the container and grant it the required permissions.

Example:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: update-pods
  namespace: default
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - update
  - patch
```

```
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: update-pods
  namespace: default
subjects:
- kind: ServiceAccount
  name: <service_account_name>
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: update-pods
```
