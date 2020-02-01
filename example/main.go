package main

import (
	"github.com/suifengpiao14/openapi-validate/cmd"
	"github.com/suifengpiao14/openapi-validate/config"
)

func main() {
	config.Load()
	cmd.Execute()
}
