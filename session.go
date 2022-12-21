package chatgpt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

type User struct {
	Exp int64 `json:"exp"`
}

type Session struct {
	AccessToken string `json:"accessToken"`
	Error       string `json:"error"`
}

func (ctx *Session) Expires() time.Time {
	var user User
	sli := strings.Split(ctx.AccessToken, ".")
	if len(sli) != 3 {
		return time.Unix(2563200, 0)
	}
	bs, err := base64.StdEncoding.DecodeString(sli[1])
	if err != nil {
		return time.Unix(2563200, 0)
	}
	if err = json.Unmarshal(bs, &user); err != nil {
		return time.Unix(2563200, 0)
	}
	return time.Unix(user.Exp, 0)
}

func (ctx *Session) IsInvalid() bool {
	return ctx.AccessToken == "" || (time.Now().Unix()+60) >= ctx.Expires().Unix()
}
