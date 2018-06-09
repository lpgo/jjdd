package main

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
	"io/ioutil"
	"jjdd/db"
	"jjdd/service"
	"jjdd/session"
	"net/http"
	"os"
	"strconv"
	//"strings"
	//"github.com/Luxurioust/excelize"
	//"fmt"
	"log"
	"time"
)

/*
	创建标准的RESTful  api
*/

//文章类型
var clazz map[string][]string = map[string][]string{
	"文件简报": []string{"交管简报", "通知通报", "人事文件", "交安委文件", "大队活动"},
	"一级栏目": []string{"领导讲话", "大队概括", "督察通报", "每月警星"},
	"党建队建": []string{"支部活动", "纪律教育", "学习培训", "交警风采"},
	"交管动态": []string{"秩序整治", "事故预防", "科技信息", "交管宣传"},
	"学习园地": []string{"法律法规", "规章制度", "经验调研", "学习交流"}}

type Any interface{}

type CustomContext struct {
	echo.Context
	*session.Session
}

type Template struct {
	templates *template.Template
}

//通用请求返回
type Resp struct {
	Error string `json:"error"`
	Text  string `json:"text"`
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]db.Any); isMap {
		viewContext["reverse"] = c.Echo().Reverse
		viewContext["Hex"] = bson.ObjectId.Hex
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func (c *CustomContext) GetSession(key string) interface{} {
	return c.GetValue(key)
}

func (c *CustomContext) SetSession(key string, value interface{}) {
	c.PutValue(key, value)
}

func sessionHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session.Reset(c)
		cc := &CustomContext{c, session.GetSession(c)}
		return next(cc)
	}
}

func apiGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if cc, ok := c.(*CustomContext); !ok {
			return errors.New("server cc error")
		} else {
			if _, o := cc.GetSession("user").(*db.User); o {
				return next(c)
			}
			return c.Redirect(http.StatusMovedPermanently, "/login.html")
		}
	}
}

func noCache(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Response().Header()
		header["Pragma"] = []string{"no-cache"}
		header["Cache-Control"] = []string{"no-cache"}
		header["Expires"] = []string{"0"}
		return next(c)
	}
}

func refresh(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Echo().Renderer = newRenderer()
		return next(c)
	}
}

func okGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok && he.Code == code {
		c.Logger().Error(err)
		c.Redirect(http.StatusMovedPermanently, "/error.html")
	} else {
		c.Logger().Error(err)
	}

}

func newRenderer() *Template {
	temp := template.New("jjdd")
	temp.Delims("<%", "%>")

	funmap := make(template.FuncMap, 1)
	funmap["Two"] = Two
	funmap["Ten"] = Ten
	funmap["AddSpace"] = AddSpace
	funmap["Format"] = time.Time.Format
	funmap["Add"] = Add
	funmap["Substring"] = Substring
	funmap["IsNew"] = IsNew
	funmap["Include"] = Include
	temp.Funcs(funmap)
	return &Template{
		templates: template.Must(temp.ParseGlob("views/*.html")),
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	e := echo.New()
	e.Renderer = newRenderer()
	e.Static("/", "public")

	//日志
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost"},
		AllowCredentials: true,
	}))

	//热更新模板，发布时去除
	e.Use(refresh)
	e.Use(noCache)

	//Session控制
	e.Use(sessionHandler)
	//最大失效时间相当于没有限制
	session.InitContext(30 * 60)

	//检查RESTful权限
	admin := e.Group("/admin")
	admin.Use(apiGroup)
	/*--------页面--------*/
	admin.GET("/page/publish", publishPage)
	admin.GET("/page/publish_hongtou", publishHongtouPage)
	admin.GET("/page/admin", adminPage)
	admin.GET("/page/dadui_admin", daduiAdminPage)
	admin.GET("/page/zhongdui_admin", zhongduiAdminPage)
	admin.GET("/page/modifyPage", modifyArticlePage)
	admin.GET("/page/set_article", setArticlePage)

	admin.GET("/page/add_user", addUserPage)
	admin.GET("/page/user_list", userListPage)
	admin.GET("/page/modify_user", modifyUserPage)
	admin.GET("/page/changePwd", changePwdPage)

	admin.GET("/page/dep_list", depListPage)
	admin.GET("/page/add_dep", addDepPage)
	admin.GET("/page/modify_dep", modifyDepPage)

	admin.GET("/page/add_directory", addDirectoryPage)
	admin.GET("/page/directory_list", directoryListPage)
	admin.GET("/page/modify_directory", modifyDirectoryPage)

	admin.GET("/page/saveRota", saveRotaPage)
	admin.GET("/page/add_link", addLinkPage)
	admin.GET("/page/link_list", linkListPage)
	admin.GET("/page/modify_link", modifyLinkPage)

	admin.GET("/page/add_subject", addSubjectPage)
	admin.GET("/page/subject_list", subjectListPage)
	admin.GET("/page/modify_subject", modifySubjectPage)

	admin.GET("/page/add_notice", addNoticePage)
	admin.GET("/page/notice_list", noticeListPage)
	admin.GET("/page/modify_notice", modifyNoticePage)

	/*----------------操作--------*/
	//文章
	admin.Any("/publish", publishArticle)
	admin.Any("/preview", previewArticle)
	admin.Any("/previewById", previewArticleById)
	admin.Any("/viewArticle", viewArticle)
	admin.Any("/getArticles", getArticles)
	admin.Any("/delArticle", delArticle)
	admin.Any("/auditing", auditingArticle)
	admin.Any("/modify", modifyArticle)
	admin.Any("/setArticle", setArticle)
	//用户
	admin.Any("/addUser", addUser)
	admin.Any("/getUserList", getUserList)
	admin.Any("/delUser", delUser)
	admin.Any("/modifyUser", modifyUser)
	admin.Any("/changePwd", changePwd)
	//部门
	admin.Any("/addDep", addDep)
	admin.Any("/getDepList", getDepList)
	admin.Any("/delDep", delDep)
	admin.Any("/modifyDep", modifyDep)
	//通讯录
	admin.Any("/getDirectory", getDirectoryList)
	admin.Any("/getDirectoryListByDepName", getDirectoryListByDepName)
	admin.Any("/addDirectory", addDirectory)
	admin.Any("/delDirectory", delDirectory)
	admin.Any("/modifyDirectory", modifyDirectory)

	//值班表
	admin.Any("/saveRota", saveRota)
	admin.Any("/addRotas", addRotas)
	admin.Any("/getRotas", getRotas)

	//链接
	admin.Any("/addLink", addLink)
	admin.Any("/getLink", getLinkList)
	admin.Any("/getLinkListByDepName", getLinkListByDepName)
	admin.Any("/delLink", delLink)
	admin.Any("/modifyLink", modifyLink)

	//专题
	admin.Any("/addSubject", addSubject)
	admin.Any("/getSubjectList", getSubjectList)
	admin.Any("/delSubject", delSubject)
	admin.Any("/modifySubject", modifySubject)

	admin.Any("/addNotice", addNotice)
	admin.Any("/getNoticeList", getNoticeList)
	admin.Any("/delNotice", delNotice)
	admin.Any("/modifyNotice", modifyNotice)
	admin.Any("/setNotice", setNotice)
	admin.Any("/cancelNotice", cancelNotice)
	e.Any("/haveNotice", haveNotice)

	//获取审核未通过数量
	admin.Any("/noPassCount", noPassCount)

	admin.Any("/logout", logout)

	e.Any("/getDirectoryByname1", getDirectoryByName)
	e.Any("/searchDirectoryByName", searchDirectoryByName)
	e.Any("/searchDirectoryByJob", searchDirectoryByJob)
	e.Any("/searchDirectoryByTel", searchDirectoryByTel)
	e.Any("/searchDirectoryByPhone", searchDirectoryByPhone)
	e.Any("/showArticle", showArticleById)
	e.Any("/signArticle", signArticle)
	e.Any("/searchArticle", searchArticle)
	e.Any("/statistics", statisticsPage)

	e.GET("/directory", directoryPage)
	e.GET("/search", searchPage)
	e.GET("/list", listPage)
	e.GET("/subjectList", subjectArticleListPage)
	e.GET("/noLeftList", noLeftListPage)
	e.GET("/summarization", summarizationPage)
	e.GET("/", indexPage)
	e.GET("/login.html", loginPage)
	e.Any("/login", login)
	e.Any("/notice", noticePage)
	e.Any("/zhibanbiao", rotaListPage)
	e.Any("/searchphone", searchPhone)
	//test

	//登录
	//e.POST("/login", login)

	e.Any("/download", download)

	//文件上传
	e.POST("/uploadImage", uploadImage)
	e.POST("/uploadFile", uploadFile)
	e.POST("/imageManager", imageManager)
	//处理微信支付回调
	//e.Post("/mch", weixin.MchServer)

	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Start(":80")

}

