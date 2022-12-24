# Tips
The third-party library only supports linux, so this project only supports linux.

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
)

func main() {
	retry := 3
	// Make sure the session is initialized once
	session := chatgpt.NewSessionWithCredential("example@gmail.com", "123456", "I-1123123KASD").AutoRefresh()
	chat := chatgpt.NewChat(session)
	for i := 0; i < retry; i++ {
		res, err := chat.Send("hi")
		if err == nil{
			fmt.Println(res.Message.Content.Parts)
			break
        }
		if i == retry - 1 {
			fmt.Println(err)
		}
	}
}
```
2.AccessToken login
```golang
package main

import (
	"fmt"
	"github.com/yxw21/chatgpt"
)

func main() {
	retry := 3
	// Make sure the session is initialized once
	session := chatgpt.NewSessionWithAccessToken("AccessToken").AutoRefresh()
	chat := chatgpt.NewChat(session)
	for i := 0; i < retry; i++ {
		res, err := chat.Send("hi")
		if err == nil{
			fmt.Println(res.Message.Content.Parts)
			break
		}
		if i == retry - 1 {
			fmt.Println(err)
		}
	}
}
```
# Access token (Seems to expire in 7 days)

https://chat.openai.com/api/auth/session
