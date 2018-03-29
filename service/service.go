package service

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"jjdd/db"
	"log"
	"regexp"
	"strings"
)

func LoginByName(name, pwd string) *db.User {
	var user db.User
	if err := db.FindOne("user", bson.M{"name": name, "pwd": pwd}, &user); err != nil {
		log.Println(err)
		return nil
	}
	return &user
}

func ChangePwd(name, oldPwd, newPwd string) error {
	return db.UpdateByCond("user", bson.M{"name": name, "pwd": oldPwd}, bson.M{"$set": bson.M{"pwd": newPwd}})
}

//后台不显示文章内容
func GetArticlesByPage(page, count int, cond bson.M) []db.Article {
	var articles = make([]db.Article, count)

	err := db.FindPartOrder("article", cond, (page-1)*count, count, &articles, "-time")

	if err != nil {
		log.Println(err)
	}
	for index, _ := range articles {
		articles[index].Content = ""
	}
	return articles
}

//用于前台
func GetIndexArticlesByPage(page, count int, cond bson.M) []db.Article {
	var articles = make([]db.Article, count)

	err := db.FindPartOrder("article", cond, (page-1)*count, count, &articles, "-time")

	if err != nil {
		log.Println(err)
	}

	for index, _ := range articles {
		articles[index].Content = template.HTML(NoHtml(string(articles[index].Content)))
		if len([]rune(string(articles[index].Content))) > 300 {
			articles[index].Content = template.HTML([]rune(string(articles[index].Content))[:300])
			articles[index].Content += "..."
		}
	}
	return articles
}

func DelArticle(id string) error {
	return db.Delete("article", id)
}

func GetArticlesCount(cond bson.M) int {
	return db.GetCount("article", cond)
}

func AddUser(user db.User) error {
	var u db.User
	if err := db.FindOne("user", bson.M{"name": user.Name}, &u); err == nil {
		return errors.New("用户名重复")
	} else {
		if err := db.Add("user", &user); err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
}

func AddDep(dep db.Department) error {
	var u db.Department
	if err := db.FindOne("department", bson.M{"name": dep.Name}, &u); err == nil {
		return errors.New("部门名重复")
	} else {
		if err := db.Add("department", &dep); err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
}

func GetUsersByPage(page int) []db.User {
	var users = make([]db.User, 15)
	if err := db.FindPart("user", nil, (page-1)*15, 15, &users); err != nil {
		log.Println(err)
	}
	for index, _ := range users {
		users[index].Password = ""
	}
	return users
}

func GetDepsByPage(page int) []db.Department {
	var deps = make([]db.Department, 15)
	if err := db.FindPartOrder("department", nil, (page-1)*15, 15, &deps, "order"); err != nil {
		log.Println(err)
	}
	return deps
}

func GetSubjectsByPage(page int) []db.Subject {
	var subs = make([]db.Subject, 15)
	if err := db.FindPart("subject", nil, (page-1)*15, 15, &subs); err != nil {
		log.Println(err)
	}
	return subs
}

func GetAllDeps() []db.Department {
	var deps = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		log.Println(err)
	}
	return deps
}

func GetAllSubjects() []db.Subject {
	var deps = make([]db.Subject, 5)
	if err := db.GetAll("subject", &deps); err != nil {
		log.Println(err)
	}
	return deps
}

func GetDirectorysByName(name string) []db.Directory {
	var deps = make([]db.Directory, 15)
	if err := db.GetAllByOrder("directory", bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: name, Options: "ixs"}}}, &deps); err != nil {
		log.Println(err)
	}
	return deps
}

func GetLinksByName(name string) []db.Link {
	var deps = make([]db.Link, 15)
	if err := db.GetAllByOrder("link", bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: name, Options: "ixs"}}}, &deps); err != nil {
		log.Println(err)
	}
	return deps
}

func GetDirectorysByDep(dep string, page int) []db.Directory {
	var deps = make([]db.Directory, 15)
	if err := db.FindPartOrder("directory", bson.M{"dep": dep}, (page-1)*15, 15, &deps, "order"); err != nil {
		log.Println(err)
	}
	return deps
}

func GetLinksByDep(dep string, page int) []db.Link {
	var deps = make([]db.Link, 15)
	if err := db.FindPartOrder("link", bson.M{"category": dep}, (page-1)*15, 15, &deps, "order"); err != nil {
		log.Println(err)
	}
	return deps
}

func GetRota() (db.Rota, bool) {
	var rota db.Rota
	if err := db.FindOne("rota", nil, &rota); nil != err {
		log.Println(err)
		return rota, false
	} else {
		return rota, true
	}
}

func GetLinks() []db.Link {
	var links []db.Link
	if err := db.GetAll("link", &links); nil != err {
		log.Println(err)

	}
	return links
}

func GetHotArticle() db.Article {
	var article db.Article
	if err := db.FindOne("article", bson.M{"isHot": true, "isAuditing": true}, &article); err != nil {
		log.Println(err)
	}
	return article
}

func GetImageArticles() []db.Article {
	var articles = make([]db.Article, 20)
	if err := db.FindManyOrder("article", bson.M{"isImage": true, "isAuditing": true}, "time", 20, &articles); err != nil {
		log.Println(err)
	}
	return articles
}

func GetTrafficArticles() []db.Article {
	var articles = make([]db.Article, 20)
	if err := db.FindManyOrder("article", bson.M{"isTraffic": true, "isAuditing": true}, "time", 8, &articles); err != nil {
		log.Println(err)
	}
	return articles
}

func DelUser(id string) error {
	return db.Delete("user", id)
}

func DelDirectory(id string) error {
	var oldDep db.Directory
	if err := db.GetById("directory", id, &oldDep); err != nil {
		log.Println(err)
		return err
	}
	db.UpdateByCond("directory", bson.M{"dep": oldDep.Department, "order": bson.M{"$gt": oldDep.Order}}, bson.M{"$inc": bson.M{"order": -1}})
	return db.Delete("directory", id)
}

func DelLink(id string) error {
	var oldDep db.Link
	if err := db.GetById("link", id, &oldDep); err != nil {
		log.Println(err)
		return err
	}
	db.UpdateByCond("link", bson.M{"category": oldDep.Category, "order": bson.M{"$gt": oldDep.Order}}, bson.M{"$inc": bson.M{"order": -1}})
	return db.Delete("link", id)
}

func DelDep(id string) error {
	var oldDep db.Department
	if err := db.GetById("department", id, &oldDep); err != nil {
		log.Println(err)
		return err
	}
	db.UpdateByCond("department", bson.M{"order": bson.M{"$gt": oldDep.Order}}, bson.M{"$inc": bson.M{"order": -1}})
	return db.Delete("department", id)
}

func SignArticle(id, dep string) error {
	if err := db.UpdateById("article", id, bson.M{"$pull": bson.M{"unSign": dep}}); err != nil {
		log.Println(err)
		return err
	}
	if err := db.UpdateById("article", id, bson.M{"$addToSet": bson.M{"signed": dep}}); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetUsersCount() int {
	return db.GetCount("user", nil)
}
func GetDepsCount() int {
	return db.GetCount("department", nil)
}
func GetSubjectsCount() int {
	return db.GetCount("subject", nil)
}
func GetDirectorysCount(dep string) int {
	return db.GetCount("directory", bson.M{"dep": dep})
}
func GetLinksCount(c string) int {
	return db.GetCount("link", bson.M{"category": c})
}

func NoHtml(src string) string {
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)

	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")
	return src
}
