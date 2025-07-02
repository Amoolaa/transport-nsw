package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"transport-nsw-exporter/internal/server"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "transport-nsw-exporter",
		Usage: "prometheus exporter for transport NSW data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "web.listen-address",
				Value: ":8080",
				Usage: "address on which to expose metrics",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			s := server.New(cmd.String("web.listen-address"), logger)
			return s.Run()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
