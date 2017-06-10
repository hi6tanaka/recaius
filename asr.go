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
	MaxConnection   int64  `json:"-"`
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

type semaphore chan struct{}

type Asr struct {
	auth    *Auth
	config  *AsrConfig
	conns   []*asrConnection
	connSem semaphore
}

func NewAsr(auth *Auth) *Asr {
	return NewAsrWithConfig(auth, &AsrConfig{ModelID: 1})
}

func (a *Asr) Close() {
	for _, c := range a.conns {
		c.Close()
	}
	close(a.connSem)
}

func NewAsrWithConfig(auth *Auth, config *AsrConfig) *Asr {
	if config.MaxConnection == 0 {
		config.MaxConnection = 5
	}
	return &Asr{
		auth:    auth,
		config:  config,
		conns:   nil,
		connSem: make(semaphore, config.MaxConnection),
	}
}

// find free connection, or makae new connection
func (a *Asr) Session() (*asrSession, error) {
	conn, err := a.newConnection()
	if err != nil {
		return nil, err
	}
	return &asrSession{conn: conn}, nil
}

// find free connection, or make new connection
func (a *Asr) Stream() (*asrStreamSession, error) {
	conn, err := a.newConnection()
	if err != nil {
		return nil, err
	}
	return newAsrStreamSession(conn), nil
}

// TODO: impl with sound data
func (a *Asr) Recognize(data []byte) error {
	return nil
}

// TODO: impl with sound file
func (a *Asr) RecognizeFile(path string) error {
	return nil
}

func (a *Asr) newConnection() (*asrConnection, error) {
	// fmt.Println("to wait:", len(a.connSem))
	a.connSem <- struct{}{}
	// fmt.Println("go through:", len(a.connSem))
	conn, err := newAsrConnection(a.auth, a.config, func(conn *asrConnection) {
		// fmt.Println("to restore:", len(a.connSem))
		<-a.connSem
		// fmt.Println("restore:", len(a.connSem))
		return
	})
	if err != nil {
		return conn, err
	}
	return conn, nil
}
