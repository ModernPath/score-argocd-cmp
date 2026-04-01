#!/usr/bin/env bash
set -euo pipefail

IMAGE="${1:?Usage: test-generate.sh IMAGE:TAG}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Test 1: score-k8s binary exists ==="
docker run --rm "$IMAGE" score-k8s version
echo "PASS"

echo ""
echo "=== Test 2: provisioners directory exists ==="
docker run --rm "$IMAGE" ls /opt/provisioners/
echo "PASS"

echo ""
echo "=== Test 3: init + generate cycle ==="
OUTPUT=$(docker run --rm \
  -v "${SCRIPT_DIR}/score-test.yaml:/work/score.yaml:ro" \
  -w /work \
  -e ARGOCD_ENV_IMAGE=nginx:1.25 \
  -e ARGOCD_ENV_NAMESPACE=test-ns \
  --user 999 \
  "$IMAGE" \
  sh -c '
    score-k8s init --no-sample 2>/dev/null && \
    score-k8s generate score.yaml \
      --image "${ARGOCD_ENV_IMAGE}" \
      --namespace "${ARGOCD_ENV_NAMESPACE}" \
      -o - 2>/dev/null
  ')

if echo "$OUTPUT" | grep -q "apiVersion"; then
  echo "PASS: output contains Kubernetes manifests"
else
  echo "FAIL: output does not contain Kubernetes manifests"
  echo "Output was:"
  echo "$OUTPUT"
  exit 1
fi

if echo "$OUTPUT" | grep -q "nginx:1.25"; then
  echo "PASS: image override applied"
else
  echo "FAIL: image override not found in output"
  exit 1
fi

if echo "$OUTPUT" | grep -q "test-ns"; then
  echo "PASS: namespace override applied"
else
  echo "FAIL: namespace override not found in output"
  exit 1
fi

echo ""
echo "=== All tests passed ==="
