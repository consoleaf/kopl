package internalerror

import (
	"fmt"
	"os"
)

type KoPluginDevErr struct {
	Message string
}

func (err *KoPluginDevErr) Error() string {
	return err.Message
}

func KoError(message string) KoPluginDevErr {
	return KoPluginDevErr{Message: message}
}

func ErrorExit(err any) {
	switch error := err.(type) {
	case KoPluginDevErr:
		fmt.Fprintln(os.Stderr, error.Error())
	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", error)
	}
	os.Exit(1)
}

func ErrorExitf(f string, args ...any) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}
