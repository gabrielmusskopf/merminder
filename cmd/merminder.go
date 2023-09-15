package main

import (
	"time"

	merminder "github.com/gabrielmusskopf/merminder/internal/app"
	"github.com/go-co-op/gocron"
	"github.com/xanzy/go-gitlab"
)

func main() {

	config := merminder.ReadConfig()
	config.LogInfo()

	var opt gitlab.ClientOptionFunc
	if !config.DefaultHost() {
		opt = gitlab.WithBaseURL(config.Repository.Host)
	}

	git, err := gitlab.NewClient(config.Repository.Token, opt)
	if err != nil {
		merminder.Fatal(err)
	}

	service := &merminder.Service{
		Notifier: *merminder.NewNotifier(config.Send.WebhookURL),
		Git:      git,
	}

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		merminder.Fatal(err)
	}

	s := gocron.NewScheduler(location)

	if config.Observe.Every != "" {
		merminder.Info("starting merminder with %s update frequency", config.Observe.Every)
		s.Every(config.Observe.Every)

	} else if len(config.Observe.At) != 0 {
		for _, at := range config.Observe.At {
			merminder.Info("starting merminder with update scheluded to %s", at)
			s.Every(1).Day().At(at)
		}

	} else {
		merminder.Fatals("frequency time is missing. Either configure 'every' or 'at'")
	}

	_, err = s.Do(func() {
		service.FetchMergeRequests()
	})
	if err != nil {
		merminder.Fatal(err)
	}

	s.StartAsync()
	select {}
}
