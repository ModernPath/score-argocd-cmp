# Custom Provisioners

This directory contains custom score-k8s provisioners that are baked into the CMP sidecar image.

The plugin loads all `*.provisioners.yaml` files from this directory during `score-k8s init`.

## File Naming Convention

Use a numbered prefix to control load order (lower number = higher priority):

```
00-volume.provisioners.yaml
01-postgres.provisioners.yaml
02-redis.provisioners.yaml
```

Custom provisioners loaded here take precedence over score-k8s built-in defaults.

## Writing Provisioners

Each provisioner file contains a YAML list of provisioner definitions. Example:

```yaml
- uri: template://my-org/volume
  type: volume
  class: default
  outputs: |
    source: {{ .SourceWorkload }}-{{ .ResourceUid }}
  manifests: |
    - apiVersion: v1
      kind: PersistentVolumeClaim
      metadata:
        name: {{ .SourceWorkload }}-{{ .ResourceUid }}
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: {{ dig "size" "1Gi" .Params }}
```

See the [score-k8s provisioner docs](https://github.com/score-spec/score-k8s) for the full template reference and available context variables.

## Upstream Defaults

score-k8s ships with built-in provisioners for common resource types (postgres, redis, volume, dns, etc.) that create in-cluster dev instances. For production, replace these with provisioners that integrate with your infrastructure (e.g., CloudNativePG, AWS RDS via Crossplane, Redis Operator).
