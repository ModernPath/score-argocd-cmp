# score-argocd-cmp

An [ArgoCD Config Management Plugin](https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/) that uses [score-k8s](https://github.com/score-spec/score-k8s) to translate [Score](https://score.dev) workload specifications into Kubernetes manifests. This enables developer self-service: app teams define workloads in Score format, and ArgoCD handles the rendering and deployment.

## How It Works

The plugin runs as a sidecar container alongside the ArgoCD repo-server. When ArgoCD syncs an Application:

1. **Discover** — Checks for `score.yaml` or `*.score.yaml` files in the repository root
2. **Init** — Initializes score-k8s with provisioners fetched from a URL (`PARAM_PROVISIONERS_URL`)
3. **Generate** — Renders each score file into Kubernetes manifests using the resolved image and namespace

```
Application CR  →  repo-server  →  score-argocd-cmp (CMP sidecar)  →  score-k8s  →  K8s manifests
```

The `score-argocd-cmp` binary wraps `score-k8s`, handling discovery, image resolution, and multi-workload support. It runs in a sidecar container next to the ArgoCD repo-server and is invoked via the CMP protocol.

## Quick Start

### 1. Build the Sidecar Image

```bash
make build
# Or with a pinned score-k8s version:
make build SCORE_K8S_VERSION=v0.10.3
```

### 2. Push to Your Registry

```bash
docker tag score-argocd-cmp:latest ghcr.io/your-org/score-argocd-cmp:v1.0.0
docker push ghcr.io/your-org/score-argocd-cmp:v1.0.0
```

### 3. Install as ArgoCD Sidecar

The plugin.yaml is baked into the image at `/home/argocd/cmp-server/config/plugin.yaml`. Add the sidecar to your ArgoCD repo-server configuration:

```yaml
repoServer:
  extraContainers:
    - name: score-k8s
      image: ghcr.io/your-org/score-argocd-cmp:v1.0.0
      command: ["/var/run/argocd/argocd-cmp-server"]
      securityContext:
        runAsNonRoot: true
        runAsUser: 999
      volumeMounts:
        - mountPath: /var/run/argocd
          name: var-files
        - mountPath: /home/argocd/cmp-server/plugins
          name: plugins
        - mountPath: /tmp
          name: score-k8s-tmp
  volumes:
    - name: score-k8s-tmp
      emptyDir: {}
```

### 4. Create an Application

Single workload (`score.yaml`):

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: hello-world
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/your-app.git
    targetRevision: main
    path: .
    plugin:
      parameters:
        - name: provisioners-url
          string: "oci://ghcr.io/your-org/score-k8s-provisioners:v1.0.0"
        - name: instance-name
          string: "hello-world"
        - name: image
          string: "your-registry/hello-world:v1.0.0"
  destination:
    server: https://kubernetes.default.svc
    namespace: production
```

Multi-workload (`*.score.yaml` files, e.g. `frontend.score.yaml` and `backend.score.yaml`):

```yaml
    plugin:
      parameters:
        - name: provisioners-url
          string: "oci://ghcr.io/your-org/score-k8s-provisioners:v1.0.0"
        - name: instance-name
          string: "my-app"
        - name: image-frontend
          string: "your-registry/frontend:v1.0.0"
        - name: image-backend
          string: "your-registry/backend:v1.0.0"
```

Image parameters are named `image` in single mode, or `image-<workload-name>` in multi-workload mode.

## Configuration

### Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `provisioners-url` | Yes | URI of the provisioners bundle (`oci://…`, `https://…`, or local path). Passed verbatim to `score-k8s init --provisioners`. |
| `instance-name` | Yes | Short, stable identifier used by provisioners as a name component for generated K8s objects (e.g. `pg-<instance-name>-<id>`). Kept separate from the Application name so rendered K8s names stay within the 63-char limit. Must be unique across deployments sharing a namespace. |
| `image` | Yes (single mode) | Container image for the workload |
| `image-<name>` | Yes (multi mode) | Container image per workload (name from score file metadata) |
| `no-default-provisioners` | No (default `false`) | When `true`, passes `--no-default-provisioners` to `score-k8s init` so only the bundle from `provisioners-url` is loaded. |
| `debug` | No (default `false`) | When `true`, enables verbose logging in score-argocd-cmp and passes `--verbose` to `score-k8s`. |

Parameters are set per-Application via `spec.source.plugin.parameters`. ArgoCD exports them to the CMP as `PARAM_*` environment variables (name uppercased, `-` → `_`, e.g. `provisioners-url` → `PARAM_PROVISIONERS_URL`) and also as a JSON blob in `ARGOCD_APP_PARAMETERS`.

### Environment Variables

| Variable | Source | Description |
|----------|--------|-------------|
| `ARGOCD_APP_NAMESPACE` | ArgoCD (automatic) | Target namespace from the Application destination. Used as the rendered manifest namespace. |
| `PARAM_PROVISIONERS_URL` | Plugin parameter | URI of the provisioners bundle. Read during `init`. |
| `PARAM_NO_DEFAULT_PROVISIONERS` | Plugin parameter | `"true"` disables built-in score-k8s provisioners. |
| `PARAM_DEBUG` / `SCORE_CMP_DEBUG` | Plugin parameter / manual | Enable verbose logging in the plugin and `score-k8s`. |
| `ARGOCD_APP_PARAMETERS` | ArgoCD (automatic) | JSON array of all plugin parameters. Read during `generate` to resolve `image` / `image-<name>`; falls back to `PARAM_IMAGE` / `PARAM_IMAGE_*` when unset. |

## Testing

```bash
# Build and run integration tests
make test

# Run Go unit tests only
make go-test

# Run tests against a specific image
./tests/test-generate.sh your-image:tag
```

## Troubleshooting

**Plugin not discovered by ArgoCD:**
- Ensure `score.yaml` or `*.score.yaml` files exist at the repository root
- Check the CMP sidecar logs: `kubectl logs -n argocd deploy/argocd-repo-server -c score-k8s`

**Generate fails or produces empty output:**
- Check the error output — stderr from `score-k8s` is captured in the error message
- Verify the `provisioners-url` parameter points to valid provisioner definitions
- Exec into the sidecar and run `score-argocd-cmp generate` manually to debug

**Image shows as `placeholder:latest`:**
- Set the `image` (or `image-<name>`) parameter in the Application CR's `spec.source.plugin.parameters`

## References

- [Score specification](https://score.dev)
- [score-k8s CLI](https://github.com/score-spec/score-k8s)
- [ArgoCD CMP documentation](https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/)

## License

MIT — see [LICENSE.txt](LICENSE.txt).
