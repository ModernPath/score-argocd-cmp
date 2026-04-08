package debug

import (
	"fmt"
	"os"
	"strings"
)

var enabled = os.Getenv("SCORE_CMP_DEBUG") != ""

func Logf(format string, args ...any) {
	if enabled {
		fmt.Fprintf(os.Stderr, "[score-argocd-cmp] "+format+"\n", args...)
	}
}

func LogCmd(name string, args []string) {
	if enabled {
		fmt.Fprintf(os.Stderr, "[score-argocd-cmp] exec: %s %s\n", name, strings.Join(args, " "))
	}
}
