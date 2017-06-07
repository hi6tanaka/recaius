package recaius

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
)

type asrConnection struct {
	ID      string
	auth    *Auth
	config  *AsrConfig
	voiceID int64
	busy    bool
}

func newAsrConnection(auth *Auth, config *AsrConfig) (*asrConnection, error) {
	url := fmt.Sprintf("%s/voices", asrURL)
	payload, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	resp, err := callApi(auth, "POST", url, bytes.NewReader(payload), "application/json")
	if err != nil {
		return nil, err
	}

	type rt struct{ UUID string }
	var t rt
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &asrConnection{
		ID:      t.UUID,
		auth:    auth,
		config:  config,
		voiceID: 1,
		busy:    true,
	}, nil
}

func (conn *asrConnection) urlSend() string {
	return fmt.Sprintf("%s/voices/%s", asrURL, conn.ID)
}
func (conn *asrConnection) urlFlush() string {
	return fmt.Sprintf("%s/voices/%s/flush", asrURL, conn.ID)
}
func (conn *asrConnection) urlResults() string {
	return fmt.Sprintf("%s/voices/%s/results", asrURL, conn.ID)
}
func (conn *asrConnection) urlDelete() string {
	return fmt.Sprintf("%s/voices/%s", asrURL, conn.ID)
}

func (conn *asrConnection) Send(buf []byte) ([]AsrResult, error) {
	var data bytes.Buffer
	w := multipart.NewWriter(&data)
	fw, err := w.CreateFormField("voice_id")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write([]byte(strconv.FormatInt(conn.voiceID, 10))); err != nil {
		return nil, err
	}
	fw, err = w.CreateFormField("voice")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(buf); err != nil {
		return nil, err
	}
	w.Close()

	resp, err := callApi(conn.auth, "PUT", conn.urlSend(), &data, w.FormDataContentType())
	if err != nil {
		return nil, err
	}
	conn.voiceID += 1
	defer resp.Body.Close()
	return conn.checkResponse(resp)
}

func (conn *asrConnection) Flush() ([]AsrResult, error) {
	data, err := json.Marshal(&asrFlushPayload{conn.voiceID})
	if err != nil {
		return nil, err
	}
	resp, err := callApi(conn.auth, "PUT", conn.urlFlush(), bytes.NewReader(data), "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return conn.checkResponse(resp)
}

func (conn *asrConnection) AskResult() ([]AsrResult, error) {
	resp, err := callApi(conn.auth, "GET", conn.urlResults(), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return conn.checkResponse(resp)
}

// TODO: support confnet
// TODO: nbest is currently interface{}.
func (conn *asrConnection) checkResponse(resp *http.Response) ([]AsrResult, error) {
	var rs []AsrResult
	if resp.StatusCode == 200 {
		resultType := conn.config.ResultType
		if resultType == "" || resultType == "one_best" {
			var r [][2]string
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				return nil, err
			}
			for _, x := range r {
				item := AsrOneBest{Type: x[0], Str: x[1]}
				rs = append(rs, AsrResult{Type: item.Type, OneBest: item})
			}
		} else if resultType == "nbest" {
			var r []AsrNBest
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				return nil, err
			}
			for _, x := range r {
				rs = append(rs, AsrResult{Type: x.Type, NBest: x})
			}
		} else {
			return nil, fmt.Errorf("result_type: %s is not supported", resultType)
		}
	}
	return rs, nil
}

func (conn *asrConnection) Close() {
	if conn.ID == "" {
		return
	}
	callApi(conn.auth, "DELETE", conn.urlDelete(), nil, "")
}
