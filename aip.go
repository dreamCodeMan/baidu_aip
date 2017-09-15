package baiduAip

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//认证API
const (
	BAIDU_OAUTH_URL           string = "https://aip.baidubce.com/oauth/2.0/token"
	BAIDU_ANTIPORN_URL        string = "https://aip.baidubce.com/rest/2.0/antiporn/v1/detect"
	BAIDU_ANTIPORN_GIF_URL    string = "https://aip.baidubce.com/rest/2.0/antiporn/v1/detect_gif"
	BAIDU_ANTITERROR_URL      string = "https://aip.baidubce.com/rest/2.0/antiterror/v1/detect"
	BAIDU_FACEAUDIT_URL       string = "https://aip.baidubce.com/rest/2.0/solution/v1/face_audit"
	BAIDU_IMAGECENSORCOMB_URL string = "https://aip.baidubce.com/api/v1/solution/direct/img_censor"
)

//设置超时
var timeout time.Duration = time.Second * 30
var client = &http.Client{
	Transport: &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, timeout)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(timeout))
			return conn, nil
		},
		ResponseHeaderTimeout: timeout,
	},
}

type BaiduClientConfig struct {
	App_ID     string
	Api_key    string
	Secret_key string
}

type BaiduAuthBackInfo struct {
	Access_Token      string `json:"access_token"`
	Session_Key       string `json:"session_key"`
	Scope             string `json:"scope"`
	Refresh_Token     string `json:"refresh_token"`
	Session_Secret    string `json:"session_secret"`
	Expires_In        int    `json:"expires_in"`
	Error             string `json:"error"`
	Error_Description string `json:"error_description"`
}

// 检查文件或目录是否存在
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

//获得token信息
func (c *BaiduClientConfig) auth() (BaiduAuthBackInfo, error) {
	responseBody := new(BaiduAuthBackInfo)
	baiduUrl := fmt.Sprintf("%s/?grant_type=client_credentials&client_id=%s&client_secret=%s", BAIDU_OAUTH_URL, c.Api_key, c.Secret_key)
	resp, err := client.Get(baiduUrl)
	if err != nil {
		return *responseBody, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return *responseBody, err
	}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return *responseBody, err
	}
	return *responseBody, nil

}

//处理图片信息
func (c *BaiduClientConfig) requestBaidu(pic, api_url string) (content []byte, err error) {
	var picCon []byte

	baiduBackInfo, err := c.auth()
	if err != nil {
		return content, err
	}
	if baiduBackInfo.Error != "" {
		return content, fmt.Errorf("error=%s;error_description=%s", baiduBackInfo.Error, baiduBackInfo.Error_Description)
	}

	if strings.HasPrefix(pic, "http") {
		resp, err := client.Get(pic)
		if err != nil {
			return content, err
		}
		if resp.StatusCode != http.StatusOK {
			return content, fmt.Errorf("error:status code is %d", resp.StatusCode)
		}
		defer resp.Body.Close()
		picCon, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return content, err
		}
	} else {
		if !Exist(pic) {
			return content, fmt.Errorf("%s not find", pic)
		}
		f, err := os.Open(pic)
		if err != nil {
			return content, err
		}
		defer f.Close()
		picCon, err = ioutil.ReadAll(f)
		if err != nil {
			return content, err
		}
	}
	base64Pic := base64.StdEncoding.EncodeToString(picCon)

	v := url.Values{}
	if strings.Contains(api_url, "face") {
		v.Set("images", base64Pic)
	} else {
		v.Set("image", base64Pic)
	}
	v.Set("aipSdk", "golang")
	v.Set("aipVersion", "1_5_0")
	v.Set("access_token", baiduBackInfo.Access_Token)
	info := ioutil.NopCloser(strings.NewReader(v.Encode()))
	req, err := http.NewRequest("POST", api_url, info)
	if err != nil {
		return content, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return content, err
	}
	defer resp.Body.Close()

	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return content, err
	}

	return content, nil
}

//识别一般图片
func (c *BaiduClientConfig) AntiPorn(pic string) (content []byte, err error) {
	return c.requestBaidu(pic, BAIDU_ANTIPORN_URL)
}

//识别一般gif图片
func (c *BaiduClientConfig) AntiPornGif(pic string) (content []byte, err error) {
	return c.requestBaidu(pic, BAIDU_ANTIPORN_GIF_URL)
}

//暴恐图像识别
func (c *BaiduClientConfig) AntiTerror(pic string) (content []byte, err error) {
	return c.requestBaidu(pic, BAIDU_ANTITERROR_URL)
}

//头像审核
func (c *BaiduClientConfig) FaceAudit(pic string) (content []byte, err error) {
	return c.requestBaidu(pic, BAIDU_FACEAUDIT_URL)
}

//组合审核
func (c *BaiduClientConfig) ImageCensorComb(pic string) (content []byte, err error) {
	return c.requestBaidu(pic, BAIDU_FACEAUDIT_URL)
}
