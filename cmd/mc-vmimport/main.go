package main

import (
	"context"
	"fmt"
	"os"

	app "github.com/lwmacct/251203-mc-metrics/internal/command/import"
)

func main() {
	if err := app.Command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
