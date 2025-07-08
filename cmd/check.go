package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	internalerror "github.com/Consoleaf/kopl/internal_error"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Perform static checks on the project",
	Run: func(cmd *cobra.Command, args []string) {
		projPath, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		_ = assertFileExists(path.Join(projPath, "_meta.lua"))
		_ = assertFileExists(path.Join(projPath, "main.lua"))

		ensureLuacheck()

		fmt.Fprintln(os.Stderr, "Running luacheck...")

		luacheck := exec.Command("luacheck", ".", "--exclude-files", "koreader")
		luacheck.Stdout = os.Stdout
		luacheck.Stderr = os.Stderr
		err = luacheck.Run()
		if err != nil {
			internalerror.ErrorExitf("While running luacheck:\n%v\n", err)
		}
	},
}

func assertFileExists(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		internalerror.ErrorExitf("File doesn't exist: %s\n%q\n", path, err)
	}
	return file
}

func ensureLuacheck() {
	_, err := exec.LookPath("luacheck")
	if err == nil {
		return
	}

	importLuarocksPath()
	_, err = exec.LookPath("luacheck")
	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, "luacheck not found. Trying to install it using luarocks...")

	luarocks := getLuarocks()
	cmd := exec.Command(luarocks, "install", "--local", "luacheck")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		internalerror.ErrorExit(err)
	}
}

func getLuarocks() string {
	luarocks, err := exec.LookPath("luarocks")
	if err != nil {
		internalerror.ErrorExit("luarocks not found")
	}
	return luarocks
}

func importLuarocksPath() {
	luarocks := getLuarocks()
	luaRocksPathVar, err := exec.Command(luarocks, "path", "--lr-bin").Output()
	if err != nil {
		internalerror.ErrorExitf("While getting luarocks PATH: %v", err)
	}
	if !strings.Contains(os.Getenv("PATH"), string(luaRocksPathVar)) {
		fmt.Println("WARNING: Luarocks PATH isn't set up. This CLI will call Luacheck directly, but other tools might not.")

		os.Setenv(
			"PATH",
			strings.Join(
				[]string{
					os.Getenv("PATH"),
					string(luaRocksPathVar),
				},
				string(os.PathListSeparator),
			),
		)
	}
}
