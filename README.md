# Tips
The third-party library only supports linux, so this project only supports linux.

# Dependency

### Xvfb

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
### Key
You need a key to crack the verification code, you can go to the website `nopecha.com` to register, it is very cheap.

```
https://nopecha.com
```
### Chrome

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
1. username and password
```golang
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
	"log"
)

func main() {
	browser, closeBrowser, err := chatgpt.NewBrowser(chatgpt.BrowserOptions{
		ExtensionKey: "I-ABCDEFGHIJKL",
		//Proxy:        "socks5://38.91.107.224:36699",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer closeBrowser()
	// Make sure the session is initialized once
	session := chatgpt.NewSessionWithCredential(browser, "example@gmail.com", "123456").AutoRefresh()
	chat := chatgpt.NewChat(browser, session)
	res, err := chat.Send("hi")
	if err != nil{
		log.Panic(err)
	}
	fmt.Println(res.Message.Content.Parts)
}
```
2.AccessToken login
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
	// Make sure the session is initialized once
	session := chatgpt.NewSessionWithAccessToken(browser, "AccessToken").AutoRefresh()
	chat := chatgpt.NewChat(browser, session)
	res, err := chat.Send("hi")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(res.Message.Content.Parts)
}
```
# Access token (Seems to expire in 7 days)

https://chat.openai.com/api/auth/session
