package conf

import (
	"testing"
)

func Test_getConfig(t *testing.T) {
	toml := getConfig("/root/go/gopath/src/hotel/app.toml")
	if toml == nil {
		t.Errorf("get config failed")
	}
}

func TestGet(t *testing.T) {
	apikey := Get("printer.apikey")
	if apikey != "0706719eeb1582baefaa3e18581ffe3f4616567d" {
		t.Errorf("Get apikey failed. Got %s, expected 0706719eeb1582baefaa3e18581ffe3f4616567d.", apikey)
	}
}
