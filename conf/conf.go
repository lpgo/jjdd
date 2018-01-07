package conf

import (
	"github.com/achun/tom-toml"
	"log"
)

var config toml.Toml

func Get(key string) string {
	return getConfig("app.toml")[key].String()
}

func getConfig(file string) toml.Toml {
	if config != nil {
		return config
	}
	var err error
	config, err = toml.LoadFile(file)
	if err != nil {
		log.Println(err)
		return nil
	} else {
		return config
	}
}
