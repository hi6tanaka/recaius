package recaius

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestRecognize(t *testing.T) {
	id := os.Getenv("RECAIUS_ASR_ID")
	pass := os.Getenv("RECAIUS_ASR_PASS")

	if id == "" {
		t.Fatal("id required")
	}
	if pass == "" {
		t.Fatal("password required")
	}

	auth := &Auth{
		SpeechRecogJa: &ServiceInfo{id, pass},
	}
	if err := auth.Login(); err != nil {
		t.Fatal("login failed:", err)
	}
	defer func() {
		if err := auth.Logout(); err != nil {
			t.Fatal("logout failed:", err)
		}
	}()

	asr := NewAsr(auth)
	sess, err := asr.Session()
	if err != nil {
		t.Fatal("create session error:", err)
	}
	defer sess.Close()

	data, err := ioutil.ReadFile("sample.wav") // You need to prepare
	if err != nil {
		t.Fatal("file open error: sample.wav:", err)
	}
	sampleOffset := 44
	fmt.Println("Start sending:", sampleOffset, len(data))

	for i := sampleOffset; i < len(data); i += 32000 {
		j := i + 32000
		if j > len(data) {
			j = len(data)
		}
		fmt.Printf("Sending: [%d:%d]\n", i, j)

		if err := sess.Send(data[i:j]); err != nil {
			t.Fatal("Send error:", err)
		}
	}
	fmt.Printf("Waiting")
	results, err := sess.FlushWait()
	if err != nil {
		t.Fatal("FlushWait error:", err)
	}
	for _, r := range results {
		fmt.Println(r.OneBest.Type, r.OneBest.Str)
	}
}

func TestAsyncRecognize(t *testing.T) {
	id := os.Getenv("RECAIUS_ASR_ID")
	pass := os.Getenv("RECAIUS_ASR_PASS")

	if id == "" {
		t.Fatal("id required")
	}
	if pass == "" {
		t.Fatal("password required")
	}

	auth := &Auth{
		SpeechRecogJa: &ServiceInfo{id, pass},
	}
	if err := auth.Login(); err != nil {
		t.Fatal("login failed:", err)
	}
	defer func() {
		if err := auth.Logout(); err != nil {
			t.Fatal("logout failed:", err)
		}
	}()

	asr := NewAsr(auth)
	sess, err := asr.Stream()
	if err != nil {
		t.Fatal("create session error:", err)
	}
	defer sess.Close()

	data, err := ioutil.ReadFile("sample.wav") // You need to prepare
	if err != nil {
		t.Fatal("file open error: sample.wav:", err)
	}
	sampleOffset := 44
	fmt.Println("Start sending:", sampleOffset, len(data))

	go func() {
		ch := sess.Response()
		for {
			select {
			case r, open := <-ch:
				if !open {
					return
				}
				if r.Err != nil {
					t.Fatal("Error:", r.Err)
				}
				fmt.Println("Response:", r.OneBest.Type, r.OneBest.Str)
			}
		}
	}()

	for i := sampleOffset; i < len(data); i += 32000 {
		j := i + 32000
		if j > len(data) {
			j = len(data)
		}
		fmt.Printf("Sending: [%d:%d]\n", i, j)

		sess.Send(data[i:j])
	}
	sess.Flush()
	sess.StartWatch()
}
