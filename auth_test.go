package recaius

import (
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	id := os.Getenv("RECAIUS_ASR_ID")
	pass := os.Getenv("RECAIUS_ASR_PASS")

	if id == "" {
		t.Fatal("id required")
	}
	if pass == "" {
		t.Fatal("password required")
	}

	auth := Auth{
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

}
