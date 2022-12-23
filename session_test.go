package chatgpt

import (
	"fmt"
	"testing"
)

func TestNewSessionWithCredential(t *testing.T) {
	retry := 3
	// use credential
	session := NewSessionWithCredential("example@gmail.com", "password", "I-ASDA123ASDA").AutoRefresh()
	if session.AccessTokenIsInvalid() {
		for i := 0; i < retry; i++ {
			err := session.RefreshToken()
			if err == nil {
				fmt.Println(session.GetAccessToken())
				fmt.Println(session.GetClearance())
				fmt.Println(session.GetUserAgent())
				break
			}
			if i == retry-1 {
				t.Fatal(err)
			}
		}
	}
}

func TestNewSessionWithAccessToken(t *testing.T) {
	retry := 3
	// use access token
	session := NewSessionWithAccessToken("jwt").AutoRefresh()
	if session.ClearanceIsInValid() {
		for i := 0; i < retry; i++ {
			err := session.RefreshClearance()
			if err == nil {
				fmt.Println(session.GetClearance())
				fmt.Println(session.GetUserAgent())
				break
			}
			if i == retry-1 {
				t.Fatal(err)
			}
		}
	}
}
