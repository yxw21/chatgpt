package chatgpt

import (
	"fmt"
	"testing"
)

func TestSession(t *testing.T) {
	session := NewSessionWithCredential("example@gmail.com", "123456", "I-1123123KASD").AutoRefresh()
	if session.IsInvalid() {
		session.RefreshToken()
	}
	fmt.Println(session.GetAccessToken())
	fmt.Println(session.GetClearance())
	fmt.Println(session.GetUserAgent())
	//session := NewSessionWithAccessToken("jwt").AutoRefresh()
	//fmt.Println(session)
}
