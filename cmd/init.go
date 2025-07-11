package cmd

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	internalerror "github.com/Consoleaf/kopl/internal_error"
	"github.com/Consoleaf/kopl/luatemplates"
	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a koplugin project",
	ValidArgs: []cobra.Completion{
		cobra.CompletionWithDesc("path", "Path to the new project. Defaults to '.'"),
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			internalerror.ErrorExit(err)
		}
		projectName := path.Base(path.Join(cwd, args[0]))
		if !strings.HasSuffix(projectName, ".koplugin") {
			internalerror.ErrorExitf("KOreader plugin names end with .koreader. Got: %q", projectName)
		}
		projectName = strings.Replace(projectName, ".koplugin", "", 1)

		os.MkdirAll(args[0], 0o755)
		err = os.Chdir(path.Join(cwd, args[0]))
		if err != nil {
			internalerror.ErrorExit(err)
		}

		repo, err := git.PlainInitWithOptions(".", &git.PlainInitOptions{
			InitOptions: git.InitOptions{
				// DefaultBranch: "main",
			},
		})
		if err != nil {
			internalerror.ErrorExitf("While running `git init`: \n%v", err)
		}
		_ = repo
		vars := luatemplates.TemplateArgsForInit{
			ProjectName: convertToPascalCase(projectName),
		}

		writeTemplate(luatemplates.MetaFileTemplate, vars)
		writeTemplate(luatemplates.MainFileTemplate, vars)
		writeTemplate(luatemplates.LuaRcTemplate, vars)
		writeTemplate(luatemplates.IgnoreTemplate, vars)

		submoduleCmd := exec.Command("git", "submodule", "add", "--depth", "1", "https://github.com/koreader/koreader.git")
		submoduleCmd.Stdout = os.Stdout
		submoduleCmd.Stderr = os.Stderr
		err = submoduleCmd.Run()
		if err != nil {
			internalerror.ErrorExit(err)
		}
	},
}

func writeTemplate(tmpl template.Template, args luatemplates.TemplateArgsForInit) {
	filename := strings.Replace(tmpl.Name(), ".tmpl", "", 1)
	f, err := os.Create(filename)
	if err != nil {
		internalerror.ErrorExitf("While trying to create %v: \n%v", filename, err)
	}
	err = tmpl.Execute(f, args)
	if err != nil {
		internalerror.ErrorExitf("While trying to write to %v: \n%v", filename, err)
	}
}

func convertToPascalCase(text string) string {
    words := strings.FieldsFunc(text, func(r rune) bool {
        return !unicode.IsLetter(r) && !unicode.IsNumber(r)
    })
    var pascalCaseText string
    for _, word := range words {
        pascalCaseText += strings.Title(word)
    }
    return pascalCaseText
}