func download(c echo.Context) error {
	header := c.Response().Header()

	f, err := os.Open("public/firefox.exe")
	if err != nil {
		return err
	}
	info, _ := f.Stat()
	info.Size()
	header["Content-Length"] = []string{string(info.Size())}
	return c.Stream(http.StatusOK, "application/x-msdownload", f)
}

func addLink(c echo.Context) error {
	link := db.Link{
		Id:       bson.NewObjectId(),
		Name:     c.FormValue("linkName"),
		Url:      c.FormValue("url"),
		Category: c.FormValue("category"),
		Order:    service.GetLinksCount(c.FormValue("category")) + 1,
	}
	if err := db.Add("link", link); err != nil {
		c.Logger().Warn(err)
		return MyRedirect(c, "/error.html")
	} else {
		return MyRedirect(c, "/admin/page/link_list")
	}
}

func addUser(c echo.Context) error {
	params, _ := c.FormParams()
	user := db.User{
		Id:          bson.NewObjectId(),
		Name:        c.FormValue("name"),
		Password:    "888888",
		Department:  c.FormValue("dep"),
		Role:        c.FormValue("role"),
		Authorities: params["quanxian"],
	}

	if user.Role == "中队" {
		user.Authorities = append(user.Authorities, "后台首页")
	}

	if err := service.AddUser(user); err != nil {
		c.Logger().Warn(err)
		return c.Render(http.StatusOK, "adduser", map[string]bool{"NameError": true})
	} else {
		return userListPage(c)
	}
}

func addDep(c echo.Context) error {
	dep := db.Department{
		Id:    bson.NewObjectId(),
		Name:  c.FormValue("depname"),
		Order: service.GetDepsCount() + 1,
	}

	if err := service.AddDep(dep); err != nil {
		c.Logger().Warn(err)
		return c.Render(http.StatusOK, "adddep", map[string]bool{"NameError": true})
	} else {
		return MyRedirect(c, "/admin/page/dep_list")
	}
}

func addSubject(c echo.Context) error {
	sub := db.Subject{
		Id:   bson.NewObjectId(),
		Name: c.FormValue("subjectName"),
		Pic:  c.FormValue("pic"),
	}

	if c.FormValue("isHot") == "true" {
		sub.IsHot = true
		db.UpdateByCond("subject", bson.M{"isHot": true}, bson.M{"$set": bson.M{"isHot": false}})
	} else {
		sub.IsHot = false
	}

	if err := db.Add("subject", sub); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return MyRedirect(c, "/admin/page/subject_list")
	}
}

func addDirectory(c echo.Context) error {
	dir := db.Directory{
		Id:         bson.NewObjectId(),
		Name:       c.FormValue("directoryName"),
		Department: c.FormValue("dep"),
		Job:        c.FormValue("job"),
		Tel:        c.FormValue("tel"),
		Phone:      c.FormValue("phone"),
		Order:      service.GetDirectorysCount(c.FormValue("dep")) + 1,
	}
	if err := db.Add("directory", dir); err != nil {
		c.Logger().Warn(err)
		return c.Render(http.StatusOK, "adddirectory", nil)
	} else {
		return MyRedirect(c, "/admin/page/directory_list")
	}
}

func addNotice(c echo.Context) error {
	notice := db.Notice{
		Id:      bson.NewObjectId(),
		Title:   c.FormValue("title"),
		Content: c.FormValue("content"),
		Dep:     c.FormValue("dep"),
		IsShow:  false,
		Time:    time.Now(),
	}
	if err := db.Add("notice", notice); err != nil {
		c.Logger().Warn(err)
		return c.Render(http.StatusOK, "addnotice", nil)
	} else {
		return MyRedirect(c, "/admin/page/notice_list")
	}
}

func saveRota(c echo.Context) error {
	params, _ := c.FormParams()
	rota := db.Rota{
		Lingdao: params["lingdao"][0],
		Zuzhang: params["zuzhang"][0],
		Chujing: params["chujing"],
		Zhiban:  params["zhiban"],
		Beiqing: params["beiqing"],
		Jiejing: params["jiejing"],
		Tel:     params["tel"][0],
	}
	if err := db.UpsertByCond("rota", bson.M{"time": params["time"][0]}, rota); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return MyRedirect(c, "/admin/page/saveRota")
	}
}

//新的保存值班表
func addRotas(c echo.Context) error {
	var rota db.ZRota
	c.Bind(&rota)

	user := c.(*CustomContext).GetSession("user").(*db.User)
	if user.Role == "大队" {
		rota.Dep = "大队"
	} else {
		rota.Dep = user.Department
	}

	startTime := c.QueryParam("starttime")
	endTime := c.QueryParam("endtime")

	start := service.ToRotaTime(startTime)
	end := service.ToRotaTime(endTime)

	for t := start; t.Before(end); t = t.AddDate(0, 0, 1) {
		rota.Time = t
		if err := db.UpsertByCond("zrota", bson.M{"time": t, "dep": rota.Dep}, rota); err != nil {
			c.Logger().Warn(err)
		}
	}

	rota.Time = end
	if err := db.UpsertByCond("zrota", bson.M{"time": end, "dep": rota.Dep}, rota); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"success": true})
	}
}

