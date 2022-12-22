# Dependency

### Xvfb (Only linux)

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
	session "github.com/yxw21/chatgpt/session"
)

func main() {
	chat := chatgpt.NewChat(session.NewSessionWithCredential("example@gmail.com", "123456", "I-1123123KASD").AutoRefresh())
	res, err := chat.Send("hi")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Message.Content.Parts)
}
```
2.access token login
```golang
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
	session "github.com/yxw21/chatgpt/session"
)

func main() {
	chat := chatgpt.NewChat(session.NewSessionWithAccessToken("{access token}"))
	res, err := chat.Send("hi")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Message.Content.Parts)
}
```
# Access token (expires in about a day)

https://chat.openai.com/api/auth/session
