package ntfy

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	baseUri = "https://ntfy.sh"
	topic   = "xbd_au"
)

var target, _ = url.JoinPath(baseUri, topic)

func Notify(body string) error {
	client := &http.Client{}

	r := strings.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, target, r)
	if err != nil {
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
