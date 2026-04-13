#!/bin/sh
# check-tools.sh — Vérifie les outils de développement requis
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'
BOLD='\033[1m'

ERRORS=0

check() {
  local name="$1"
  local cmd="$2"
  local install_hint="$3"
  local required="$4"

  if command -v "$cmd" >/dev/null 2>&1; then
    version=$($cmd --version 2>/dev/null | head -1 || $cmd version 2>/dev/null | head -1 || echo "installed")
    printf "  ${GREEN}✅${NC} %-20s %s\n" "$name" "$version"
  else
    if [ "$required" = "required" ]; then
      printf "  ${RED}❌${NC} %-20s ${RED}manquant${NC} → %s\n" "$name" "$install_hint"
      ERRORS=$((ERRORS + 1))
    else
      printf "  ${YELLOW}⚠️${NC}  %-20s ${YELLOW}optionnel${NC} → %s\n" "$name" "$install_hint"
    fi
  fi
}

echo ""
echo "${BOLD}🔧 Vérification des outils de développement${NC}"
echo "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "${BOLD}  Essentiels :${NC}"
check "Go"              "go"              "https://go.dev/dl/"                          "required"
check "golangci-lint"   "golangci-lint"   "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"  "required"
check "Wire"            "wire"            "go install github.com/google/wire/cmd/wire@latest"                     "required"
check "mockgen"         "mockgen"         "go install go.uber.org/mock/mockgen@latest"                            "required"
check "Docker"          "docker"          "https://docs.docker.com/get-docker/"         "required"

echo ""
echo "${BOLD}  Analyse statique :${NC}"
check "staticcheck"     "staticcheck"     "go install honnef.co/go/tools/cmd/staticcheck@latest"    "optional"
check "gosec"           "gosec"           "go install github.com/securego/gosec/v2/cmd/gosec@latest" "optional"
check "gocyclo"         "gocyclo"         "go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"  "optional"

echo ""
echo "${BOLD}  Infra (optionnel) :${NC}"
check "Skaffold"        "skaffold"        "https://skaffold.dev/docs/install/"          "optional"
check "kubectl"         "kubectl"         "https://kubernetes.io/docs/tasks/tools/"     "optional"
check "k6"              "k6"              "https://k6.io/docs/getting-started/installation/" "optional"

echo ""
if [ $ERRORS -gt 0 ]; then
  echo "${RED}❌ $ERRORS outil(s) requis manquant(s). Installez-les avant de continuer.${NC}"
  exit 1
else
  echo "${GREEN}✅ Tous les outils requis sont installés.${NC}"
fi
echo ""

