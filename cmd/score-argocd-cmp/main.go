package main

import (
	"encoding/json"
	"fmt"
	"os"
	"score-argocd-cmp/internal/discover"
	"score-argocd-cmp/internal/generate"
	"score-argocd-cmp/internal/initialize"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: score-argocd-cmp <discover|discover-params|init|generate>\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "discover":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		files, err := discover.Files(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(strings.Join(files, "\n"))

	case "discover-params":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		params, err := discover.Params(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		out, err := json.Marshal(params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))

	case "init":
		if err := initialize.Run(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "generate":
		dir := "."
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		output, err := generate.Run(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: score-argocd-cmp <discover|discover-params|init|generate>\n", os.Args[1])
		os.Exit(1)
	}
}
