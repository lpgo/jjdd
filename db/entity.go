package db

import (
	//"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"time"
)

//用户
type User struct {
	Id       bson.ObjectId `bson:"_id" json:"id"`
	Name     string        `bson:"name" json:"name"`
	Password string        `bson:"pwd" json:"-"`
	Role     string        `bson:"role" json:"role"`
}

type Article struct {
	Id         bson.ObjectId `bson:"_id" json:"id"`
	Title      string        `bson:"title" json:"title"`
	Creator    string        `bson:"creator" json:"creator"`
	Assessor   string        `bson:"assessor" json:"assessor"`
	Signature  string        `bson:"signature" json:"signature"`
	From       string        `bson:"from" json:"from"`
	Pic        string        `bson:"pic" json:"pic"`
	Content    template.HTML `bson:"content" json:"content"`
	Time       time.Time     `bson:"time" json:"time"`
	Category   string        `bson:"category" json:"category"`
	Subject    string        `bson:"subject" json:"subject"`
	IsAuditing bool          `bson:"isAuditing" json:"isAuditing"`
	Hits       int64         `bson:"hits" json:"hits"`
}

func ArticleFromMap(data map[string]Any) Article {
	return Article{
		Title:     data["title"].(string),
		Creator:   data["creator"].(string),
		Assessor:  data["assessor"].(string),
		Signature: data["signature"].(string),
		From:      data["from"].(string),
		Pic:       data["pic"].(string),
		Content:   data["content"].(template.HTML),
		Category:  data["category"].(string),
		Subject:   data["subject"].(string),
		Id:        bson.ObjectIdHex(data["id"].(string)),
	}
}