func getRotas(c echo.Context) error {
	var dep string
	user := c.(*CustomContext).GetSession("user").(*db.User)
	if user.Role == "大队" {
		dep = "大队"
	} else {
		dep = user.Department
	}

	startTime := c.QueryParam("starttime")
	endTime := c.QueryParam("endtime")

	start := service.ToRotaTime(startTime)
	end := service.ToRotaTime(endTime)

	c.Logger().Error("====", start, end)

	rotas := []db.ZRota{}
	if err := db.FindMany("zrota", bson.D{bson.DocElem{Name: "dep", Value: dep}, bson.DocElem{Name: "time", Value: bson.M{"$gte": start}}, bson.DocElem{Name: "time", Value: bson.M{"$lte": end}}}, &rotas); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		c.Logger().Error(rotas)
		return c.JSON(http.StatusOK, rotas)
	}
}

func delArticle(c echo.Context) error {
	if err := service.DelArticle(c.QueryParam("id")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delUser(c echo.Context) error {
	if err := service.DelUser(c.QueryParam("id")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delDep(c echo.Context) error {
	if err := service.DelDep(c.QueryParam("id")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delSubject(c echo.Context) error {
	if err := db.Delete("subject", c.QueryParam("id")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delDirectory(c echo.Context) error {
	if err := service.DelDirectory(c.QueryParam("id")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delLink(c echo.Context) error {
	if err := service.DelLink(c.QueryParam("id")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delNotice(c echo.Context) error {
	if err := service.DelNotice(c.QueryParam("id")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func getArticles(c echo.Context) error {
	cond := make(bson.M, 1)
	if "" != c.QueryParam("searchValue") {
		cond["title"] = bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchValue"), Options: "ixs"}}
	}
	if "true" == c.QueryParam("isRed") {
		cond["isRed"] = true
	}
	if "false" == c.QueryParam("isPass") {
		cond["isPass"] = false
	} else if "true" == c.QueryParam("isPass") {
		cond["true"] = true
	}
	if "false" == c.QueryParam("isRed") {
		cond["isRed"] = false
	}
	if "true" == c.QueryParam("isHot") {
		cond["isHot"] = true
	}
	if "true" == c.QueryParam("isImage") {
		cond["isImage"] = true
	}
	if "true" == c.QueryParam("isTraffic") {
		cond["isTraffic"] = true
	}
	if "" != c.QueryParam("class") {
		cond["class"] = c.QueryParam("class")
	}
	if "" != c.QueryParam("category") {
		cond["category"] = c.QueryParam("category")
	}

	user := c.(*CustomContext).GetSession("user").(*db.User)
	if user.Role != "大队" {
		cond["dep"] = user.Department
	}

	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		c.Logger().Warn(err)
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetArticlesByPage(1, 15, cond), "count": service.GetArticlesCount(cond)})
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetArticlesByPage(page, 15, cond), "count": service.GetArticlesCount(cond)})
	}
}

//发布时预览
func previewArticle(c echo.Context) error {
	if c.FormValue("isRed") == "true" {
		return previewHongtouArticle(c)
	}
	article := db.Article{
		Subject:   c.FormValue("subject"),
		Title:     c.FormValue("title"),
		Creator:   c.FormValue("creator"),
		Assessor:  c.FormValue("assessor"),
		Signature: c.FormValue("signature"),
		From:      c.FormValue("from"),
		Content:   template.HTML(c.FormValue("content")),
		Category:  c.FormValue("category"),
		Pic:       c.FormValue("pic"),
		Id:        bson.NewObjectId(),
		Class:     GetClass(c.FormValue("category")),
	}
	c.(*CustomContext).SetSession("article", article)
	return c.Render(http.StatusOK, "preview", map[string]db.Any{"Auditing": false, "Article": article})
}

//发布时预览
func previewHongtouArticle(c echo.Context) error {

	str1 := `<p style="margin-top:5px;margin-bottom:5px;margin-left: 0;line-height:150%">
    <br/>
	</p>
	<p style="margin-top:5px;margin-bottom:5px;margin-left: 0;line-height:150%">
	    <br/>
	</p>
	<p style="margin-top: 5px; margin-bottom: 5px; margin-left: 0px; line-height: 37px; text-align: right;position: relative;">
	    <img src="/images/zhangzi.gif" style="position: absolute;width: 175px; height: 180px;right: 100px;top: -60px;"/> 
	 &nbsp; &nbsp;<span style="font-family: 仿宋, FangSong; font-size: 21px;padding-right:60px"> 府谷县公安局交通警察大队</span>
	</p>
	<p style="margin: 5px 0px; text-indent: 43px; line-height: 37px; text-align: right;">
	    <span style="font-size: 21px; font-family: 仿宋, FangSong;padding-right:80px">`
	str2 := `</span>
	</p>
	<p style="margin: 5px 0px; text-indent: 43px; line-height: 37px; text-align: right;">
	    <span style="font-size: 21px; font-family: 仿宋, FangSong;padding-right:80px"><br/></span>
	</p>
	<p style="margin: 5px 0px; text-indent: 43px; line-height: 37px; text-align: right;">
	    <span style="font-size: 21px; font-family: 仿宋, FangSong;padding-right:80px"><br/></span>
	</p>
	<p style="margin: 5px 0px; text-indent: 43px; line-height: 37px;">
	    <br/>
	</p>`

	year, month, day := GetChineseDate()

	article := db.Article{
		Subject:   c.FormValue("subject"),
		Title:     c.FormValue("title"),
		Creator:   c.FormValue("creator"),
		Assessor:  c.FormValue("assessor"),
		Signature: c.FormValue("signature"),
		From:      c.FormValue("from"),
		Content:   template.HTML(c.FormValue("content")),
		Attach:    template.HTML(str1 + year + "年" + month + "月" + day + "日" + str2),
		Category:  c.FormValue("category"),
		Year:      c.FormValue("year"),
		No:        c.FormValue("no"),
		Header:    c.FormValue("header"),
		IsRed:     true,
		Id:        bson.NewObjectId(),
		Class:     GetClass(c.FormValue("category")),
	}

	if c.FormValue("needSign") == "true" {
		article.NeedSign = true
	}

	c.(*CustomContext).SetSession("article", article)
	return c.Render(http.StatusOK, "hongtou", map[string]db.Any{"Auditing": false, "Article": article})
}

func previewArticleById(c echo.Context) error {
	var article db.Article
	if err := db.GetById("article", c.QueryParam("id"), &article); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		if article.IsRed {
			return c.Render(http.StatusOK, "hongtou", map[string]db.Any{"Auditing": true, "Article": article})
		} else {
			return c.Render(http.StatusOK, "preview", map[string]db.Any{"Auditing": true, "Article": article})
		}
	}
}

func viewArticle(c echo.Context) error {
	var article db.Article
	if err := db.GetById("article", c.QueryParam("id"), &article); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		if article.IsRed {
			return c.Render(http.StatusOK, "hongtou", map[string]db.Any{"Show": true, "Single": true, "Article": article})
		} else {
			return c.Render(http.StatusOK, "preview", map[string]db.Any{"Show": true, "Single": true, "Article": article})
		}
	}
}

func summarizationPage(c echo.Context) error {
	id := service.GetSummarization()
	var article db.Article
	if err := db.GetById("article", id, &article); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.Render(http.StatusOK, "preview", map[string]db.Any{"Article": article, "Show": true, "Single": true})
}

func showArticleById(c echo.Context) error {

	var article db.Article
	if err := db.GetById("article", c.QueryParam("id"), &article); err != nil {
		c.Logger().Warn(err)
		return err
	}
	if article.IsRed {
		return showHongtouArticleById(c, article)
	}

	var pre, next []db.Article
	db.UpdateById("article", c.QueryParam("id"), bson.M{"$inc": bson.M{"hits": 1}})
	if err := db.FindManyOrder("article", bson.M{"time": bson.M{"$lt": article.Time}, "isRed": false, "isPass": true, "category": article.Category}, "-time", 1, &pre); err != nil {
		c.Logger().Warn(err)
		return err
	}
	if err := db.FindManyOrder("article", bson.M{"time": bson.M{"$gt": article.Time}, "isRed": false, "isPass": true, "category": article.Category}, "time", 1, &next); err != nil {
		c.Logger().Warn(err)
		return err
	}

	data := map[string]db.Any{"Show": true, "Article": article}
	if len(pre) == 1 {
		data["Pre"] = pre[0]
	}
	if len(next) == 1 {
		data["Next"] = next[0]
	}

	return c.Render(http.StatusOK, "preview", data)
}

func showHongtouArticleById(c echo.Context, article db.Article) error {
	var pre, next []db.Article
	db.UpdateById("article", c.QueryParam("id"), bson.M{"$inc": bson.M{"hits": 1}})
	if err := db.FindManyOrder("article", bson.M{"time": bson.M{"$lt": article.Time}, "isRed": true, "isPass": true, "category": article.Category}, "-time", 1, &pre); err != nil {
		c.Logger().Warn(err)
		return err
	}
	if err := db.FindManyOrder("article", bson.M{"time": bson.M{"$gt": article.Time}, "isRed": true, "isPass": true, "category": article.Category}, "time", 1, &next); err != nil {
		c.Logger().Warn(err)
		return err
	}

	data := map[string]db.Any{"Show": true, "Article": article}
	if len(pre) == 1 {
		data["Pre"] = pre[0]
	}
	if len(next) == 1 {
		data["Next"] = next[0]
	}

	return c.Render(http.StatusOK, "hongtou", data)
}

//签收红头文件
func signArticle(c echo.Context) error {
	user := service.LoginByName(c.FormValue("name"), c.FormValue("pwd"))
	if user == nil {
		return MyRedirect(c, "/error.html")
	}
	if err := service.SignArticle(c.FormValue("id"), user.Department); err != nil {
		return MyRedirect(c, "/error.html")
	} else {
		return MyRedirect(c, "/showArticle?id="+c.FormValue("id"))
	}
}

func modifyArticlePage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)

	var article db.Article
	if err := db.GetById("article", c.QueryParam("id"), &article); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		navBar, subNav := service.CreateMenuHtml(user, article.Category)
		if article.IsRed {
			return c.Render(http.StatusOK, "publish_hongtou", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Modify": true, "Article": article, "Menu": GetClass(article.Category), "Subjects": service.GetAllSubjects()})
		} else {
			return c.Render(http.StatusOK, "publish", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Modify": true, "Article": article, "Menu": GetClass(article.Category), "Subjects": service.GetAllSubjects()})
		}
	}
}

func publishArticle(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	article := c.(*CustomContext).GetSession("article").(db.Article)
	article.Time = time.Now()
	article.Department = user.Department

	if article.IsRed {
		for _, dep := range service.GetAllDeps() {
			article.UnSign = append(article.UnSign, dep.Name)
		}
	}

	if err := db.Add("article", &article); err != nil {
		c.Logger().Warn(err)
		return c.Redirect(http.StatusMovedPermanently, "/error.html")
	} else {
		return MyRedirect(c, "/admin/page/admin?category="+article.Category)
	}
}

func auditingArticle(c echo.Context) error {
	var pass bool = false
	if c.QueryParam("result") == "pass" {
		pass = true
	}
	if err := db.UpdateById("article", c.QueryParam("id"), bson.M{"$set": bson.M{"isAuditing": true, "isPass": pass, "reason": c.QueryParam("reason")}}); err != nil {
		c.Logger().Warn(err)
		return c.Redirect(http.StatusMovedPermanently, "/error.html")
	} else {
		return MyRedirect(c, "/admin/page/dadui_admin?isRed="+c.QueryParam("isRed"))
	}
}

func modifyArticle(c echo.Context) error {
	article := db.Article{
		Subject:   c.FormValue("subject"),
		Title:     c.FormValue("title"),
		Creator:   c.FormValue("creator"),
		Assessor:  c.FormValue("assessor"),
		Signature: c.FormValue("signature"),
		From:      c.FormValue("from"),
		Content:   template.HTML(c.FormValue("content")),
		Category:  c.FormValue("category"),
		Pic:       c.FormValue("pic"),
		Year:      c.FormValue("year"),
		No:        c.FormValue("no"),
	}

	if c.FormValue("needSign") == "true" {
		article.NeedSign = true
	} else {
		article.NeedSign = false
	}

	if err := db.UpdateById("article", c.FormValue("id"), bson.M{"$set": bson.M{"subject": article.Subject, "title": article.Title, "creator": article.Creator, "assessor": article.Assessor,
		"signature": article.Signature, "from": article.From, "content": article.Content, "category": article.Category, "pic": article.Pic, "needSign": article.NeedSign, "year": article.Year, "no": article.No, "isAuditing": false, "time": time.Now()}}); err != nil {
		c.Logger().Warn(err)
		return c.Redirect(http.StatusMovedPermanently, "/error.html")
	} else {
		user := c.(*CustomContext).GetSession("user").(*db.User)
		if user.Role == "大队" {
			return MyRedirect(c, "/admin/page/dadui_admin?isRed=true")
		} else {
			return MyRedirect(c, "/admin/page/zhongdui_admin")
		}
	}

}

func setArticle(c echo.Context) error {
	update := bson.M{"category": c.FormValue("category")}
	if c.FormValue("isHot") != "" {
		update["isHot"] = true
	} else {
		update["isHot"] = false
	}
	if c.FormValue("isImage") != "" {
		update["isImage"] = true
	} else {
		update["isImage"] = false
	}
	if c.FormValue("isTraffic") != "" {
		update["isTraffic"] = true
	} else {
		update["isTraffic"] = false
	}
	if c.FormValue("isTop") != "" {
		update["isTop"] = true
	} else {
		update["isTop"] = false
	}

	if update["isHot"].(bool) {
		db.UpdateByCond("article", bson.M{"isHot": true}, bson.M{"$set": bson.M{"isHot": false}})
	}

	if update["isTop"].(bool) {
		db.UpdateByCond("article", bson.M{"isTop": true, "isTraffic": update["isTraffic"], "category": c.FormValue("category")}, bson.M{"$set": bson.M{"isTop": false}})
	}

	if err := db.UpdateById("article", c.FormValue("id"), bson.M{"$set": update}); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return MyRedirect(c, "/admin/page/dadui_admin?isRed="+c.FormValue("isRed"))
	}
}

func login(c echo.Context) error {
	if user := service.LoginByName(c.FormValue("name"), c.FormValue("pwd")); user != nil {
		c.(*CustomContext).SetSession("user", user)
		if user.Role == "大队" {
			return MyRedirect(c, "/admin/page/dadui_admin?isRed=true")
		} else {
			return MyRedirect(c, "/admin/page/zhongdui_admin")
		}

	} else {
		return c.Render(http.StatusOK, "login", map[string]bool{"error": true})
	}
}

func publishPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, c.QueryParam("category"))
	data := map[string]db.Any{"navBar": navBar, "subNav": subNav, "Article": c.(*CustomContext).GetSession("article"), "User": user, "Category": c.QueryParam("category"), "Menu": GetClass(c.QueryParam("category")), "Subjects": service.GetAllSubjects()}
	if c.QueryParam("action") == "edit" {
		data["Edit"] = true
	}
	if c.QueryParam("action") == "create" {
		data["Create"] = true
	}
	if c.QueryParam("header") != "" {
		data["Header"] = c.QueryParam("header")
	}

	if c.QueryParam("isRed") == "true" {
		return c.Render(http.StatusOK, "publish_hongtou", data)
	} else {
		return c.Render(http.StatusOK, "publish", data)
	}
}

func publishHongtouPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, c.QueryParam("category"))
	data := map[string]db.Any{"navBar": navBar, "subNav": subNav, "Article": c.(*CustomContext).GetSession("article"), "User": user, "Category": c.QueryParam("category"), "Subjects": service.GetAllSubjects()}
	if c.QueryParam("action") == "edit" {
		data["Edit"] = true
	}
	if c.QueryParam("action") == "create" {
		data["Create"] = true
	}
	return c.Render(http.StatusOK, "publish_hongtou", data)
}

func hongtouPage(c echo.Context) error {
	return c.Render(http.StatusOK, "hongtou", nil)
}

func listPage(c echo.Context) error {
	data := make(map[string]db.Any, 1)

	if c.QueryParam("class") != "" {
		data["Class"] = c.QueryParam("class")
	}
	if c.QueryParam("category") != "" {
		data["Category"] = c.QueryParam("category")
		if _, ok := data["Class"]; !ok {
			data["Class"] = GetClass(c.QueryParam("category"))
		}
	}

	data["AllCategorys"] = GetAllCategorys(data["Class"].(string))

	return c.Render(http.StatusOK, "list", data)
}

func subjectArticleListPage(c echo.Context) error {
	return c.Render(http.StatusOK, "subjectArticleList", map[string]db.Any{"Subject": c.QueryParam("subject"), "AllSubjects": service.GetAllSubjects()})
}

func noLeftListPage(c echo.Context) error {
	return c.Render(http.StatusOK, "noleftlist", map[string]db.Any{"IsLeader": c.QueryParam("isLeader"), "IsTraffic": c.QueryParam("isTraffic"), "Subject": c.QueryParam("subject"), "Dep": c.QueryParam("dep")})
}

func adminPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, c.QueryParam("category"))
	return c.Render(http.StatusOK, "admin", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Category": c.QueryParam("category"), "Menu": GetClass(c.QueryParam("category")), "IsRed": c.QueryParam("isRed"), "Header": c.QueryParam("header")})
}
func zhongduiAdminPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "后台首页")
	return c.Render(http.StatusOK, "zhongduiadmin", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "后台首页", "IsPass": c.QueryParam("isPass")})
}
func daduiAdminPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "红头文件")
	return c.Render(http.StatusOK, "daduiadmin", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Category": c.QueryParam("category"), "Menu": "文章审核", "IsRed": c.QueryParam("isRed"), "Header": c.QueryParam("header")})
}
func changePwdPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, c.QueryParam("category"))
	return c.Render(http.StatusOK, "changepwd", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "后台首页"})
}

