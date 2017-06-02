package recaius

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// authentication
type ServiceInfo struct {
	ServiceId string `json:"service_id",omitempty`
	Password  string `json:"password",omitempty`
}

type Auth struct {
	SpeechRecogJa *ServiceInfo `json:"speech_recog_jaJP,omitempty"`
	SpeechRecogEn *ServiceInfo `json:"speech_recog_enUS,omitempty"`
	SpeechRecogZh *ServiceInfo `json:"speech_recog_zhCH,omitempty"`
	ExpirySec     int64        `json:"expiry_sec,omitempty"`
	Token         string       `json:"-"`
}

type ResponseToken struct {
	Token     string `json:"token"`
	ExpirySec int64  `json:"expiry_sec"`
}

/// TODO: check if token already taken
func (a *Auth) Login() error {
	body, err := json.Marshal(a)
	if err != nil {
		return err
	}
	resp, err := http.Post(tokenURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		var rt ResponseToken
		if err := json.NewDecoder(resp.Body).Decode(&rt); err != nil {
			return err
		}
		a.Token = rt.Token
		a.ExpirySec = rt.ExpirySec
	} else if resp.StatusCode >= 400 {
		var rt ResponseError
		if err := json.NewDecoder(resp.Body).Decode(&rt); err != nil {
			return err
		}
		return rt
	}
	return nil
}

func (a *Auth) Logined() bool {
	return a.Token != ""
}

/// TODO: impl
func (a *Auth) Extend() error {
	return nil
}

/// TODO: impl
func (a *Auth) Info() {

}

func (a *Auth) Logout() error {
	if a.Logined() {
		client := http.Client{}
		req, err := makeAuthorizedRequest(&client, a, "DELETE", tokenURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			var rt ResponseError
			if err := json.NewDecoder(resp.Body).Decode(&rt); err != nil {
				return err
			}
			return rt
		}
	}
	return nil
}

// util

func makeAuthorizedRequest(client *http.Client, auth *Auth, method string, url string, body io.Reader) (*http.Request, error) {
	if !auth.Logined() {
		return nil, fmt.Errorf("You need login first")
	}
	req, err := http.NewRequest(method, url, body)
	if err == nil {
		req.Header.Add("X-Token", auth.Token)
	}
	return req, err
}
