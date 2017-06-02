package recaius

import "time"

type asrSession struct {
	conn     *asrConnection
	results  []asrResult
	buffered bool // 結果が残っている（かもしれない）
}

func (sess *asrSession) Send(data []byte) error {
	rs, err := sess.conn.Send(data)
	if err != nil {
		return err
	}
	sess.buffered = true
	sess.storeResults(rs)
	return nil
}

func (sess *asrSession) Flush() error {
	rs, err := sess.conn.Flush()
	if err != nil {
		return err
	}
	sess.storeResults(rs)
	return nil
}

func (sess *asrSession) Close() {
	sess.conn.Close()
}

func (sess *asrSession) Wait() ([]asrResult, error) {
	if !sess.buffered {
		return sess.results, nil
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rs, err := sess.conn.AskResult()
			if err != nil {
				return nil, err
			}
			sess.storeResults(rs)
			if !sess.buffered {
				return sess.results, nil
			}
		}
	}
}

func (sess *asrSession) FlushWait() ([]asrResult, error) {
	if err := sess.Flush(); err != nil {
		return nil, err
	}
	return sess.Wait()
}

func (sess *asrSession) storeResults(rs []asrResult) {
	for _, r := range rs {
		sess.results = append(sess.results, r)
		if r.Type == "NO_DATA" {
			sess.buffered = false
		}
	}
}
