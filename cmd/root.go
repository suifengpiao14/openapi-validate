package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	nlog "github.com/nuveo/log"
	"github.com/spf13/cobra"
	"github.com/suifengpiao14/openapi-validate/config"
	"github.com/suifengpiao14/openapi-validate/config/router"
	"github.com/suifengpiao14/openapi-validate/controllers"
	"github.com/suifengpiao14/openapi-validate/middlewares"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Serve validate API request,response etc.",
	Long:  `Serve validate API request,response, api ui and api log`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

var daemon bool

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	RootCmd.AddCommand(versionCmd)

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// NotFound catch all not match route
func NotFound(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "{\"name\":\" world\",\"hello\":\"hello\",\"user\":[{\"id\":1,\"name\":\"test\"},{\"id\":2,\"name\":\"test2\"}]}")
}

// MakeHandler reagister all routes
func MakeHandler() http.Handler {
	n := middlewares.GetApp()
	r := router.Get()
	validateController := controllers.NewValidate()
	r.HandleFunc("/api/v1/validate/request", validateController.Request).Methods("POST")
	r.HandleFunc("/api/v1/validate/response", validateController.Response).Methods("POST")
	r.NotFoundHandler = http.HandlerFunc(NotFound)

	n.UseHandler(r)
	return n
}

func startServer() {
	http.Handle(config.ServerConfig.ContextPath, MakeHandler())
	l := log.New(os.Stdout, "[openapi] ", 0)

	if config.ServerConfig.Debug {
		nlog.DebugMode = config.ServerConfig.Debug
		nlog.Warningln("You are running openapi in debug mode.")
	}
	addr := fmt.Sprintf("%s:%d", config.ServerConfig.HTTPHost, config.ServerConfig.HTTPPort)
	l.Printf("listening on %s and serving on %s", addr, config.ServerConfig.ContextPath)
	if config.ServerConfig.HTTPSMode {
		l.Fatal(http.ListenAndServeTLS(addr, config.ServerConfig.HTTPSCert, config.ServerConfig.HTTPSKey, nil))
	}
	l.Fatal(http.ListenAndServe(addr, nil))
}
