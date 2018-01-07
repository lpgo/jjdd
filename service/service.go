package service

import (
	"bytes"
	"first/db"
	"fmt"
	"github.com/Luxurioust/excelize"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"regexp"
	"strconv"
	"time"
)

func LoginByName(name, pwd string) *db.User {
	var user db.User
	if err := db.FindOne("user", bson.M{"name": name, "pwd": pwd}, &user); err != nil {
		log.Println(err)
		return nil
	}
	return &user
}
