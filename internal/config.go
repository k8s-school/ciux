package internal

import (
	"os"

	"github.com/k8s-school/ciux/log"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	defaults "github.com/mcuadros/go-defaults"
)

// ReadConfig reads ciux config file
// it uses repositoryPath if not null or current directory
func ReadConfig(repositoryPath string) Config {
	var configPath string
	var err error
	if len(repositoryPath) == 0 {
		configPath, err = os.Getwd()
		cobra.CheckErr(err)
	} else {
		configPath = repositoryPath
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".ciux")

	err = viper.ReadInConfig()
	cobra.CheckErr(err)
	log.Debugf("Use config file: %s", viper.ConfigFileUsed())

	c := new(Config)
	defaults.SetDefaults(c)
	err = mapstructure.Decode(viper.AllSettings(), c)
	cobra.CheckErr(err)
	return *c
}

type Dependency struct {
	Url   string `mapstructure:"url" default:""`
	Clone bool   `mapstructure:"clone" default:"false"`
	Pull  bool   `mapstructure:"pull" default:"false"`
}

type Config struct {
	Registry     string       `mapstructure:"registry" default:""`
	Dependencies []Dependency `mapstructure:"dependencies"`
}
