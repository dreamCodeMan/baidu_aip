## 百度图片审查golang sdk

## Example
```go
package main

import (
	"fmt"
	"github.com/dreamCodeMan/baidu_aip"
)

func main() {
	baiduClient := &baiduAip.BaiduClientConfig{
		App_ID:     "YOU APP_ID",
		Api_key:    "YOU API_KEY",
		Secret_key: "YOU SECRET_KEY",
	}
	con, err := baiduClient.AntiPorn("1.jpg")

	fmt.Println(string(con), err)
}

```