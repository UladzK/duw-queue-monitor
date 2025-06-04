package main

import (
	"fmt"
	"uladzk/duw_kolejka_checker/internal/statuscollector"

	"github.com/caarlos0/env/v11"
)

func main() {
	// TODO: add graceful shutdown to finalize status processing and save the state before exiting

	var cfg statuscollector.Config

	err := env.Parse(&cfg)

	if err != nil {
		panic("Failed to get environment variables: " + err.Error())
	}

	handler := statuscollector.NewHandler(&cfg)

	fmt.Println("Status collector started")
	handler.Run()

	fmt.Println("Status collector stopped")
}
