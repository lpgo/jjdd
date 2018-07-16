package service

import (
	"errors"
	"fmt"
	"html/template"
	"jjdd/db"
	"log"
	"regexp"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var menuSlice []string = []string{
	"文章审核",
	"后台首页",
	"网站管理",
	"一级栏目",
	"文件简报",
	"党建队建",
	"交管动态",
	"学习园地"}

var menuClass map[string][]string = map[string][]string{
	"文章审核": []string{"红头文件", "普通文章"},
	"后台首页": []string{"后台首页"},
	"网站管理": []string{"值班管理", "用户管理", "部门管理", "通讯录管理", "链接管理", "专题管理", "通知管理"},
	"一级栏目": []string{"领导讲话", "大队概括", "督察通报", "每月警星"},
	"文件简报": []string{"重要文件", "通知通报", "交管简报", "人事文件", "交安委文件", "大队活动"},
	"党建队建": []string{"支部活动", "纪律教育", "学习培训", "警营文化"},
	"交管动态": []string{"秩序整治", "事故预防", "科技信息", "交管宣传"},
	"学习园地": []string{"法律法规", "规章制度", "经验调研", "学习交流", "规范执法"}}

var menuItemHtml map[string]string = map[string]string{
	"后台首页":  `<a href="/admin/page/zhongdui_admin">后台首页</a>`,
	"红头文件":  `<a href="/admin/page/dadui_admin?isRed=true">红头文件</a>`,
	"普通文章":  `<a href="/admin/page/dadui_admin?isRed=false">普通文章</a>`,
	"用户管理":  `<a href="/admin/page/user_list">用户管理</a>`,
	"部门管理":  `<a href="/admin/page/dep_list">部门管理</a>`,
	"通讯录管理": `<a href="/admin/page/directory_list">通讯录管理</a>`,
	"链接管理":  `<a href="/admin/page/link_list">链接管理</a>`,
	"值班管理":  `<a href="/admin/page/saveRota">值班管理</a>`,
	"专题管理":  `<a href="/admin/page/subject_list">专题管理</a>`,
	"通知管理":  `<a href="/admin/page/notice_list">通知管理</a>`,
	"领导讲话":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('领导讲话')">领导讲话</a>`,
	"大队概括":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('大队概括')">大队概括</a>`,
	"督察通报":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('督察通报')+'&isRed=true&header=duchatongbao.gif'">督察通报</a>`,
	"每月警星":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('每月警星')">每月警星</a>`,
	"重要文件":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('重要文件')+'&isRed=true'">重要文件</a>`,
	"通知通报":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('通知通报')+'&isRed=true'">通知通报</a>`,
	"人事文件":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('人事文件')+'&isRed=true'">人事文件</a>`,
	"交管简报":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('交管简报')+'&isRed=true&header=jiaoguanjianbao.gif'">交管简报</a>`,
	"交安委文件": `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('交安委文件')+'&isRed=true&header=jiaoanwei.jpg'">交安委文件</a>`,
	"大队活动":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('大队活动')">大队活动</a>`,
	"支部活动":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('支部活动')">支部活动</a>`,
	"纪律教育":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('纪律教育')">纪律教育</a>`,
	"学习培训":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('学习培训')">学习培训</a>`,
	"警营文化":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('警营文化')">警营文化</a>`,
	"秩序整治":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('秩序整治')">秩序整治</a>`,
	"事故预防":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('事故预防')">事故预防</a>`,
	"科技信息":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('科技信息')">科技信息</a>`,
	"交管宣传":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('交管宣传')">交管宣传</a>`,
	"法律法规":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('法律法规')">法律法规</a>`,
	"规章制度":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('规章制度')">规章制度</a>`,
	"经验调研":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('经验调研')">经验调研</a>`,
	"学习交流":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('学习交流')">学习交流</a>`,
	"规范执法":  `<a href="javascript:window.location.href='/admin/page/admin?category='+encodeURIComponent('规范执法')">规范执法</a>`}

func GetMenu(category string) string {
	for k, v := range menuClass {
		for _, c := range v {
			if c == category {
				return k
			}
		}
	}
	return "后台首页"
}

func GetAllSubMenu(sub string) []string {
	return menuClass[sub]
}

func GetAllAuth() []string {
	data := []string{}
	for _, v := range menuClass {
		for _, a := range v {
			data = append(data, a)
		}
	}
	return data
}

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

func GetNoticesByPage(page int) []db.Notice {
	var subs = make([]db.Notice, 15)
	if err := db.FindPartOrder("notice", nil, (page-1)*15, 15, &subs, "-time"); err != nil {
		log.Println(err)
	}
	return subs
}

func GetNotice(id string) *db.Notice {
	var notice db.Notice
	if id == "" {
		if err := db.FindOne("notice", bson.M{"isShow": true}, &notice); err != nil {
			return nil
		} else {
			return &notice
		}
	} else {
		if err := db.GetById("notice", id, &notice); err != nil {
			return nil
		} else {
			return &notice
		}
	}

}

func GetTongZhi() *db.Article {
	var tongzhi db.Article
	if err := db.FindOne("article", bson.M{"isNotice": true}, &tongzhi); err != nil {
		log.Println(err)
		return nil
	}
	noticeDate := tongzhi.NoticeTime.AddDate(0, 0, 1)
	if time.Now().Before(noticeDate) {
		return &tongzhi
	} else {
		return nil
	}
}

func GetAllDeps() []db.Department {
	var deps = make([]db.Department, 10)
	if err := db.GetAllByOrder("department", nil, &deps); err != nil {
		log.Println(err)
	}
	log.Println(deps)
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

func GetRota() template.HTML {
	var rota db.ZRota
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.Local)
	if err := db.FindOne("zrota", bson.M{"dep": "大队", "time": t}, &rota); nil != err {
		log.Println(err)
		return template.HTML("")
	} else {
		return RotaToHtml(rota)
	}
}

func RotaToHtml(rota db.ZRota) template.HTML {

	result := ""

	layout1 :=
		`<tr>
			<td>%s</td>
			<td colspan="%d">
			    <span style='cursor:pointer' onclick='searchphone("%s","%s")'>%s</span>
			</td>
		</tr>`
	layout2 :=
		`<tr>
			<td rowspan="%d">%s</td>
			<td>
			    <span style='cursor:pointer' onclick='searchphone("%s","%s")'>%s</span>
			</td>`

	for _, item := range rota.Duty {

		l := len(item.Staff)
		var rowspan, colspan int
		if l < 2 {
			colspan = 2
			rowspan = 1
			if len(item.Staff) == 0 {
				result += fmt.Sprintf(layout1, item.Name, colspan, rota.Dep, "", "")
			} else {
				result += fmt.Sprintf(layout1, item.Name, colspan, rota.Dep, item.Staff[0], addSpace(item.Staff[0]))
			}

		} else {
			if l%2 > 0 {
				rowspan = l/2 + 1
			} else {
				rowspan = l / 2
			}
			for i, p := range item.Staff {

				if i%2 == 0 {
					if i == 0 {
						result += fmt.Sprintf(layout2, rowspan, item.Name, rota.Dep, p, addSpace(p))
					} else {
						if i == l {
							result += fmt.Sprintf(`<tr><td><span style='cursor:pointer' onclick='searchphone("%s","%s")'>%s</span></td></tr>`, rota.Dep, p, addSpace(p))
						} else {
							result += fmt.Sprintf(`<tr><td><span style='cursor:pointer' onclick='searchphone("%s","%s")'>%s</span></td>`, rota.Dep, p, addSpace(p))
						}
					}
				} else {
					result += fmt.Sprintf(`<td><span style='cursor:pointer' onclick='searchphone("%s","%s")'>%s</span></td></tr>`, rota.Dep, p, addSpace(p))
				}
			}
		}
	}

	return template.HTML(result)
}

func GetLinks() []db.Link {
	var links []db.Link
	if err := db.GetAllByOrder("link", nil, &links); nil != err {
		log.Println(err)

	}
	return links
}

//获取首页文章（七条）
func GetIndexArticle(class, category string) []db.Article {
	var articles = make([]db.Article, 7)
	cond := bson.M{"isPass": true}

	if class != "" {
		cond["class"] = class
	}
	if category != "" {
		cond["category"] = category
	}

	if err := db.FindManyOrder("article", cond, "-time", 7, &articles); err != nil {
		log.Println(err)
	}
	return articles
}

func GetHotArticle() db.Article {
	var article db.Article
	if err := db.FindOne("article", bson.M{"isHot": true, "isPass": true}, &article); err != nil {
		log.Println(err)
	}
	return article
}

func GetImageArticles() []db.Article {
	var articles = make([]db.Article, 20)
	if err := db.FindManyOrder("article", bson.M{"isImage": true, "isPass": true}, "-time", 20, &articles); err != nil {
		log.Println(err)
	}
	return articles
}

func GetTrafficArticles() []db.Article {

	var a db.Article
	if err := db.FindOne("article", bson.M{"isTraffic": true, "isTop": true, "isPass": true}, &a); err != nil {
		log.Println(err)
		var articles = make([]db.Article, 8)
		if err := db.FindManyOrder("article", bson.M{"isTraffic": true, "isTop": false, "isPass": true}, "-time", 8, &articles); err != nil {
			log.Println(err)
		}
		return articles
	} else {
		var articles = make([]db.Article, 7)
		if err := db.FindManyOrder("article", bson.M{"isTraffic": true, "isTop": false, "isPass": true}, "-time", 7, &articles); err != nil {
			log.Println(err)
		}
		var as = make([]db.Article, len(articles)+1)
		as[0] = a
		copy(as[1:], articles)

		return as
	}

}

func GetLeaderArticles() []db.Article {
	var a db.Article
	if err := db.FindOne("article", bson.M{"category": "领导讲话", "isTop": true, "isPass": true}, &a); err != nil {
		log.Println(err)
		var articles = make([]db.Article, 8)
		if err := db.FindManyOrder("article", bson.M{"category": "领导讲话", "isTop": false, "isPass": true}, "-time", 8, &articles); err != nil {
			log.Println(err)
		}
		return articles
	} else {
		var articles = make([]db.Article, 7)
		if err := db.FindManyOrder("article", bson.M{"category": "领导讲话", "isTop": false, "isPass": true}, "-time", 7, &articles); err != nil {
			log.Println(err)
		}
		var as = make([]db.Article, len(articles)+1)
		as[0] = a
		copy(as[1:], articles)

		return as
	}
}

func GetDuChaArticles() []db.Article {
	var articles = make([]db.Article, 7)
	if err := db.FindManyOrder("article", bson.M{"category": "督察通报", "isPass": true}, "-time", 7, &articles); err != nil {
		log.Println(err)
	}
	return articles
}

func GetStarArticles() []db.Article {
	var articles = make([]db.Article, 7)
	if err := db.FindManyOrder("article", bson.M{"category": "每月警星", "isPass": true}, "-time", 7, &articles); err != nil {
		log.Println(err)
	}
	return articles
}

func GetSummarization() string {
	var articles = make([]db.Article, 1)
	if err := db.FindManyOrder("article", bson.M{"category": "大队概括", "isPass": true}, "-time", 1, &articles); err != nil {
		log.Println(err)
	}
	if len(articles) > 0 {
		return articles[0].Id.Hex()
	} else {
		return ""
	}

}

func DelUser(id string) error {
	return db.Delete("user", id)
}

func DelNotice(id string) error {
	return db.Delete("notice", id)
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
func GetNoticesCount() int {
	return db.GetCount("notice", nil)
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

//稿件统计(当月)
func Statistics(auditing bool) []bson.M {
	year := time.Now().Year()
	start := time.Date(year, 0, 0, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, 0, 0, 0, 0, 0, 0, time.Local)

	resp := []bson.M{}
	query := func(c *mgo.Collection) error {
		if auditing {
			pipe := c.Pipe([]bson.M{
				{"$match": bson.D{{Name: "isPass", Value: true}, {Name: "time", Value: bson.M{"$gte": start}}, {Name: "time", Value: bson.M{"$lt": end}}}},
				//{"$project": bson.M{"time": bson.M{"$add": bson.D{{"$time", 28800000}}}, "com": 1, "driver": 1, "capacity": 1, "project": 1, "car": 1, "way": 1, "part": 1, "strength": 1, "price": 1, "total": 1, "carFee": 1, "autoFee": 1, "driverFee": 1}},
				{"$group": bson.M{"_id": bson.M{"from": "$dep", "month": bson.M{"$month": "$time"}}, "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
				{"$match": bson.M{"_id.month": int(time.Now().Month())}},
			})
			return pipe.All(&resp)
		} else {
			pipe := c.Pipe([]bson.M{
				{"$match": bson.D{{Name: "time", Value: bson.M{"$gte": start}}, {Name: "time", Value: bson.M{"$lt": end}}}},
				//{"$project": bson.M{"time": bson.M{"$add": bson.D{{"$time", 28800000}}}, "com": 1, "driver": 1, "capacity": 1, "project": 1, "car": 1, "way": 1, "part": 1, "strength": 1, "price": 1, "total": 1, "carFee": 1, "autoFee": 1, "driverFee": 1}},
				{"$group": bson.M{"_id": bson.M{"from": "$dep", "month": bson.M{"$month": "$time"}}, "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
				{"$match": bson.M{"_id.month": int(time.Now().Month())}},
			})
			return pipe.All(&resp)
		}
	}
	if err := db.WitchCollection("article", query); err != nil {
		log.Println(err)
	}
	return resp
}

func StatisticsPerMonth(auditing bool, year int) []bson.M {
	start := time.Date(year, 0, 0, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, 0, 0, 0, 0, 0, 0, time.Local)

	resp := []bson.M{}
	query := func(c *mgo.Collection) error {
		if auditing {
			pipe := c.Pipe([]bson.M{
				{"$match": bson.D{{Name: "isPass", Value: true}, {Name: "time", Value: bson.M{"$gte": start}}, {Name: "time", Value: bson.M{"$lt": end}}}},
				//{"$project": bson.M{"time": bson.M{"$add": bson.D{{"$time", 28800000}}}, "com": 1, "driver": 1, "capacity": 1, "project": 1, "car": 1, "way": 1, "part": 1, "strength": 1, "price": 1, "total": 1, "carFee": 1, "autoFee": 1, "driverFee": 1}},
				{"$group": bson.M{"_id": bson.M{"from": "$dep", "month": bson.M{"$month": "$time"}}, "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
			})
			return pipe.All(&resp)
		} else {
			pipe := c.Pipe([]bson.M{
				{"$match": bson.D{{Name: "time", Value: bson.M{"$gte": start}}, {Name: "time", Value: bson.M{"$lt": end}}}},
				//{"$project": bson.M{"time": bson.M{"$add": bson.D{{"$time", 28800000}}}, "com": 1, "driver": 1, "capacity": 1, "project": 1, "car": 1, "way": 1, "part": 1, "strength": 1, "price": 1, "total": 1, "carFee": 1, "autoFee": 1, "driverFee": 1}},
				{"$group": bson.M{"_id": bson.M{"from": "$dep", "month": bson.M{"$month": "$time"}}, "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
			})
			return pipe.All(&resp)
		}
	}
	if err := db.WitchCollection("article", query); err != nil {
		log.Println(err)
	}
	return resp
}

func StatisticsOneYear(auditing bool, year int) []bson.M {

	start := time.Date(year, 0, 0, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, 0, 0, 0, 0, 0, 0, time.Local)
	resp := []bson.M{}
	query := func(c *mgo.Collection) error {
		if auditing {
			pipe := c.Pipe([]bson.M{
				{"$match": bson.D{{Name: "isPass", Value: true}, {Name: "time", Value: bson.M{"$gte": start}}, {Name: "time", Value: bson.M{"$lt": end}}}},
				//{"$project": bson.M{"time": bson.M{"$add": bson.D{{"$time", 28800000}}}, "com": 1, "driver": 1, "capacity": 1, "project": 1, "car": 1, "way": 1, "part": 1, "strength": 1, "price": 1, "total": 1, "carFee": 1, "autoFee": 1, "driverFee": 1}},
				{"$group": bson.M{"_id": bson.M{"from": "$dep", "year": bson.M{"$year": "$time"}}, "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
			})
			return pipe.All(&resp)
		} else {
			pipe := c.Pipe([]bson.M{
				{"$match": bson.D{{Name: "time", Value: bson.M{"$gte": start}}, {Name: "time", Value: bson.M{"$lt": end}}}},
				//{"$project": bson.M{"time": bson.M{"$add": bson.D{{"$time", 28800000}}}, "com": 1, "driver": 1, "capacity": 1, "project": 1, "car": 1, "way": 1, "part": 1, "strength": 1, "price": 1, "total": 1, "carFee": 1, "autoFee": 1, "driverFee": 1}},
				{"$group": bson.M{"_id": bson.M{"from": "$dep", "year": bson.M{"$year": "$time"}}, "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1}},
			})
			return pipe.All(&resp)
		}
	}
	if err := db.WitchCollection("article", query); err != nil {
		log.Println(err)
	}
	return resp
}

func ToRotaTime(t string) time.Time {
	if ti, err := time.ParseInLocation("2006-01-02 15:04:05", t+" 08:00:00", time.Local); err != nil {
		return time.Now()
	} else {
		return ti
	}
}

func addSpace(s string) string {
	str := []rune(s)
	if len(str) == 2 {
		return string(str[0]) + "&emsp;" + string(str[1])
	} else {
		return s
	}
}

//subMenu 二级菜单名称，用于在一级菜单选中状态
func CreateMenuHtml(user *db.User, subMenu string) (template.HTML, template.HTML) {
	data := map[string][]string{}

	for _, a := range user.Authorities {
		menu := GetMenu(a)
		if subs, ok := data[menu]; !ok {
			data[menu] = []string{a}
		} else {
			data[menu] = append(subs, a)
		}
	}

	var navBar, subNav string
	hot := GetMenu(subMenu)

	for _, k := range menuSlice {
		if v, ok := data[k]; ok {
			if k == hot {
				navBar += `<li class="m on"><h3><a  href="#">` + k + `</a></h3></li><li class="s">|</li>`
			} else {
				navBar += `<li class="m"><h3><a  href="#">` + k + `</a></h3></li><li class="s">|</li>`
			}
			subNav += "<li>"
			for _, s := range v {
				subNav += menuItemHtml[s] + "|"
			}
			l := len([]rune(subNav))
			subNav = string(string([]rune(subNav)[:l-1]))
			subNav += "</li>"
		}
	}

	return template.HTML(navBar), template.HTML(subNav)

}

func GetSeachKeys() []db.SearchKey {
	keys := make([]db.SearchKey, 0)
	db.FindManyOrder("searchKey", nil, "-count", 6, &keys)
	return keys
}
