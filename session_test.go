package chatgpt

import (
	"fmt"
	"testing"
)

func TestNewSessionWithCredential(t *testing.T) {
	retry := 3
	browser, closeBrowser, err := NewBrowser("I-ABCDEFGHIJK")
	if err != nil {
		t.Fatal(err)
	}
	defer closeBrowser()
	// use credential
	session := NewSessionWithCredential(browser, "example@gmail.com", "password").AutoRefresh()
	if session.AccessTokenIsInvalid() {
		for i := 0; i < retry; i++ {
			err := session.RefreshToken()
			if err == nil {
				fmt.Println(session.GetAccessToken())
				fmt.Println(session.GetClearance())
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
	browser, closeBrowser, err := NewBrowser("")
	if err != nil {
		t.Fatal(err)
	}
	defer closeBrowser()
	// use access token
	session := NewSessionWithAccessToken(browser, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1UaEVOVUpHTkVNMVFURTRNMEZCTWpkQ05UZzVNRFUxUlRVd1FVSkRNRU13UmtGRVFrRXpSZyJ9").AutoRefresh()
	if session.ClearanceIsInValid() {
		for i := 0; i < retry; i++ {
			err := session.RefreshClearance()
			if err == nil {
				fmt.Println(session.GetClearance())
				break
			}
			if i == retry-1 {
				t.Fatal(err)
			}
		}
	}
}