func setArticlePage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)

	var article db.Article
	if err := db.GetById("article", c.QueryParam("id"), &article); err != nil {
		c.Logger().Warn(err)
		return err
	}
	navBar, subNav := service.CreateMenuHtml(user, article.Category)
	return c.Render(http.StatusOK, "setarticle", map[string]db.Any{"navBar": navBar, "subNav": subNav, "Clazz": clazz, "User": user, "Id": c.QueryParam("id"), "Article": article, "Menu": article.Class})
}

func saveRotaPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "值班管理")
	rota := service.GetRota()
	return c.Render(http.StatusOK, "rota", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Rota": rota, "Menu": "网站管理"})
}

func addLinkPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "链接管理")
	return c.Render(http.StatusOK, "addlink", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "网站管理", "Category": c.QueryParam("category")})
}

func linkListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "链接管理")
	return c.Render(http.StatusOK, "linklist", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "网站管理"})
}

func addNoticePage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "通知管理")
	return c.Render(http.StatusOK, "addnotice", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "网站管理", "Update": false})
}

func modifyNoticePage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	var notice db.Notice
	if err := db.GetById("notice", c.QueryParam("id"), &notice); err != nil {
		c.Logger().Warn(err)
	}
	navBar, subNav := service.CreateMenuHtml(user, "通知管理")
	return c.Render(http.StatusOK, "addnotice", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "网站管理", "Update": true, "Data": notice})
}

func noticeListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "通知管理")
	return c.Render(http.StatusOK, "noticelist", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "网站管理"})
}

func rotaListPage(c echo.Context) error {

	dep := c.QueryParam("dep")
	time := service.ToRotaTime(c.QueryParam("time"))

	var rota db.ZRota
	if err := db.FindOne("zrota", bson.M{"dep": dep, "time": time}, &rota); err != nil {
		c.Logger().Error(err)
	}

	deps := service.GetAllDeps()

	return c.Render(http.StatusOK, "rotalist", map[string]db.Any{"Dep": dep, "Time": c.QueryParam("time"), "Deps": deps, "Rota": service.RotaToHtml(rota)})
}

func indexPage(c echo.Context) error {
	rota := service.GetRota()
	links := service.GetLinks()
	hotArticle := service.GetHotArticle()
	imageArticles := service.GetImageArticles()
	subjects := service.GetAllSubjects()
	trafficArticles := service.GetTrafficArticles()
	leaderArticles := service.GetLeaderArticles()
	duChaArticles := service.GetDuChaArticles()
	starArticles := service.GetStarArticles()
	summarizationId := service.GetSummarization()

	//统计
	statistics := service.Statistics(true)
	all := service.Statistics(false)
	for _, e := range statistics {
		for _, a := range all {
			if e["_id"].(bson.M)["from"].(string) == a["_id"].(bson.M)["from"].(string) {
				e["all"] = a["count"]
			}
		}
	}

	arts := map[string][]db.Article{}
	//文章
	for k, v := range clazz {
		arts[k] = service.GetIndexArticle(k, "")
		for _, c := range v {
			arts[c] = service.GetIndexArticle(k, c)
		}
	}

	var notice bool
	if db.GetCount("notice", bson.M{"isShow": true}) > 0 {
		notice = true
	}

	return c.Render(http.StatusOK, "index", map[string]db.Any{"notice": notice, "leaderArticles": leaderArticles, "summarizationId": summarizationId, "duChaArticles": duChaArticles, "starArticles": starArticles, "arts": arts, "statistics": statistics, "now": time.Now(), "week": GetWeek(time.Now().Weekday()), "rota": rota, "links": links, "trafficArticles": trafficArticles, "hotArticle": hotArticle, "imageArticles": imageArticles, "subjects": subjects})
}

