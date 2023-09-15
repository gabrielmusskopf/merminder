package merminder

import (
	"bytes"
	"fmt"
	"net/http"
)

type Notifier struct {
	Url string
}

func NewNotifier(url string) *Notifier {
	return &Notifier{
		Url: url,
	}
}

func (n *Notifier) Notify(body string) error {
	resp, err := http.Post(n.Url, "application/json", bytes.NewBuffer([]byte(body)))
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

	Info("%s posted status to configured webhook", resp.Status)

	return nil
}
