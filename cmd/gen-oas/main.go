package main

import (
	"fmt"
	"net/http"
	"os"

	"oas-sandbox/internal/server"
)

func main() {
	os.Exit(run())
}

func run() int {
	api := server.NewAPI(http.NewServeMux(), "oas-sandbox", nil, nil)
	body, err := api.OpenAPI().YAML()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "generate OpenAPI: %v\n", err)
		return 1
	}

	fmt.Print(string(body))
	return 0
}
