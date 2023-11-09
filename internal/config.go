package internal

import (
	"os"

	"github.com/k8s-school/ciux/log"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	defaults "github.com/mcuadros/go-defaults"
)

// NewConfig reads ciux config file to buld a Config struct
// it uses repositoryPath if not null or current directory
func NewConfig(repositoryPath string) (Config, error) {
	var configPath string
	config := new(Config)
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
	if err != nil {
		return *config, err
	}
	log.Debugf("Use config file: %s", viper.ConfigFileUsed())

	defaults.SetDefaults(config)
	err = mapstructure.Decode(viper.AllSettings(), config)
	if err != nil {
		return *config, err
	}
	return *config, nil
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

// WriteOutConfig writes out the shell configuration file
// used be the CI/CD pipeline
func (c *Config) WriteOutConfig() []Dependency {
	return c.Dependencies
}
