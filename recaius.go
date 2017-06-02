package recaius

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ResponseError struct {
	Code     int64  `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
}

/// ResponseError has an error interface
func (e ResponseError) Error() string {
	return fmt.Sprintf("Error code: %d message: \"%s\" more_info: \"%s\"", e.Code, e.Message, e.MoreInfo)
}

// You must Close response if not nil
func callApi(auth *Auth, method string, url string, body io.Reader, contentType string) (*http.Response, error) {
	client := &http.Client{}
	req, err := makeAuthorizedRequest(client, auth, method, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var rt ResponseError
		if err := json.NewDecoder(resp.Body).Decode(&rt); err != nil {
			return nil, err
		}
		return nil, rt
	}

	return resp, nil
}
