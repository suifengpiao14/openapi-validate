package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/nuveo/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "doc",
	Short: "doc is a tool validate api request and response struct",
	Long:  "doc is a tool validate api request and response struct",
}

// Config config struct
type Config struct {
	Adapter          string
	HTTPHost         string // HTTPHost Declare which http address the PREST used
	HTTPPort         int    // HTTPPort Declare which http port the PREST used
	ContextPath      string
	SSLMode          string
	SSLCert          string
	SSLKey           string
	SSLRootCert      string
	JWTKey           string
	JWTAlgo          string
	MigrationsPath   string
	DocAdapter       string
	CORSAllowOrigin  []string
	CORSAllowHeaders []string
	Debug            bool
	EnableDefaultJWT bool
	EnableCache      bool
	HTTPSMode        bool
	HTTPSCert        string
	HTTPSKey         string
}

var (
	// ServerConfig Config variable
	ServerConfig *Config

	configFile string

	defaultFile = "./config.toml"
)

func viperCfg() {
	configFile = getDefaultConf(os.Getenv("API_CONFIG"))

	dir, file := filepath.Split(configFile)
	file = strings.TrimSuffix(file, filepath.Ext(file))
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvPrefix("API")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(replacer)
	viper.AddConfigPath(dir)
	viper.SetConfigName(file)
	viper.SetConfigType("toml")
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", 3000)
	viper.SetDefault("ssl.mode", "disable")
	viper.SetDefault("debug", false)
	viper.SetDefault("jwt.default", true)
	viper.SetDefault("jwt.algo", "HS256")
	viper.SetDefault("cors.allowheaders", []string{"*"})
	viper.SetDefault("cache.enable", true)
	viper.SetDefault("context", "/")
	viper.SetDefault("https.mode", false)
	viper.SetDefault("https.cert", "/etc/certs/cert.crt")
	viper.SetDefault("https.key", "/etc/certs/cert.key")
	hDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)

	}
	viper.SetDefault("doc.location", filepath.Join(hDir, "doc"))
}

func getDefaultConf(config string) (cfg string) {
	cfg = config
	if config == "" {
		cfg = defaultFile
		_, err := os.Stat(cfg)
		if err != nil {
			cfg = ""
		}
	}
	return
}

// Parse pREST config
func Parse(cfg *Config) (err error) {
	err = viper.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			if configFile != "" {
				log.Fatal(fmt.Sprintf("File %s not found. Aborting.\n", configFile))
			}
		default:
			return
		}
	}
	cfg.HTTPHost = viper.GetString("http.host")
	cfg.HTTPPort = viper.GetInt("http.port")
	cfg.SSLMode = viper.GetString("ssl.mode")
	cfg.SSLCert = viper.GetString("ssl.cert")
	cfg.SSLKey = viper.GetString("ssl.key")
	cfg.SSLRootCert = viper.GetString("ssl.rootcert")
	cfg.JWTKey = viper.GetString("jwt.key")
	cfg.JWTAlgo = viper.GetString("jwt.algo")
	cfg.DocAdapter = viper.GetString("doc.adapter")
	cfg.CORSAllowOrigin = viper.GetStringSlice("cors.alloworigin")
	cfg.CORSAllowHeaders = viper.GetStringSlice("cors.allowheaders")
	cfg.Debug = viper.GetBool("debug")
	cfg.EnableDefaultJWT = viper.GetBool("jwt.default")
	cfg.EnableCache = viper.GetBool("cache.enable")
	cfg.ContextPath = viper.GetString("context")
	cfg.HTTPSMode = viper.GetBool("https.mode")
	cfg.HTTPSCert = viper.GetString("https.cert")
	cfg.HTTPSKey = viper.GetString("https.key")
	return
}

// Load configuration
func Load() {
	viperCfg()
	ServerConfig = &Config{}
	if err := Parse(ServerConfig); err != nil {
		panic(err)
	}
}
