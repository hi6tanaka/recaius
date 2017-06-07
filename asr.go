package recaius

type AsrConfig struct {
	AudioType       string `json:"audio_type,omitempty"`
	EnergyThreshold int64  `json:"energy_threshold,omitempty"`
	ResultType      string `json:"result_type,omitempty"`
	ResultCount     int64  `json:"result_count,omitempty"`
	ModelID         int64  `json:"model_id"`
	PhshToTalk      bool   `json:"phsh_to_talk,omitempty"`
	DataLog         int64  `json:"data_log,omitempty"`
	Comment         string `json:"comment,omitempty"`
	Retry           bool   `json:"-"`
	MaxRetry        int64  `json:"-"`
	PollingInterval int64  `json:"-"` // millisecond
}

type asrFlushPayload struct {
	VoiceID int64 `json:"voice_id"`
}

type AsrOneBest struct {
	Type string
	Str  string
}

type AsrNBestElement struct {
	Str        string
	Confidence float64
	Yomi       string
	Begin      int64
	End        int64
}

type AsrNBest struct {
	Type       string
	Status     string
	ResultTemp string
	Result     [][]AsrNBestElement
	ResultRaw  interface{} `json:"result"`
}

type AsrResult struct {
	Type    string
	Err     error
	OneBest AsrOneBest
	NBest   AsrNBest
	// ConfNet *asrConfNet
}

type asr struct {
	auth   *Auth
	config *AsrConfig
	conns  []asrConnection
}

func NewAsr(auth *Auth) *asr {
	return &asr{
		auth:   auth,
		config: &AsrConfig{ModelID: 1},
		conns:  []asrConnection{},
	}
}

func (a *asr) Close() {
	for _, c := range a.conns {
		c.Close()
	}
}

func NewAsrWithConfig(auth *Auth, config *AsrConfig) *asr {
	return &asr{
		auth:   auth,
		config: config,
		conns:  []asrConnection{},
	}
}

// find free connection, or make new connection
func (a *asr) Session() (*asrSession, error) {
	conn, err := a.newConnection()
	if err != nil {
		return nil, err
	}
	return &asrSession{conn: conn}, nil
}

// find free connection, or make new connection
func (a *asr) Stream() (*asrStreamSession, error) {
	conn, err := a.newConnection()
	if err != nil {
		return nil, err
	}
	return newAsrStreamSession(conn), nil
}

// TODO: impl with sound data
func (a *asr) Recognize(data []byte) error {
	return nil
}

// TODO: impl with sound file
func (a *asr) RecognizeFile(path string) error {
	return nil
}

func (a *asr) newConnection() (*asrConnection, error) {
	return newAsrConnection(a.auth, a.config)
	// return &a.conns[0], nil
}
