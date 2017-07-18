package recaius

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func TestRecognize(t *testing.T) {
	auth := login(t)
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

func TestParallel(t *testing.T) {
	auth := login(t)
	defer func() {
		if err := auth.Logout(); err != nil {
			t.Fatal("logout failed:", err)
		}
	}()
	asr := NewAsr(auth)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		ii := i
		fmt.Println("Parallel:", ii)
		wg.Add(1)
		go func() {
			sess, err := asr.Session()
			fmt.Println("Got session:", ii)
			if err != nil {
				t.Fatal("create session error:", err)
			}
			defer sess.Close()
			fmt.Println("Reading wav...", ii)
			for data := range readWav("sample.wav", t) {
				// fmt.Printf("Sending(%d)...", ii)
				if err := sess.Send(data); err != nil {
					t.Fatal("Send error:", err)
				}
			}
			fmt.Println("Waiting result:", ii)
			results, err := sess.FlushWait()
			// _, err = sess.FlushWait()
			if err != nil {
				t.Fatal("FlushWait error:", err)
			}
			for _, r := range results {
				fmt.Println(r.OneBest.Type, r.OneBest.Str)
			}
			fmt.Println("Done:", ii)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestAsyncRecognize(t *testing.T) {
	auth := login(t)
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

		sess.Send(data[i:j])
	}
	sess.Flush()
	sess.StartWatch()
}

func TestNBest(t *testing.T) {
	auth := login(t)
	defer func() {
		if err := auth.Logout(); err != nil {
			t.Fatal("logout failed:", err)
		}
	}()

	asr := NewAsrWithConfig(auth, &AsrConfig{
		ModelID:    1,
		ResultType: "nbest",
	})
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
		fmt.Println(r.NBest.Type, r.NBest)
	}
}

func login(t *testing.T) *Auth {
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
		AutoLogin:     true,
	}
	if err := auth.Login(); err != nil {
		t.Fatal("login failed:", err)
	}
	return auth
}

func readWav(path string, t *testing.T) <-chan []byte {
	data, err := ioutil.ReadFile(path) // You need to prepare
	if err != nil {
		t.Fatal("file open error: sample.wav:", err)
	}
	sampleOffset := 44
	ch := make(chan []byte)
	go func() {
		defer func() { close(ch) }()
		for i := sampleOffset; i < len(data); i += 32000 {
			j := i + 32000
			if j > len(data) {
				j = len(data)
			}
			ch <- data[i:j]
		}
	}()
	return ch
}
