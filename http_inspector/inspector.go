package httpinspector

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	koreaderinspector "github.com/Consoleaf/koreader-http-inspector"
	"github.com/spf13/cobra"
)

var (
	Host string
	Port int

	Inspector *koreaderinspector.HTTPInspectorClient
)

func AddArgs(command *cobra.Command) {
	command.Flags().StringVarP(
		&Host,
		"host",
		"H",
		"192.168.15.244",
		"Network address of the KOReader instance. Defaults to 192.168.15.244 (default for Usbnetlite)",
	)
	command.Flags().IntVarP(
		&Port,
		"port",
		"p",
		8080,
		"HTTP Inspector port. Defaults to 8080",
	)
}

func Initialize() {
	var err error
	Inspector, err = koreaderinspector.New(fmt.Sprintf("http://%s:%d/", Host, Port))
	if err != nil {
		log.Fatal(err)
	}
	level := slog.LevelInfo
	_, present := os.LookupEnv("DEBUG")
	if present {
		level = slog.LevelDebug
	}
	Inspector.Logger = *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
