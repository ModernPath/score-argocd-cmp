#!/usr/bin/env bash
set -euo pipefail

IMAGE="${1:?Usage: test-generate.sh IMAGE:TAG}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Test 1: score-k8s binary exists ==="
docker run --rm "$IMAGE" score-k8s --version
echo "PASS"

echo ""
echo "=== Test 2: score-argocd-cmp binary exists ==="
docker run --rm "$IMAGE" score-argocd-cmp 2>&1 || true
docker run --rm "$IMAGE" which score-argocd-cmp
echo "PASS"

echo ""
echo "=== Test 3: score-k8s supports --provisioners flag ==="
docker run --rm "$IMAGE" score-k8s init --help 2>&1 | grep -q "\-\-provisioners"
echo "PASS"

echo ""
echo "=== Test 4: plugin.yaml is in the correct location ==="
docker run --rm "$IMAGE" cat /home/argocd/cmp-server/config/plugin.yaml | grep -q "score-k8s"
echo "PASS"

echo ""
echo "=== Test 5: discover-params — single score.yaml ==="
OUTPUT=$(docker run --rm \
  -v "${SCRIPT_DIR}/score-test.yaml:/work/score.yaml:ro" \
  --user 999 \
  "$IMAGE" \
  sh -c 'cd /work && score-argocd-cmp discover-params')
if echo "$OUTPUT" | grep -q '"name":"image"'; then
  echo "PASS: single mode outputs image parameter"
else
  echo "FAIL: expected image parameter"
  echo "Output: $OUTPUT"
  exit 1
fi

echo ""
echo "=== Test 6: discover-params — multiple *.score.yaml ==="
OUTPUT=$(docker run --rm \
  -v "${SCRIPT_DIR}/backend.score.yaml:/work/backend.score.yaml:ro" \
  -v "${SCRIPT_DIR}/frontend.score.yaml:/work/frontend.score.yaml:ro" \
  --user 999 \
  "$IMAGE" \
  sh -c 'cd /work && score-argocd-cmp discover-params')
if echo "$OUTPUT" | grep -q '"image-backend"' && echo "$OUTPUT" | grep -q '"image-frontend"'; then
  echo "PASS: multi mode outputs image-backend and image-frontend"
else
  echo "FAIL: expected image-backend and image-frontend"
  echo "Output: $OUTPUT"
  exit 1
fi

echo ""
echo "=== Test 7: discover-params — 0 files errors ==="
if docker run --rm --user 999 "$IMAGE" sh -c 'cd /work && score-argocd-cmp discover-params' 2>/dev/null; then
  echo "FAIL: expected non-zero exit for 0 score files"
  exit 1
else
  echo "PASS: exits non-zero for 0 score files"
fi

echo ""
echo "=== Test 8: single score file — init + generate cycle ==="
OUTPUT=$(docker run --rm \
  -v "${SCRIPT_DIR}/score-test.yaml:/work/score.yaml:ro" \
  -e ARGOCD_APP_PARAMETERS='[{"name":"image","string":"nginx:1.25"}]' \
  -e ARGOCD_APP_NAMESPACE=test-ns \
  --user 999 \
  "$IMAGE" \
  sh -c '
    cd /work &&
    score-argocd-cmp init &&
    score-argocd-cmp generate
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
echo "=== Test 9: multi-workload generation with per-workload images ==="
PROV_DIR="$(cd "${SCRIPT_DIR}/../provisioners" 2>/dev/null && pwd || echo "")"
if [ -z "$PROV_DIR" ] || [ ! -d "$PROV_DIR" ]; then
  echo "SKIP: provisioners directory not found (expected in ../provisioners or mount externally)"
  echo "To run this test, provide provisioners via PROVISIONERS_DIR env var"
  PROV_DIR="${PROVISIONERS_DIR:-}"
fi

PROV_VOLUMES=""
PROV_INIT_FLAGS=""
if [ -n "$PROV_DIR" ] && [ -d "$PROV_DIR" ]; then
  for f in "$PROV_DIR"/*.provisioners.yaml; do
    [ -f "$f" ] && PROV_VOLUMES="$PROV_VOLUMES -v $f:/src/provisioners/$(basename $f):ro"
    [ -f "$f" ] && PROV_INIT_FLAGS="$PROV_INIT_FLAGS --provisioners /src/provisioners/$(basename $f)"
  done
fi

OUTPUT=$(docker run --rm \
  -v "${SCRIPT_DIR}/backend.score.yaml:/src/backend.score.yaml:ro" \
  -v "${SCRIPT_DIR}/frontend.score.yaml:/src/frontend.score.yaml:ro" \
  $PROV_VOLUMES \
  -e ARGOCD_APP_PARAMETERS='[{"name":"image-backend","string":"my-registry/backend:v1.0"},{"name":"image-frontend","string":"my-registry/frontend:v2.0"}]' \
  -e ARGOCD_APP_NAMESPACE=multi-ns \
  -e ARGOCD_ENV_DOMAIN=app.example.com \
  --user 999 \
  "$IMAGE" \
  bash -c "
    cp /src/*.score.yaml /work/ 2>/dev/null; cp /src/provisioners/* /work/ 2>/dev/null || true
    cd /work &&
    score-argocd-cmp init $PROV_INIT_FLAGS &&
    score-argocd-cmp generate
  ")

if echo "$OUTPUT" | grep -q "my-registry/backend:v1.0"; then
  echo "PASS: backend image override applied"
else
  echo "FAIL: backend image override not found"
  echo "Output was:"
  echo "$OUTPUT"
  exit 1
fi

if echo "$OUTPUT" | grep -q "my-registry/frontend:v2.0"; then
  echo "PASS: frontend image override applied"
else
  echo "FAIL: frontend image override not found"
  exit 1
fi

if echo "$OUTPUT" | grep -q "multi-ns"; then
  echo "PASS: namespace override applied to multi-workload output"
else
  echo "FAIL: namespace override not found in multi-workload output"
  exit 1
fi

if echo "$OUTPUT" | grep -q "pgvector"; then
  echo "PASS: pgvector postgres provisioner used"
else
  echo "FAIL: pgvector image not found (custom postgres provisioner not loaded)"
  exit 1
fi

if echo "$OUTPUT" | grep -q "HTTPRoute"; then
  echo "PASS: HTTPRoute generated by route provisioner"
else
  echo "FAIL: HTTPRoute not found in output"
  exit 1
fi

echo ""
echo "=== All tests passed ==="
