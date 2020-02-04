package main

import (
	"github.com/suifengpiao14/openapi-validate/cmd"
	"github.com/suifengpiao14/openapi-validate/config"
	"github.com/suifengpiao14/openapi-validate/middlewares"
)

func main() {
	// DocFile openapi json file path
	docFile := "doc/openapi.json"
	DocBytes, err := Asset(docFile)
	if err != nil {
		panic("openapi json file not found")
	}
	middlewares.DocBytes = DocBytes
	config.Load()
	cmd.Execute()
}
