package luatemplates

import (
	"embed"
	"text/template"

	internalerror "github.com/Consoleaf/kopl/internal_error"
)

//go:embed *.tmpl
var templates embed.FS

var (
	MetaFileTemplate template.Template
	MainFileTemplate template.Template
	LuaRcTemplate    template.Template
	IgnoreTemplate   template.Template
)

type TemplateArgsForInit struct {
	ProjectName string
}

func init() {
	MetaFileTemplate = parse("_meta.lua")
	MainFileTemplate = parse("main.lua")
	LuaRcTemplate = parse(".luarc.json")
	IgnoreTemplate = parse(".ignore")
}

func parse(filename string) template.Template {
	tmpl, err := template.ParseFS(templates, filename+".tmpl")
	if err != nil {
		internalerror.ErrorExitf("While parsing _meta.lua.tmpl: %q", err)
	}
	return *tmpl
}
