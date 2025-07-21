package httpinspector

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	koreaderinspector "github.com/Consoleaf/koreader-http-inspector"
	"github.com/spf13/cobra"
)

var (
	Host string = "192.168.15.244"
	Port int    = 8080

	Inspector *koreaderinspector.HTTPInspectorClient
)

func AddArgs(command *cobra.Command) {
	envHost, exists := os.LookupEnv("KOREADER_INSPECTOR_HOST")
	if exists {
		Host = envHost
	}
	envPort, exists := os.LookupEnv("KOREADER_INSPECTOR_PORT")
	if exists {
		port, err := strconv.Atoi(envPort)
		if err == nil {
			Port = port
		}
	}

	command.Flags().StringVarP(
		&Host,
		"host",
		"H",
		Host,
		"Network address of the KOReader instance. You can also set this in envvar KOREADER_INSPECTOR_HOST",
	)
	command.Flags().IntVarP(
		&Port,
		"port",
		"p",
		Port,
		"HTTP Inspector port. You can also set this in envvar KOREADER_INSPECTOR_PORT")
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
