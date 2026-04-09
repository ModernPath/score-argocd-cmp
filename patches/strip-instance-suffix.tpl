{{/*
  Strip the random InstanceSuffix that score-k8s injects into workload labels.

  Background:
    score-k8s/internal/command/generate.go:164 generates a 5-byte random hex
    suffix per workload (e.g. "backend-bfa69db624") and persists it in
    .score-k8s/state.yaml. The suffix is used as the value of:
      - Deployment .metadata.labels.app.kubernetes.io/instance
      - Deployment .spec.selector.matchLabels.app.kubernetes.io/instance  (immutable!)
      - Deployment .spec.template.metadata.labels.app.kubernetes.io/instance
      - workload Service .metadata.labels.app.kubernetes.io/instance
      - workload Service .spec.selector.app.kubernetes.io/instance

    In the score-argocd-cmp pipeline the .score-k8s state directory is thrown
    away after each render, so the suffix is regenerated on every Argo sync,
    producing a label diff and an immutable-selector update that the API server
    rejects. We strip the suffix here so the rendered manifests are stable.

  Identification:
    workload-owned manifests carry the `k8s.score.dev/workload-name` annotation
    (set in score-k8s/internal/convert/workloads.go:169). Resource provisioner
    manifests do NOT carry this annotation, so they are left untouched.

  Output discipline:
    These docs live inside a Go-template comment so the rendered patch stream
    is empty when no Deployment/Service workload manifests exist. score-k8s
    skips patch application entirely on empty input
    (score-k8s/internal/patching/patching.go:96), so an empty render is a
    no-op rather than a YAML decode error.
*/}}
{{- range $i, $m := .Manifests }}
{{- $wn := dig "metadata" "annotations" "k8s.score.dev/workload-name" "" $m }}
{{- if and (eq $m.kind "Deployment") (ne $wn "") }}
- op: set
  path: {{ $i }}.metadata.labels.app\.kubernetes\.io/instance
  value: {{ $wn | quote }}
  description: strip InstanceSuffix from Deployment label (workload {{ $wn }})
- op: set
  path: {{ $i }}.spec.selector.matchLabels.app\.kubernetes\.io/instance
  value: {{ $wn | quote }}
  description: strip InstanceSuffix from Deployment selector (workload {{ $wn }})
- op: set
  path: {{ $i }}.spec.template.metadata.labels.app\.kubernetes\.io/instance
  value: {{ $wn | quote }}
  description: strip InstanceSuffix from Deployment pod template (workload {{ $wn }})
{{- end }}
{{- if and (eq $m.kind "Service") (ne $wn "") }}
- op: set
  path: {{ $i }}.metadata.labels.app\.kubernetes\.io/instance
  value: {{ $wn | quote }}
  description: strip InstanceSuffix from workload Service label (workload {{ $wn }})
- op: set
  path: {{ $i }}.spec.selector.app\.kubernetes\.io/instance
  value: {{ $wn | quote }}
  description: strip InstanceSuffix from workload Service selector (workload {{ $wn }})
{{- end }}
{{- end }}
