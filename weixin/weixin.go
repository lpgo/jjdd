/*
微信小程序相关接口
*/
package weixin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chanxuehong/util"
	mch "github.com/chanxuehong/wechat.v2/mch/core"
	"github.com/chanxuehong/wechat.v2/mch/pay"
	"github.com/chanxuehong/wechat.v2/mp/core"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
	"hotel/conf"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var server core.AccessTokenServer
var client *mch.Client

type wxLogin struct {
	openid      string `json:"openid"`
	session_key string `json:"session_key"`
	errcode     string `json:"errcode"`
	errmsg      string `json:"errmsg"`
}

type PayParam struct {
	Appid      string `json:"errmsg"`
	TimeStamp  string `json:"timestamp"`
	NonceStr   string `json:"nonceStr"`
	PackageStr string `json:"packageStr"`
	Sign       string `json:"sign"`
}

func GetDeskQR(path, id string, no int) (string, error) {
	urlStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=%s", getToken())
	postData := fmt.Sprintf(`{"path": "%s?id=%s&no=%d&type=scan", "width": 430}`, path, id, no)
	resp, err := http.Post(urlStr, "application/json", strings.NewReader(postData))
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	fileName := "public/qr/" + fmt.Sprintf("%s-%d", id, no) + ".jpg"
	dst, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, resp.Body); err != nil {
		return "", err
	}

	return fileName, nil
}

//通过code获取openid和session_key
func GetOpenidByCode(code string) (string, string, error) {
	urlStr := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", conf.Get("wx.appid"), conf.Get("wx.secret"), code)
	resp, err := http.Get(urlStr)
	if err != nil {
		log.Println(err)
		return "", "", err
	}
	defer resp.Body.Close()
	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		log.Println(err1)
		return "", "", err1
	}
	var result wxLogin
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
		return "", "", err
	}
	if result.openid != "" {
		return result.openid, result.session_key, nil
	} else {
		log.Println(result.errcode + result.errmsg)
		return "", "", errors.New(result.errmsg)
	}
}

func PaySign(packageStr string) PayParam {
	appId := conf.Get("wx.appid")
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := bson.NewObjectId().Hex()
	sign := mch.JsapiSign(appId, timeStamp, nonceStr, packageStr, "MD5", conf.Get("wx.apiKey"))
	return PayParam{Appid: appId, TimeStamp: timeStamp, NonceStr: nonceStr, Sign: sign, PackageStr: packageStr}
}

//统一下单接口
func UnifiedOrder(menuName string, fee float32, ip string) (string, error) {
	req := make(map[string]string)
	req["appid"] = conf.Get("wx.appid")
	req["mch_id"] = conf.Get("wx.mch_id")
	req["nonce_str"] = bson.NewObjectId().Hex()
	req["body"] = menuName
	req["out_trade_no"] = bson.NewObjectId().Hex()
	req["total_fee"] = fmt.Sprintf("%.f", (fee * 100))
	req["spbill_create_ip"] = ip
	req["notify_url"] = conf.Get("wx.notify_url")
	req["trade_type"] = "JSAPI"

	resp, err := pay.UnifiedOrder(client, req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if resp["return_code"] == "SUCCESS" && resp["result_code"] == "SUCCESS" {
		return resp["prepay_id"], nil
	} else {
		return "", errors.New(resp["return_msg"])
	}
}

func InitWeixinApi() {
	server = core.NewDefaultAccessTokenServer(conf.Get("wx.appid"), conf.Get("wx.secret"), nil)
	hc, err := mch.NewTLSHttpClient("cert/cert.pem", "cert/key.pem")
	if err != nil {
		log.Println(err)
		return
	}
	client = mch.NewClient(conf.Get("wx.appid"), conf.Get("wx.mch_id"), conf.Get("wx.apiKey"), hc)
}

func getToken() string {
	if token, err := server.Token(); err != nil {
		log.Println(err)
		return ""
	} else {
		return token
	}

}

//微信支付回调
func MchServer(c echo.Context) error {
	var m map[string]string
	var err error
	if m, err = util.DecodeXMLToMap(c.Request().Body()); err != nil {
		log.Println(err)
		return err
	}
	sign := m["sign"]
	delete(m, "sign")
	if sign != mch.Sign(m, conf.Get("wx.apiKey"), nil) {
		log.Println("验证失败")
	}

	log.Println(m)
	return c.String(http.StatusOK, "<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>")
}
