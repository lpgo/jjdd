package db

import (
	//"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"time"
)

//链接
type Link struct {
	Id       bson.ObjectId `bson:"_id" json:"id"`
	Name     string        `bson:"name" json:"name"`
	Url      string        `bson:"url" json:"url"`
	Category string        `bson:"category" json:"category"`
	Order    int           `bson:"order" json:"order"`
	IsHidden bool          `bson:"ishidden" json:"ishidden"`
}

//值班表
type Rota struct {
	Lingdao string   `bson:"lingdao" json:"lingdao"`
	Zuzhang string   `bson:"zuzhang" json:"zuzhang"`
	Chujing []string `bson:"chujing" json:"chujing"`
	Zhiban  []string `bson:"zhiban" json:"zhiban"`
	Beiqing []string `bson:"beiqing" json:"beiqing"`
	Jiejing []string `bson:"jiejing" json:"jiejing"`
	Tel     string   `bson:"tel" json:"tel"`
}

//通讯录
type Directory struct {
	Id         bson.ObjectId `bson:"_id" json:"id"`
	Name       string        `bson:"name" json:"name"`
	Department string        `bson:"dep" json:"dep"`
	Job        string        `bson:"job" json:"job"`
	Phone      string        `bson:"phone" json:"phone"`
	Tel        string        `bson:"tel" json:"tel"`
	Order      int           `bson:"order" json:"order"`
	IsHidden   bool          `bson:"ishidden" json:"ishidden"`
}

type Department struct {
	Id       bson.ObjectId `bson:"_id" json:"id"`
	Name     string        `bson:"name" json:"name"`
	Order    int           `bson:"order" json:"order"`
	IsHidden bool          `bson:"ishidden" json:"ishidden"`
}

//用户
type User struct {
	Id         bson.ObjectId `bson:"_id" json:"id"`
	Name       string        `bson:"name" json:"name"`
	Password   string        `bson:"pwd" json:"-"`
	Role       string        `bson:"role" json:"role"`
	Department string        `bson:"dep" json:"dep"`
	IsHidden   bool          `bson:"ishidden" json:"ishidden"`
}

type Article struct {
	Id         bson.ObjectId `bson:"_id" json:"id"`
	Title      string        `bson:"title" json:"title"`
	Creator    string        `bson:"creator" json:"creator"`     //拟稿人
	Assessor   string        `bson:"assessor" json:"assessor"`   //审核人
	Signature  string        `bson:"signature" json:"signature"` //签发人
	From       string        `bson:"from" json:"from"`           //来源
	Pic        string        `bson:"pic" json:"pic"`             //标题图片
	Content    template.HTML `bson:"content" json:"content"`
	Time       time.Time     `bson:"time" json:"time"`
	Category   string        `bson:"category" json:"category"`     //分类(移动文章)
	Subject    string        `bson:"subject" json:"subject"`       //专题
	IsHot      bool          `bson:"isHot" json:"isHot"`           //头条要闻
	IsImage    bool          `bson:"isImage" json:"isImage"`       //图片新闻
	IsTraffic  bool          `bson:"isTraffic" json:"isTraffic"`   //交管要闻
	IsAuditing bool          `bson:"isAuditing" json:"isAuditing"` //是否审核
	Hits       int64         `bson:"hits" json:"hits"`
	IsHidden   bool          `bson:"ishidden" json:"ishidden"`
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
