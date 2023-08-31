package notify

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gabrielmusskopf/merminder/logger"
)

type MergeRequestInfo struct {
	Title               string
	CreatedAt           time.Time
	TotalDiscussions    int
	OpenDiscussions     int
	TimeOlderDiscussion *time.Time
	URL                 string
}

type Notifier interface {
	Notify([]MergeRequestInfo)
}

func Send(url string, body []byte) error {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buff := make([]byte, 2048)
	_, err = resp.Body.Read(buff)
	if err != nil {
        return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("An error occur: %s %s", resp.Status, string(buff))
	}

	logger.Info("%s posted status to configured webhook", resp.Status)

	return nil
}