func noticePage(c echo.Context) error {
	return c.Render(http.StatusOK, "notice", service.GetNotice(c.FormValue("id")))
}

func directoryPage(c echo.Context) error {
	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.Render(http.StatusOK, "directory", map[string]db.Any{"Deps": deps})
}

func searchPage(c echo.Context) error {
	return c.Render(http.StatusOK, "search", map[string]db.Any{})
}

func articleListPage(c echo.Context) error {
	return c.Render(http.StatusOK, "articlelist", map[string]db.Any{})
}

func loginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

func addUserPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "用户管理")
	deps := service.GetAllDeps()
	return c.Render(http.StatusOK, "adduser", map[string]db.Any{"Auth": []string{}, "navBar": navBar, "subNav": subNav, "User": user, "Deps": deps, "Menu": "网站管理"})
}

func addDirectoryPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "通讯录管理")

	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}

	dn := c.QueryParam("dep")
	if dn == "搜索结果" {
		dn = deps[0].Name
	}
	return c.Render(http.StatusOK, "adddirectory", map[string]db.Any{"navBar": navBar, "subNav": subNav, "DepName": dn, "User": user, "Update": false, "Deps": deps, "Menu": "网站管理"})
}

func directoryListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	navBar, subNav := service.CreateMenuHtml(user, "通讯录管理")
	return c.Render(http.StatusOK, "directorylist", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Deps": deps, "Menu": "网站管理"})
}

func modifyDirectoryPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	navBar, subNav := service.CreateMenuHtml(user, "通讯录管理")
	return c.Render(http.StatusOK, "adddirectory", map[string]db.Any{"navBar": navBar, "subNav": subNav, "Id": c.QueryParam("id"), "User": user, "Update": true, "Deps": deps, "Name": c.QueryParam("name"), "DepName": c.QueryParam("dep"), "Job": c.QueryParam("job"), "Tel": c.QueryParam("tel"), "Phone": c.QueryParam("phone"), "Menu": "网站管理"})
}

func modifyLinkPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "链接管理")
	return c.Render(http.StatusOK, "addlink", map[string]db.Any{"navBar": navBar, "subNav": subNav, "Id": c.QueryParam("id"), "User": user, "Update": true, "Name": c.QueryParam("name"), "Category": c.QueryParam("category"), "Url": c.QueryParam("url"), "Menu": "网站管理"})
}

func depListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "部门管理")
	return c.Render(http.StatusOK, "deplist", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "网站管理"})
}

func subjectListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "专题管理")
	return c.Render(http.StatusOK, "subjectlist", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Menu": "网站管理"})
}

func addDepPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "部门管理")
	return c.Render(http.StatusOK, "adddep", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Update": false, "Menu": "网站管理"})
}

func addSubjectPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "专题管理")
	return c.Render(http.StatusOK, "addsubject", map[string]db.Any{"navBar": navBar, "subNav": subNav, "User": user, "Update": false, "Menu": "网站管理"})
}

func modifySubjectPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "专题管理")
	return c.Render(http.StatusOK, "addsubject", map[string]db.Any{"navBar": navBar, "subNav": subNav, "Id": c.QueryParam("id"), "User": user, "Modify": true, "Name": c.QueryParam("name"), "Pic": c.QueryParam("pic"), "IsHot": c.QueryParam("isHot"), "Menu": "网站管理"})
}

func modifyDepPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "部门管理")
	return c.Render(http.StatusOK, "adddep", map[string]db.Any{"navBar": navBar, "subNav": subNav, "Id": c.QueryParam("id"), "User": user, "Update": true, "Name": c.QueryParam("name"), "Menu": "网站管理"})
}

func modifyUserPage(c echo.Context) error {

	data := make(map[string]db.Any, 4)
	data["Update"] = true
	data["Dep"] = c.QueryParam("dep")
	data["Id"] = c.QueryParam("id")
	data["Deps"] = service.GetAllDeps()
	switch c.QueryParam("role") {
	case "大队":
		data["DD"] = true
	case "中队":
		data["ZD"] = true
	case "部门":
		data["BM"] = true
	}

	var u db.User
	if err := db.GetById("user", c.QueryParam("id"), &u); err != nil {
		c.Logger().Error(err)
		return err
	}
	data["Auth"] = u.Authorities
	data["Menu"] = "网站管理"
	user := c.(*CustomContext).GetSession("user").(*db.User)
	data["User"] = user
	navBar, subNav := service.CreateMenuHtml(user, "用户管理")
	data["navBar"] = navBar
	data["subNav"] = subNav
	return c.Render(http.StatusOK, "adduser", data)
}

func modifyUser(c echo.Context) error {
	params, _ := c.FormParams()
	if c.FormValue("role") == "中队" {
		params["quanxian"] = append(params["quanxian"], "后台首页")
	}
	if err := db.UpdateById("user", c.FormValue("id"), bson.M{"$set": bson.M{"dep": c.FormValue("dep"), "role": c.FormValue("role"), "authorities": params["quanxian"]}}); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return userListPage(c)
	}
}

func modifyNotice(c echo.Context) error {
	if err := db.UpdateById("notice", c.FormValue("id"), bson.M{"$set": bson.M{"dep": c.FormValue("dep"), "title": c.FormValue("title"), "content": c.FormValue("content")}}); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return MyRedirect(c, "/admin/page/notice_list")
	}
}

func setNotice(c echo.Context) error {
	db.UpdateByCond("notice", bson.M{"isShow": true}, bson.M{"$set": bson.M{"isShow": false}})
	if err := db.UpdateById("notice", c.FormValue("id"), bson.M{"$set": bson.M{"isShow": true}}); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"success": true})
	}
}

func cancelNotice(c echo.Context) error {
	if err := db.UpdateById("notice", c.FormValue("id"), bson.M{"$set": bson.M{"isShow": false}}); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"success": true})
	}
}

