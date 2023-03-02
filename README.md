
chatgpt api has been released, it is recommended to use the official api

https://platform.openai.com/docs/api-reference/chat

<del>
# Tips
The third-party library only supports linux, so this project only supports linux.

# Dependency

### Xvfb (Required)

On Ubuntu or Debian
```
apt update
apt install xvfb
```
On CentOS
```
yum update
yum install xorg-x11-server-Xvfb
```
On Alpine
```
apk update
apk add xvfb
```
### Chrome (Required)

On Ubuntu or Debian
```
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
apt install ./google-chrome-stable_current_amd64.deb
```
On CentOS
```
wget https://dl.google.com/linux/direct/google-chrome-stable_current_x86_64.rpm
yum localinstall -y google-chrome-stable_current_x86_64.rpm
```
On Alpine
```
apk add chromium
```

# Login
It is recommended to provide both the username password and the accesstoken, because when providing the accesstoken, the program will use the username password to refresh the accesstoken only when the accesstoken is about to expire, thus reducing the user's waiting time to refresh the accesstoken.
```golang
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
	"log"
)

func main() {
	browser, closeBrowser, err := chatgpt.NewBrowser(chatgpt.BrowserOptions{
		//Proxy:        "socks5://38.91.107.224:36699",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer closeBrowser()
	// [!!!] Make sure the session is initialized once
	session := (&chatgpt.Session{
		Browser:     browser,
		Username:    "example@gmail.com",
		Password:    "password",
		AccessToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1UaEVOVUpHTkVNMVFURTRNMEZCTWpkQ05UZzVNRFUxUlRVd1FVSkRNRU13UmtGRVFrRXpSZyJ9",
	}).AutoRefresh()
	chat := chatgpt.NewChat(browser, session)
	res, err := chat.Send("hi")
	if err != nil{
		log.Panic(err)
	}
	fmt.Println(res.Message.Content.Parts)
}
```
# Access token (Seems to expire in 7 days)

https://chat.openai.com/api/auth/session
</del>
