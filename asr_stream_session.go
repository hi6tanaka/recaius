// asrResultChannel is based on
// https://github.com/eapache/channels.
// Thanks for a good reference!

package recaius

import "time"

type asrResultChannel struct {
	input  chan AsrResult
	output chan AsrResult
	length chan int
	buffer []AsrResult
}

func newAsrResultChannel() *asrResultChannel {
	ch := &asrResultChannel{
		input:  make(chan AsrResult),
		output: make(chan AsrResult),
		length: make(chan int),
		buffer: nil,
	}
	go ch.loop()
	return ch
}

func (a *asrResultChannel) In() chan<- AsrResult {
	return a.input
}

func (a *asrResultChannel) Out() <-chan AsrResult {
	return a.output
}

func (a *asrResultChannel) Len() int {
	return <-a.length
}

func (a *asrResultChannel) Close() {
	if a.input != nil {
		close(a.input)
		a.input = nil
	}
}

func (a *asrResultChannel) ClosedIn() bool {
	return a.input == nil
}

func (a *asrResultChannel) ClosedOut() bool {
	return a.output == nil
}

func (a *asrResultChannel) loop() {
	var i, o chan AsrResult
	var n AsrResult

	i = a.input

	for i != nil || o != nil {
		select {
		case e, open := <-i:
			if open {
				a.buffer = append(a.buffer, e)
				if len(a.buffer) > 0 {
					n = a.buffer[0]
					o = a.output
				}
			} else {
				i = nil
			}
		case o <- n:
			a.buffer = a.buffer[1:]
			if len(a.buffer) > 0 {
				n = a.buffer[0]
			} else {
				o = nil
			}
		case a.length <- len(a.buffer):
		}

	}
	close(a.output)
	close(a.length)
	a.output, a.length = nil, nil
}

type AsrStreamSession struct {
	conn *asrConnection
	ch   *asrResultChannel
}

func newAsrStreamSession(conn *asrConnection) *AsrStreamSession {
	return &AsrStreamSession{
		conn: conn,
		ch:   newAsrResultChannel(),
	}
}

func (sess *AsrStreamSession) Response() <-chan AsrResult {
	return sess.ch.Out()
}

func (sess *AsrStreamSession) StartWatch() {
	if sess.ch.ClosedIn() {
		return
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rs, err := sess.conn.AskResult()
			if err != nil {
				sess.ch.In() <- AsrResult{Err: err}
				return
			}
			sess.emitResults(rs)
			if sess.ch.ClosedIn() {
				return
			}
		}
	}
}

func (sess *AsrStreamSession) Send(data []byte) {
	rs, err := sess.conn.Send(data)
	if err != nil {
		sess.ch.In() <- AsrResult{Err: err}
		return
	}
	sess.emitResults(rs)
	return
}

func (sess *AsrStreamSession) Flush() {
	rs, err := sess.conn.Flush()
	if err != nil {
		sess.ch.In() <- AsrResult{Err: err}
		return
	}
	sess.emitResults(rs)
	return
}

func (sess *AsrStreamSession) Close() {
	sess.ch.Close()
	sess.conn.Close()
}

func (sess *AsrStreamSession) emitResults(rs []AsrResult) {
	ch := sess.ch.In()
	for _, r := range rs {
		if r.Type == "NO_DATA" {
			sess.ch.Close()
			return
		}
		ch <- r
	}
}
