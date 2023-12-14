package internal

import (
	"log/slog"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/labels"

	defaults "github.com/mcuadros/go-defaults"
)

// NewConfig reads ciux config file to buld a Config struct
// it uses repositoryPath if not null or current directory
func NewConfig(repositoryPath string) (ProjConfig, error) {
	var configPath = repositoryPath
	config := new(ProjConfig)
	var err error

	newviper := viper.New()
	newviper.AddConfigPath(configPath)
	newviper.SetConfigType("yaml")
	newviper.SetConfigName(".ciux")

	err = newviper.ReadInConfig()
	if err != nil {
		return *config, err
	}
	slog.Debug("Ciux config file", "file", newviper.ConfigFileUsed())

	slog.Debug("Set defaults")
	defaults.SetDefaults(config)
	err = mapstructure.Decode(newviper.AllSettings(), config)
	if err != nil {
		return *config, err
	}
	return *config, nil
}

type DepConfig struct {
	Url     string     `mapstructure:"url" default:""`
	Clone   bool       `mapstructure:"clone" default:"false"`
	Image   string     `mapstructure:"image" default:""`
	Pull    bool       `mapstructure:"pull" default:"false"`
	Package string     `mapstructure:"package" default:""`
	Labels  labels.Set `mapstructure:"labels"`
}

type ProjConfig struct {
	Registry     string      `mapstructure:"registry" default:""`
	Dependencies []DepConfig `mapstructure:"dependencies"`
}