func haveNotice(c echo.Context) error {
	if db.GetCount("notice", bson.M{"isShow": true}) > 0 {
		return c.JSON(http.StatusOK, map[string]db.Any{"have": true})
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"have": false})
	}
}

func changePwd(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	if err := service.ChangePwd(user.Name, c.FormValue("oldpass"), c.FormValue("newpass1")); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return MyRedirect(c, "/admin/logout")
	}
}

func noPassCount(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	return c.JSON(http.StatusOK, map[string]db.Any{"success": true, "count": db.GetCount("article", bson.M{"dep": user.Department, "isPass": false})})
}

func modifyDep(c echo.Context) error {
	if order, err := strconv.Atoi(c.FormValue("sort")); err == nil {
		count := service.GetDepsCount()
		var oldDep db.Department
		if err := db.GetById("department", c.FormValue("id"), &oldDep); err != nil {
			c.Logger().Warn(err)
			return err
		}

		if order >= count {
			db.UpdateByCond("department", bson.M{"order": bson.M{"$gt": oldDep.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateById("department", c.FormValue("id"), bson.M{"$set": bson.M{"order": count}})
		} else if order > 0 { //
			c.Logger().Warn("oldordder: ", oldDep.Order, ";order: ", order)
			db.UpdateByCond("department", bson.M{"order": bson.M{"$gt": oldDep.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateByCond("department", bson.M{"order": bson.M{"$gte": order}}, bson.M{"$inc": bson.M{"order": 1}})
			db.UpdateById("department", c.FormValue("id"), bson.M{"$set": bson.M{"order": order}})
		}
	}
	if err := db.UpdateById("department", c.FormValue("id"), bson.M{"$set": bson.M{"name": c.FormValue("depname")}}); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return directoryListPage(c)
}

func modifySubject(c echo.Context) error {

	if c.FormValue("isHot") == "true" {
		db.UpdateByCond("subject", bson.M{"isHot": true}, bson.M{"$set": bson.M{"isHot": false}})
		if err := db.UpdateById("subject", c.FormValue("id"), bson.M{"$set": bson.M{"name": c.FormValue("subjectName"), "pic": c.FormValue("pic"), "isHot": true}}); err != nil {
			c.Logger().Warn(err)
			return err
		}
	} else {
		if err := db.UpdateById("subject", c.FormValue("id"), bson.M{"$set": bson.M{"name": c.FormValue("subjectName"), "pic": c.FormValue("pic"), "isHot": false}}); err != nil {
			c.Logger().Warn(err)
			return err
		}
	}
	return MyRedirect(c, "/admin/page/subject_list")
}

func modifyDirectory(c echo.Context) error {
	if order, err := strconv.Atoi(c.FormValue("sort")); err == nil {
		count := service.GetDirectorysCount(c.FormValue("dep"))
		var oldDir db.Directory
		if err := db.GetById("directory", c.FormValue("id"), &oldDir); err != nil {
			c.Logger().Warn(err)
			return err
		}

		if order >= count {
			db.UpdateByCond("directory", bson.M{"dep": c.FormValue("dep"), "order": bson.M{"$gt": oldDir.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateById("directory", c.FormValue("id"), bson.M{"$set": bson.M{"order": count}})
		} else if order > 0 { //
			db.UpdateByCond("directory", bson.M{"dep": c.FormValue("dep"), "order": bson.M{"$gt": oldDir.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateByCond("directory", bson.M{"dep": c.FormValue("dep"), "order": bson.M{"$gte": order}}, bson.M{"$inc": bson.M{"order": 1}})
			db.UpdateById("directory", c.FormValue("id"), bson.M{"$set": bson.M{"order": order}})
		}

	}
	if err := db.UpdateById("directory", c.FormValue("id"), bson.M{"$set": bson.M{"name": c.FormValue("directoryName"), "dep": c.FormValue("dep"), "job": c.FormValue("job"), "tel": c.FormValue("tel"), "phone": c.FormValue("phone")}}); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return MyRedirect(c, "/admin/page/dep_list")
}

func modifyLink(c echo.Context) error {
	if order, err := strconv.Atoi(c.FormValue("sort")); err == nil {
		count := service.GetLinksCount(c.FormValue("category"))
		var oldDir db.Link
		if err := db.GetById("link", c.FormValue("id"), &oldDir); err != nil {
			c.Logger().Warn(err)
			return err
		}

		if order >= count {
			db.UpdateByCond("link", bson.M{"category": c.FormValue("category"), "order": bson.M{"$gt": oldDir.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateById("link", c.FormValue("id"), bson.M{"$set": bson.M{"order": count}})
		} else if order > 0 { //
			db.UpdateByCond("link", bson.M{"category": c.FormValue("category"), "order": bson.M{"$gt": oldDir.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateByCond("link", bson.M{"category": c.FormValue("category"), "order": bson.M{"$gte": order}}, bson.M{"$inc": bson.M{"order": 1}})
			db.UpdateById("link", c.FormValue("id"), bson.M{"$set": bson.M{"order": order}})
		}
	}
	if err := db.UpdateById("link", c.FormValue("id"), bson.M{"$set": bson.M{"name": c.FormValue("linkName"), "category": c.FormValue("category"), "url": c.FormValue("url")}}); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return MyRedirect(c, "/admin/page/link_list")
}

func userListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	navBar, subNav := service.CreateMenuHtml(user, "用户管理")
	users := make([]db.User, 10)
	if err := db.FindMany("user", nil, &users); err != nil {
		c.Logger().Warn(err)
		return err
	} else {
		return c.Render(http.StatusOK, "userlist", map[string]db.Any{"navBar": navBar, "subNav": subNav, "users": users, "User": user, "Menu": "网站管理"})
	}
}

func getUserList(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetUsersByPage(page), "count": service.GetUsersCount()})
	}
}

func getDepList(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetDepsByPage(page), "count": service.GetDepsCount()})
	}
}

func getSubjectList(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetSubjectsByPage(page), "count": service.GetSubjectsCount()})
	}
}

func getNoticeList(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetNoticesByPage(page), "count": service.GetNoticesCount()})
	}
}

func getDirectoryList(c echo.Context) error {
	data := service.GetDirectorysByName(c.QueryParam("searchValue"))
	return c.JSON(http.StatusOK, map[string]db.Any{"data": data, "count": len(data)})

}

func getLinkList(c echo.Context) error {
	data := service.GetLinksByName(c.QueryParam("searchValue"))
	return c.JSON(http.StatusOK, map[string]db.Any{"data": data, "count": len(data)})

}

func getDirectoryListByDepName(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetDirectorysByDep(c.QueryParam("depName"), page), "count": service.GetDirectorysCount(c.QueryParam("depName"))})
	}
}

func getLinkListByDepName(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetLinksByDep(c.QueryParam("depName"), page), "count": service.GetLinksCount(c.QueryParam("depName"))})
	}
}

