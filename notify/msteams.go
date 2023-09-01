package notify

import (
	"github.com/gabrielmusskopf/merminder/logger"
)

type MSTeamsNotifier struct {
	Url string
}

func NewTeamsNotifier(url string) *MSTeamsNotifier {
	return &MSTeamsNotifier{
		Url: url,
	}
}

func (ms MSTeamsNotifier) Notify(b string) {
	if err := Send(ms.Url, []byte(b)); err != nil {
		logger.Error(err)
	}
}
