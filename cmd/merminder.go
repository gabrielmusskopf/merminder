package main

import (
	merminder "github.com/gabrielmusskopf/merminder/internal/app"
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

	service.Start()
}
