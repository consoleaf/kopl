package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	httpinspector "github.com/Consoleaf/kopl/http_inspector"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(replCmd)
	httpinspector.AddArgs(replCmd)
}

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Open a REPL",
	Run: func(cmd *cobra.Command, args []string) {
		err := main(cmd, args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func main(_ *cobra.Command, _ []string) error {
	const (
		exitCommand = "exit"
		quitCommand = "quit"
	)

	httpinspector.Initialize()

	httpinspector.Inspector.Get("/ui/Repl/clean/")

	fmt.Printf(
		"KOReader Lua REPL. Type '%s' or '%s' to exit.",
		exitCommand,
		quitCommand,
	)

	prompt := promptui.Prompt{}

	var incomplete bool

	for {
		if incomplete {
			prompt.Label = "..."
		} else {
			prompt.Label = ">>>"
		}

		input, err := prompt.Run()
		if err != nil {
			return err
		}

		if input == exitCommand || input == quitCommand {
			fmt.Println("Exiting REPL.")
			break
		}

		if strings.Trim(input, "") == "" {
			continue
		}

		result, out, complete, err := Evaluate(input)
		incomplete = !complete
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error from KOReader: %v\n", err)
		} else if !complete {
		} else {
			if len(out) > 0 {
				fmt.Println("[OUT] " + strings.Join(out, "\n[OUT] "))
			}
			fmt.Printf("[RET] %v\n", result)
		}
	}

	return nil
}

type ReplResponse struct {
	Error  string        `json:"error"`
	Return any           `json:"ret"`
	Output outputStrings `json:"out"`
}

func Evaluate(code string) (any, []string, bool, error) {
	ret, err := httpinspector.Inspector.Get("/ui/Repl/repl/" + base64.StdEncoding.EncodeToString([]byte(code)))
	if err != nil {
		return "", []string{}, true, err
	}

	var _response []ReplResponse
	err = json.Unmarshal(ret, &_response)
	if err != nil {
		return "", []string{}, true, fmt.Errorf("%v\nValue: %v", err, string(ret))
	}

	response := _response[0]

	if response.Error != "" {
		if response.Error == "code is incomplete" {
			return "", []string{}, false, nil
		}
		return "", []string{}, true, fmt.Errorf("%v", response.Error)
	}

	rv := reflect.ValueOf(response.Return)
	if rv.Kind() == reflect.Map {
		nKey := reflect.ValueOf("n")
		nValue := rv.MapIndex(nKey) // Get the value associated with key "n"

		if nValue.Kind() == reflect.Interface && !nValue.IsNil() {
			nValue = nValue.Elem() // Get the concrete value stored inside the interface
		}

		var isZero bool

		if nValue.IsValid() {
			switch nValue.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				isZero = nValue.Int() == 0
			case reflect.Float32, reflect.Float64:
				isZero = nValue.Float() == 0
			}
		}

		if isZero {
			return nil, response.Output, true, nil
		}
	}
	if rv.Kind() == reflect.Slice {
		return rv.Index(0).Interface(), response.Output, true, nil
	}

	return nil, response.Output, true, nil
}

// We need this because if an empty table is sent from Lua, it'll be encoded as {}
type outputStrings []string

func (o *outputStrings) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string array first.
	var stringArray []string
	if err := json.Unmarshal(data, &stringArray); err == nil {
		*o = stringArray
		return nil
	}

	// If it's not a string array, check if it's an empty object.
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		// If it's not unmarshable as raw interface, something is fundamentally wrong
		return err
	}

	if _, ok := raw.(map[string]any); ok {
		// If it's an empty object (e.g., {}), treat it as an empty slice.
		*o = []string{}
		return nil
	}

	return fmt.Errorf("unexpected type for OutputStrings: %s", data)
}
