package config

import (
	"os"

	"github.com/gabrielmusskopf/merminder/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Repository struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
	}
	Send struct {
		WebhookURL string `yaml:"webhookUrl"`
	}
	Observe struct {
		Groups   []int    `yaml:",flow"`
		Projects []int    `yaml:",flow"`
		Every    string   `yaml:"every"`
		At       []string `yaml:",flow"`
	}
}

var config *Config

func GetConfig() *Config {
    if config == nil {
        logger.Fatals("trying to get config but is not setted yet")
    }
    return config
}

func ReadConfig() *Config {
	f, err := os.Open(".merminder.yml")
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	config = &Config{}

	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&config); err != nil {
		logger.Fatal(err)
	}

	if config.Repository.Token == "" {
		logger.Fatals("token is missing")
	}

    if config.Observe.Every != "" && len(config.Observe.At) != 0 {
        logger.Warning("cannot use 'observe.at' and 'obser.every' at the same time")
        logger.Warning("only 'observe.every' will be considered")
        config.Observe.At = make([]string, 0)
    } else {
        logger.Fatals("at least one observe frequency must be set: 'every' or 'at'")
    }

	return config
}

func (c *Config) LogInfo() {
    logger.Info("repository url: %s", c.Repository.Host)
    logger.Info("webhook url: %s", c.Send.WebhookURL)
    logger.Info("observed groups: %v", c.Observe.Groups)
    logger.Info("observed projects: %v", c.Observe.Projects)
    if c.Observe.Every != "" {
        logger.Info("every: %s", c.Observe.Every)
    } else if len(c.Observe.At) != 0 {
        logger.Info("at: %s", c.Observe.At)
    }
}

func (c *Config) DefaultHost() bool {
	return c.Repository.Host == ""
}

