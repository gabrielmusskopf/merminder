package main

import (
	"time"

	"github.com/gabrielmusskopf/merminder/config"
	"github.com/gabrielmusskopf/merminder/logger"
	"github.com/gabrielmusskopf/merminder/notify"
	"github.com/gabrielmusskopf/merminder/service"
	"github.com/go-co-op/gocron"
	"github.com/xanzy/go-gitlab"
)

func main() {

	config := config.ReadConfig()
	config.LogInfo()

	var opt gitlab.ClientOptionFunc
	if !config.DefaultHost() {
		opt = gitlab.WithBaseURL(config.Repository.Host)
	}

	git, err := gitlab.NewClient(config.Repository.Token, opt)
	if err != nil {
		logger.Fatal(err)
	}

	service := &service.Service{
		Notifier: *notify.NewNotifier(config.Send.WebhookURL),
		Git:      git,
	}

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		logger.Fatal(err)
	}

	s := gocron.NewScheduler(location)

	if config.Observe.Every != "" {
		logger.Info("starting merminder with %s update frequency", config.Observe.Every)
		s.Every(config.Observe.Every)

	} else if len(config.Observe.At) != 0 {
		for _, at := range config.Observe.At {
			logger.Info("starting merminder with update scheluded to %s", at)
			s.Every(1).Day().At(at)
		}

	} else {
		logger.Fatals("frequency time is missing. Either configure 'every' or 'at'")
	}

	_, err = s.Do(func() {
		service.FetchMergeRequests()
	})
	if err != nil {
		logger.Fatal(err)
	}

	s.StartAsync()
	select {}
}
