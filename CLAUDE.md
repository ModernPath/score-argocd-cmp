# score-argocd-cmp

ArgoCD Config Management Plugin (CMP) sidecar that renders Score workload specs into K8s manifests via score-k8s.

## Architecture

- `plugin.yaml` — CMP definition: discover (score.yaml exists?), init (load provisioners), generate (render manifests to stdout)
- `Dockerfile` — Multi-stage: builds score-k8s from source, produces minimal Alpine sidecar image
- Runs as non-root user 999 alongside argocd-repo-server
- ArgoCD passes `ARGOCD_ENV_*` vars from Application CR `spec.source.plugin.env`

## Build & Test

```
make build                              # Docker build
make test                               # Build + integration test
make build SCORE_K8S_VERSION=v0.10.3    # Pin version
```

## Score Provisioners

Provisioners map Score `resources:` types to Kubernetes objects. Files live in `provisioners/` and are baked into the image at `/opt/provisioners/`.

### Key concepts

- File format: YAML list of provisioner definitions, named `NN-name.provisioners.yaml`
- Lower numbered prefix = higher priority; custom provisioners override score-k8s built-ins
- Loaded via glob: `--provisioners /opt/provisioners/*.provisioners.yaml`
- Template engine: Go templates with Sprig functions

### Template context variables

- `{{ .SourceWorkload }}` — name of the Score workload
- `{{ .ResourceUid }}` — unique ID for the resource instance
- `{{ .Params }}` — parameters from the Score resource definition (access with `dig`)
- `{{ .Metadata }}` — workload metadata

### Provisioner structure

```yaml
- uri: template://org/resource-name    # unique identifier
  type: postgres                        # matches Score resource type
  class: default                        # optional class selector
  outputs: |                            # Go template producing YAML map of outputs
    host: {{ .SourceWorkload }}-db
    port: "5432"
  manifests: |                          # Go template producing K8s manifest list
    - apiVersion: v1
      kind: ConfigMap
      ...
```

Outputs become available to the workload as `${resources.<name>.<key>}` in score.yaml.
