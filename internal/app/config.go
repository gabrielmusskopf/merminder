package merminder

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repository struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
	}
	Send struct {
		Notification     string `yaml:"notification"`
		TemplateFilePath string `yaml:"templateFilePath"`
		WebhookURL       string `yaml:"webhookUrl"`
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
		Fatals("trying to get config but is not setted yet")
	}
	return config
}

func ReadConfig() *Config {
	f, err := os.Open(".merminder.yml")
    if err != nil {
        f, err = os.Open(".merminder.yaml")
        if err != nil {
            Fatals("config file .merminder.yml or .merminder.yaml was not found")
        }
    }
	defer f.Close()

	config = &Config{}

	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&config); err != nil {
		Fatal(err)
	}

	if config.Repository.Token == "" {
		Fatals("token is missing")
	}

	if config.Observe.Every != "" && len(config.Observe.At) != 0 {
		Warning("cannot use 'observe.at' and 'obser.every' at the same time")
		Warning("only 'observe.every' will be considered")
		config.Observe.At = make([]string, 0)
	} else {
		Fatals("at least one observe frequency must be set: 'every' or 'at'")
	}

	if config.Send.TemplateFilePath == "" {
		Info("template file path not set, using default instead")
		config.Send.TemplateFilePath = "default.tmpl"
	}

	return config
}

func (c *Config) LogInfo() {
	Info("repository url: %s", c.Repository.Host)
    if c.Send.Notification != "off" {
        Info("webhook url: %s", c.Send.WebhookURL)
        Info("observed groups: %v", c.Observe.Groups)
        Info("observed projects: %v", c.Observe.Projects)
        Info("template file path: %v", c.Send.TemplateFilePath)
    } else {
        Info("notification send disabled: %s", c.Send.Notification)
    }
	if c.Observe.Every != "" {
		Info("every: %s", c.Observe.Every)
	} else if len(c.Observe.At) != 0 {
		Info("at: %s", c.Observe.At)
	}
}

func (c *Config) DefaultHost() bool {
	return c.Repository.Host == ""
}

func (c *Config) NotificationEnabled() bool {
    return c.Send.Notification != "off"
}
