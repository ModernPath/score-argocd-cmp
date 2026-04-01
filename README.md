# score-argocd-cmp

An [ArgoCD Config Management Plugin](https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/) that uses [score-k8s](https://github.com/score-spec/score-k8s) to translate [Score](https://score.dev) workload specifications into Kubernetes manifests. This enables developer self-service: app teams define workloads in Score format, and ArgoCD handles the rendering and deployment.

## How It Works

The plugin runs as a sidecar container alongside the ArgoCD repo-server. When ArgoCD syncs an Application:

1. **Discover** - The plugin checks if the repository contains a `score.yaml` at the root
2. **Init** - Initializes score-k8s with custom provisioners baked into the sidecar image
3. **Generate** - Renders `score.yaml` into Kubernetes manifests using the configured image and namespace

```
Application CR  →  repo-server  →  CMP sidecar  →  score-k8s  →  K8s manifests
                                    (this plugin)
```

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

Add the following to your ArgoCD Helm values (or equivalent repo-server configuration):

```yaml
configs:
  cmp:
    create: true
    plugins:
      score-k8s:
        discover:
          find:
            command: ["sh", "-c"]
            args: ["find . -maxdepth 1 -name 'score.yaml' | head -1 | grep -q ."]
        init:
          command: ["sh", "-c"]
          args:
            - |
              score-k8s init --no-sample \
                --provisioners /opt/provisioners/*.provisioners.yaml
        generate:
          command: ["sh", "-c"]
          args:
            - |
              score-k8s generate score.yaml \
                --image "${ARGOCD_ENV_IMAGE:-placeholder:latest}" \
                --namespace "${ARGOCD_ENV_NAMESPACE:-default}" \
                -o - 2>/dev/null

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
        - mountPath: /home/argocd/cmp-server/config/plugin.yaml
          subPath: plugin.yaml
          name: argocd-cmp-cm
        - mountPath: /tmp
          name: score-k8s-tmp
  volumes:
    - name: score-k8s-tmp
      emptyDir: {}
```

### 4. Create an Application

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
      env:
        - name: IMAGE
          value: "your-registry/hello-world:v1.0.0"
        - name: NAMESPACE
          value: "production"
  destination:
    server: https://kubernetes.default.svc
    namespace: production
```

The `plugin.env` entries become `ARGOCD_ENV_IMAGE` and `ARGOCD_ENV_NAMESPACE` inside the CMP.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ARGOCD_ENV_IMAGE` | `placeholder:latest` | Container image for the workload |
| `ARGOCD_ENV_NAMESPACE` | `default` | Target Kubernetes namespace |

These are set per-Application via `spec.source.plugin.env` in the Application CR.

### Custom Provisioners

Platform teams can add custom provisioners to the `provisioners/` directory before building the image. Provisioners define how Score resource types (postgres, redis, volume, etc.) map to Kubernetes resources.

See [provisioners/README.md](provisioners/README.md) for the naming convention and format.

## Testing

```bash
# Build and run integration tests
make test

# Run tests against a specific image
./tests/test-generate.sh your-image:tag
```

## Troubleshooting

**Plugin not discovered by ArgoCD:**
- Ensure `score.yaml` is at the repository root (not in a subdirectory)
- Check the CMP sidecar logs: `kubectl logs -n argocd deploy/argocd-repo-server -c score-k8s`

**Generate produces no output:**
- stderr is suppressed in the generate command. To debug, exec into the sidecar and run the generate command without `2>/dev/null`
- Verify provisioners exist at `/opt/provisioners/` inside the container

**`-o -` flag not supported:**
- If score-k8s does not support `-o -` for stdout output, replace the generate command with:
  ```yaml
  args:
    - |
      score-k8s generate score.yaml \
        --image "${ARGOCD_ENV_IMAGE:-placeholder:latest}" \
        --namespace "${ARGOCD_ENV_NAMESPACE:-default}" 2>/dev/null \
      && cat manifests.yaml
  ```

**Image shows as `placeholder:latest`:**
- Set `ARGOCD_ENV_IMAGE` via `spec.source.plugin.env` in the Application CR

## References

- [Score specification](https://score.dev)
- [score-k8s CLI](https://github.com/score-spec/score-k8s)
- [ArgoCD CMP documentation](https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/)
- [Engine's blog post on Score + ArgoCD](https://score.dev/blog/engine-self-service-kubernetes-platform-with-score/)
