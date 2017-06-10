package recaius

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
	AutoLogin     bool         `json:"-"`
	expireAt      time.Time    `json:"-"`
	token         string       `json:"-"`
}

type ResponseToken struct {
	Token     string `json:"token"`
	ExpirySec int64  `json:"expiry_sec"`
}

/// TODO: check if token already taken
func (a *Auth) Login() error {
	if a.ExpirySec < 0 {
		a.ExpirySec = 3600
	} else if a.ExpirySec < 600 {
		a.ExpirySec = 600
	}
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
		return a.setToken(resp.Body)
	} else if resp.StatusCode >= 400 {
		return a.errorResponse(resp.Body)
	}
	return fmt.Errorf("server error: code=%d", resp.StatusCode)
}

func (a *Auth) Logined() bool {
	return a.token != ""
}

// Login If not logined yet
func (a *Auth) Extend() error {
	if !a.Logined() {
		return a.Login()
	}
	client := http.Client{}
	req, err := makeTokenRequest("PUT", tokenURL, a.token, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 300 {
		return a.setToken(resp.Body)
	} else if resp.StatusCode >= 400 {
		// re-login attempt
		return a.Login()
	}
	return fmt.Errorf("server error: code=%d", resp.StatusCode)
}

func (a *Auth) Token() (string, error) {
	if a.token == "" {
		if a.AutoLogin {
			err := a.Extend()
			return a.token, err
		}
		return a.token, fmt.Errorf("need login")
	} else if time.Now().Add(5 * time.Minute).After(a.expireAt) {
		if a.AutoLogin {
			err := a.Extend()
			return a.token, err
		}
		return a.token, nil // it's caller's responsibility
	}
	return a.token, nil
}

func (a *Auth) Logout() error {
	if a.Logined() {
		client := http.Client{}
		req, err := makeTokenRequest("DELETE", tokenURL, a.token, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return a.errorResponse(resp.Body)
		}
	}
	return nil
}

func (a *Auth) errorResponse(b io.Reader) error {
	var rt ResponseError
	if err := json.NewDecoder(b).Decode(&rt); err != nil {
		return err
	}
	return rt
}

func (a *Auth) setToken(b io.Reader) error {
	var rt ResponseToken
	if err := json.NewDecoder(b).Decode(&rt); err != nil {
		return err
	}
	a.token = rt.Token
	a.expireAt = time.Now().Add(time.Duration(rt.ExpirySec) * time.Second)
	return nil
}

// util

func makeTokenRequest(method string, url string, token string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err == nil {
		req.Header.Add("X-Token", token)
	}
	return req, err
}

func (a *Auth) MakeAuthorizedRequest(method string, url string, body io.Reader) (*http.Request, error) {
	token, err := a.Token()
	if err != nil {
		return nil, err
	}
	return makeTokenRequest(method, url, token, body)
}
