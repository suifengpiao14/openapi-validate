package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionNumber = "0.0.1"

// versionCmd show version openapi-validate
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of openapi-validate",
	Long:  `All software has versions. This is pREST's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Serve a RESTful API for http request and response", versionNumber)
	},
}
