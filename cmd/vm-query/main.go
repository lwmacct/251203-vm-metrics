package main

import (
	"context"
	"fmt"
	"os"

	app "github.com/lwmacct/251203-vm-metrics/internal/command/query"
)

func main() {
	if err := app.Command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
