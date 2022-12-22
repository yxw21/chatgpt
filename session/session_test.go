package chatgpt

import (
	"fmt"
	"testing"
)

func TestNewSessionWithCredential(t *testing.T) {
	// use credential
	session := NewSessionWithCredential("example@gmail.com", "123456", "I-1123123KASD").AutoRefresh()
	if session.IsInvalid() {
		if err := session.RefreshToken(); err != nil {
			t.Fatal(err)
		}
	}
	fmt.Println(session.GetAccessToken())
	fmt.Println(session.GetClearance())
	fmt.Println(session.GetUserAgent())
}

func TestNewSessionWithAccessToken(t *testing.T) {
	// use access token
	session := NewSessionWithAccessToken("jwt").AutoRefresh()
	if session.GetClearance() == "" {
		if err := session.RefreshClearance(); err != nil {
			t.Fatal(err)
		}
	}
	fmt.Println(session.GetClearance())
	fmt.Println(session.GetUserAgent())
}
