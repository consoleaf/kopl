package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const MinimalRequiredReplKoplugin = "v0.0.3"

func init() {
	rootCmd.AddCommand(replCmd)
	AddInspectorArgs(replCmd)
	AddSSHFlags(replCmd)
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

	InitializeInspector()

	err := ensureReplPlugin()
	if err != nil {
		return err
	}

	fmt.Printf(
		"KOReader Lua REPL. Type '%s' or '%s' to exit.\n",
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
			logger.Error(err.Error())
		} else if !complete {
		} else {
			if len(out) > 0 {
				fmt.Println("[OUT] " + strings.Join(out, "\n[OUT] "))
			}
			if result != "<nil>" {
				fmt.Printf("[RET] %v\n", result)
			}
		}
	}

	return nil
}

type ReplResponse struct {
	Error  string        `json:"error"`
	Return string        `json:"ret"`
	Output outputStrings `json:"out"`
}

func Evaluate(code string) (any, []string, bool, error) {
	ret, err := Inspector.Get("/ui/Repl/repl/" + base64.StdEncoding.EncodeToString([]byte(code)))
	if err != nil {
		return "", []string{}, true, err
	}

	var _response []ReplResponse
	err = json.Unmarshal(ret, &_response)
	if err != nil {
		return "", []string{}, true, fmt.Errorf(
			"couldn't parse response: %v\nValue: %v",
			err,
			string(ret),
		)
	}

	response := _response[0]

	if response.Error != "" {
		if response.Error == "code is incomplete" {
			return "", []string{}, false, nil
		}
		return "", []string{}, true, fmt.Errorf("error from REPL: %v", response.Error)
	}

	return response.Return, response.Output, true, nil
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

func ensureReplPlugin() error {
	logger.Warn("Checking repl.koplugin")
	res, err := Inspector.Get("ui/Repl/fullname")
	if err != nil {
		return err
	}
	if string(res) == "Repl" {
		if !isPluginOutdated() {
			return nil
		}
		logger.Warn("repl.koplugin is outdated. Updating...")
	}
	logger.Info("Installing 'consoleaf/repl.koplugin'")
	return installImpl("consoleaf/repl.koplugin")
}

func isPluginOutdated() bool {
	res, _ := Inspector.Get("ui/Repl/version")
	parsed := string(res)
	if strings.Contains(parsed, "Res: No such table/object key: version") {
		return true
	}
	return semver.Compare(parsed, MinimalRequiredReplKoplugin) < 0
}
