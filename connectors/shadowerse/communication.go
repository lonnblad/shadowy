package shadowerse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendShadeRequest(port string, shade ShadeRequest) error {
	url := fmt.Sprintf("http://127.0.0.1:%s/shade_requests", port)
	return sendToShadowerse(url, shade)
}

func SendShadeResponse(port string, shade ShadeResponse) error {
	url := fmt.Sprintf("http://127.0.0.1:%s/shade_responses", port)
	return sendToShadowerse(url, shade)
}

func sendToShadowerse(url string, shade interface{}) error {
	body, err := json.Marshal(shade)
	if err != nil {
		return fmt.Errorf("couldn't marshal shade: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("couldn't create a new request: %w", err)
	}

	client := new(http.Client)

	if _, err := client.Do(req); err != nil {
		return fmt.Errorf("couldn't execute the request: %w", err)
	}

	return nil
}
