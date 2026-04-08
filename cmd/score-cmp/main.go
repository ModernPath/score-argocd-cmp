package main

import (
	"encoding/json"
	"fmt"
	"os"
	"score-argocd-cmp/internal/discover"
	"score-argocd-cmp/internal/resolve"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: score-cmp <discover|discover-params|resolve-image>\n")
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

	case "resolve-image":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: score-cmp resolve-image <score-file>\n")
			os.Exit(1)
		}
		dir := "."
		img, err := resolve.Run(os.Args[2], dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(img)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: score-cmp <discover|discover-params|resolve-image>\n", os.Args[1])
		os.Exit(1)
	}
}