func getDirectoryByName(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder("directory", bson.M{"dep": c.QueryParam("depName")}, "order", 0, &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func getLinkByName(c echo.Context) error {
	var deps []db.Link = make([]db.Link, 5)
	if err := db.FindManyOrder("link", bson.M{"category": c.QueryParam("depName")}, "order", 0, &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByName(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder("directory", bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByJob(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder("directory", bson.M{"job": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByTel(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder("directory", bson.M{"tel": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByPhone(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder("directory", bson.M{"phone": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		c.Logger().Warn(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchArticle(c echo.Context) error {
	cond := bson.M{"isPass": true}

	if "" != c.QueryParam("searchValue") {
		cond["title"] = bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchValue"), Options: "ixs"}}
	}

	if "" != c.QueryParam("category") {
		cond["category"] = c.QueryParam("category")
	}

	if "" != c.QueryParam("class") {
		cond["class"] = c.QueryParam("class")
	}

	if "" != c.QueryParam("dep") {
		cond["dep"] = c.QueryParam("dep")
	}

	if "true" == c.QueryParam("isTraffic") {
		cond["isTraffic"] = true
	}

	if "true" == c.QueryParam("isRed") {
		cond["isRed"] = true
	}

	if "true" == c.QueryParam("isImage") {
		cond["isImage"] = true
	}

	if "" != c.QueryParam("subject") {
		if "no" == c.QueryParam("subject") {
			cond["subject"] = bson.M{"$ne": "不属于专题稿件"}
		} else {
			cond["subject"] = c.QueryParam("subject")
		}

	}

	c.Logger().Print(cond)

	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetIndexArticlesByPage(1, 15, cond), "count": service.GetArticlesCount(cond)})
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetIndexArticlesByPage(page, 15, cond), "count": service.GetArticlesCount(cond)})
	}

}

func searchPhone(c echo.Context) error {
	var d db.Directory
	if err := db.FindOne("directory", bson.M{"dep": c.FormValue("dep"), "name": c.FormValue("name")}, &d); err != nil {
		return c.JSON(http.StatusOK, map[string]db.Any{"success": false})
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"success": true, "data": d})
	}
}

func statisticsPage(c echo.Context) error {
	return c.JSON(http.StatusOK, service.Statistics(false))
}

func uploadImage(c echo.Context) error {
	file, err := c.FormFile("upfile")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	fileName := "/user/images/" + strconv.FormatInt(time.Now().UnixNano(), 10) + file.Filename
	dst, err := os.Create("public" + fileName)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"state": "SUCCESS", "url": fileName, "title": file.Filename, "origin": file.Filename})
}

func uploadFile(c echo.Context) error {
	file, err := c.FormFile("upfile")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	fileName := "/user/files/" + strconv.FormatInt(time.Now().UnixNano(), 10) + file.Filename
	dst, err := os.Create("public" + fileName)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"state": "SUCCESS", "url": fileName, "title": file.Filename, "original": file.Filename, "fileType": "doc", "type": "doc"})
}

func imageManager(c echo.Context) error {
	if fs, err := ioutil.ReadDir("public/user/images"); err != nil {
		return err
	} else {
		var str string
		for _, f := range fs {
			str += ("/user/images/" + f.Name() + "ue_separate_ue")
		}
		return c.String(http.StatusOK, str)
	}
}

//获取请求中的JSON对象
func fromJsonReq(c echo.Context, dataPointer interface{}) error {
	if jsonBlob, err := ioutil.ReadAll(c.Request().Body); err != nil {
		c.Logger().Print("error:", err)
		return err
	} else {
		if err = json.Unmarshal(jsonBlob, dataPointer); err != nil {
			c.Logger().Print("error:", err)
			return err
		}
	}
	return nil
}

func MyRedirect(c echo.Context, url string) error {
	header := c.Response().Header()
	header["Pragma"] = []string{"no-cache"}
	header["Cache-Control"] = []string{"no-cache"}
	header["Expires"] = []string{"0"}
	return c.Redirect(http.StatusMovedPermanently, url)
}

func logout(c echo.Context) error {
	c.(*CustomContext).Destroy()
	return MyRedirect(c, "/login.html")
}

func Add(a, b int) int {
	return a + b
}

func Two(a int) bool {
	return a > 0 && (a%2 != 0)
}

func Ten(a int) bool {
	return a > 0 && (a%9 == 0)
}

func GetClass(category string) string {
	for k, v := range clazz {
		for _, c := range v {
			if c == category {
				return k
			}
		}
	}
	return ""
}

func GetAllCategorys(class string) []string {
	return clazz[class]
}

func AddSpace(s string) template.HTML {
	str := []rune(s)
	if len(str) == 2 {
		return template.HTML(string(str[0]) + "&emsp;" + string(str[1]))
	} else {
		return template.HTML(s)
	}
}

func GetChineseDate() (string, string, string) {
	now := time.Now()
	var year, month, day string
	for _, v := range itoa(now.Year()) {
		year += getword(string(v))
	}

	if now.Month()+1 == 10 {
		month = "十"
	} else if now.Month()+1 > 10 {
		month = "十" + getword(string(itoa(int(now.Month()))[1]))
	} else {
		month = getword(itoa(int(now.Month())))
	}

	if now.Day() < 10 {
		day = getword(itoa(now.Day()))
	} else if now.Day() < 20 {
		day = "十" + getword(string(itoa(now.Day())[1]))
	} else {

		if getword(string(itoa(now.Day())[1])) != "〇" {
			day = getword(string(itoa(now.Day())[0])) + "十" + getword(string(itoa(now.Day())[1]))
		} else {
			day = getword(string(itoa(now.Day())[0])) + "十"
		}
	}

	return year, month, day
}
func getword(s string) string {
	var str string
	switch s {
	case "1":
		str = "一"
	case "2":
		str = "二"
	case "3":
		str = "三"
	case "4":
		str = "四"
	case "5":
		str = "五"
	case "6":
		str = "六"
	case "7":
		str = "七"
	case "8":
		str = "八"
	case "9":
		str = "九"
	case "0":
		str = "〇"

	}
	return str
}

func itoa(s int) string {
	d := strconv.Itoa(s)
	return d
}

func GetWeek(w time.Weekday) string {
	var s string
	switch w {
	case time.Sunday:
		s = "星期日"
	case time.Monday:
		s = "星期一"
	case time.Tuesday:
		s = "星期二"
	case time.Saturday:
		s = "星期六"
	case time.Thursday:
		s = "星期四"
	case time.Wednesday:
		s = "星期三"
	case time.Friday:
		s = "星期五"
	}
	return s
}

func Substring(s string, l int) string {
	if len([]rune(s)) > l {
		return string([]rune(s)[:l])
	} else {
		return s
	}
}

func IsNew(t time.Time) bool {
	now := time.Now()
	if t.Year() == now.Year() && t.Month() == now.Month() && (now.Day()-t.Day() < 2) {
		return true
	} else {
		return false
	}
}

func Include(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
