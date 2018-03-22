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
	"log"
	"time"
)

/*
	创建标准的RESTful  api
*/

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
			//res := strings.Split(c.Path(), "/")[2]
			if _, o := cc.GetSession("user").(*db.User); o {
				/*
					for _, f := range user.Functions {
						for _, p := range f.Permissions {
							if p.Resource == res {
								for _, m := range p.Methods {
									if m == c.Request().Method {
										return next(c)
									}
								}
							}
						}
					}*/
				return next(c)
			}
			return c.Redirect(http.StatusMovedPermanently, "/login.html")
		}
	}
}

func okGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	e := echo.New()

	temp := template.New("jjdd")
	temp.Delims("<%", "%>")

	funmap := make(template.FuncMap, 1)
	funmap["Two"] = Two
	temp.Funcs(funmap)
	t := &Template{
		templates: template.Must(temp.ParseGlob("views/*.html")),
	}
	e.Renderer = t
	e.Static("/", "public")

	//日志
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost"},
		AllowCredentials: true,
	}))

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
	admin.GET("/page/preview", previewPage)
	admin.GET("/page/hongtou", hongtouPage)
	admin.GET("/page/admin", adminPage)
	admin.GET("/page/modifyPage", modifyArticlePage)
	admin.GET("/page/add_user", addUserPage)
	admin.GET("/page/user_list", userListPage)
	admin.GET("/page/modify_user", modifyUserPage)

	admin.GET("/page/dep_list", depListPage)
	admin.GET("/page/add_dep", addDepPage)
	admin.GET("/page/modify_dep", modifyDepPage)

	admin.GET("/page/add_directory", addDirectoryPage)
	admin.GET("/page/directory_list", directoryListPage)
	admin.GET("/page/modify_directory", modifyDirectoryPage)

	admin.GET("/page/saveRota", saveRotaPage)

	/*----------------操作--------*/
	//文章
	admin.Any("/publish", publishArticle)
	admin.Any("/preview", previewArticle)
	admin.Any("/previewById", previewArticleById)
	admin.Any("/getArticles", getArticles)
	admin.Any("/delArticle", delArticle)
	admin.Any("/auditing", auditingArticle)
	admin.Any("/modify", modifyArticle)
	//用户
	admin.Any("/addUser", addUser)
	admin.Any("/getUserList", getUserList)
	admin.Any("/delUser", delUser)
	admin.Any("/modifyUser", modifyUser)
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

	e.Any("/getDirectoryByname1", getDirectoryByName)
	e.Any("/searchDirectoryByName", searchDirectoryByName)
	e.Any("/searchDirectoryByJob", searchDirectoryByJob)
	e.Any("/searchDirectoryByTel", searchDirectoryByTel)
	e.Any("/searchDirectoryByPhone", searchDirectoryByPhone)

	e.GET("/directory", directoryPage)
	e.GET("/", indexPage)
	e.GET("/list.html", listPage)
	e.GET("/login.html", loginPage)
	e.Any("/login", login)

	//登录
	//e.POST("/login", login)

	//文件上传
	e.POST("/uploadImage", uploadImage)
	e.POST("/uploadFile", uploadFile)
	e.POST("/imageManager", imageManager)
	//处理微信支付回调
	//e.Post("/mch", weixin.MchServer)
	e.Start(":80")

}

func addUser(c echo.Context) error {
	user := db.User{
		Id:         bson.NewObjectId(),
		Name:       c.FormValue("name"),
		Password:   "888888",
		Department: c.FormValue("dep"),
		Role:       c.FormValue("role"),
	}
	if err := service.AddUser(user); err != nil {
		log.Println(err)
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
		log.Println(err)
		return c.Render(http.StatusOK, "adddep", map[string]bool{"NameError": true})
	} else {
		return MyRedirect(c, "/admin/page/dep_list")
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
		log.Println(err)
		return c.Render(http.StatusOK, "adddirectory", nil)
	} else {
		return MyRedirect(c, "/admin/page/directory_list")
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
	if err := db.Upsert("rota", rota); err != nil {
		log.Println(err)
		return err
	} else {
		return MyRedirect(c, "/admin/page/admin")
	}
}

func delArticle(c echo.Context) error {
	if err := service.DelArticle(c.QueryParam("id")); err != nil {
		log.Println(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delUser(c echo.Context) error {
	if err := service.DelUser(c.QueryParam("id")); err != nil {
		log.Println(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delDep(c echo.Context) error {
	if err := service.DelDep(c.QueryParam("id")); err != nil {
		log.Println(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func delDirectory(c echo.Context) error {
	if err := service.DelDirectory(c.QueryParam("id")); err != nil {
		log.Println(err)
		return err
	} else {
		return c.NoContent(http.StatusOK)
	}
}

func getArticles(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetArticlesByPage(1, c.QueryParam("searchValue")), "count": service.GetArticlesCount(c.QueryParam("searchValue"))})
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetArticlesByPage(page, c.QueryParam("searchValue")), "count": service.GetArticlesCount(c.QueryParam("searchValue"))})
	}
}

//发布时预览
func previewArticle(c echo.Context) error {
	/*
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
	*/
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
	}
	c.(*CustomContext).SetSession("article", article)
	return c.Render(http.StatusOK, "preview", article)
}

func previewArticleById(c echo.Context) error {
	var article db.Article
	if err := db.GetById("article", c.QueryParam("id"), &article); err != nil {
		log.Println(err)
		return err
	} else {
		return c.Render(http.StatusOK, "preview", map[string]db.Any{"Auditing": true, "Article": article})
	}
}

func modifyArticlePage(c echo.Context) error {
	var article db.Article
	if err := db.GetById("article", c.QueryParam("id"), &article); err != nil {
		log.Println(err)
		return err
	} else {
		return c.Render(http.StatusOK, "publish", map[string]db.Any{"Modify": true, "Article": article})
	}
}

func publishArticle(c echo.Context) error {
	article := c.(*CustomContext).GetSession("article").(db.Article)
	article.Time = time.Now()
	if err := db.Add("article", &article); err != nil {
		log.Println(err)
		return c.Redirect(http.StatusMovedPermanently, "/error.html")
	} else {
		return c.String(http.StatusOK, "add ok")
	}
}

func auditingArticle(c echo.Context) error {
	var pass bool = false
	if c.QueryParam("result") == "pass" {
		pass = true
	}
	if err := db.UpdateById("article", c.QueryParam("id"), bson.M{"$set": bson.M{"isAuditing": pass}}); err != nil {
		log.Println(err)
		return c.Redirect(http.StatusMovedPermanently, "/error.html")
	} else {
		return MyRedirect(c, "/admin/page/admin")
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
	}

	if err := db.UpdateById("article", c.FormValue("id"), bson.M{"$set": bson.M{"subject": article.Subject, "title": article.Title, "creator": article.Creator, "assessor": article.Assessor,
		"signature": article.Signature, "from": article.From, "content": article.Content, "category": article.Category, "pic": article.Pic}}); err != nil {
		log.Println(err)
		return c.Redirect(http.StatusMovedPermanently, "/error.html")
	} else {
		return MyRedirect(c, "/admin/page/admin")
	}

}

func login(c echo.Context) error {
	if user := service.LoginByName(c.FormValue("name"), c.FormValue("pwd")); user != nil {
		c.(*CustomContext).SetSession("user", user)
		return MyRedirect(c, "/admin/page/admin")
	} else {
		return c.Render(http.StatusOK, "login", map[string]bool{"error": true})
	}
}

func publishPage(c echo.Context) error {
	return c.Render(http.StatusOK, "publish", map[string]db.Any{"Modify": false, "Article": c.(*CustomContext).GetSession("article")})
}

func publishHongtouPage(c echo.Context) error {
	return c.Render(http.StatusOK, "publish_hongtou", c.(*CustomContext).GetSession("article"))
}

func previewPage(c echo.Context) error {
	return c.Render(http.StatusOK, "preview", nil)
}

func hongtouPage(c echo.Context) error {
	return c.Render(http.StatusOK, "hongtou", nil)
}

func listPage(c echo.Context) error {
	return c.Render(http.StatusOK, "list", nil)
}
func adminPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	return c.Render(http.StatusOK, "admin", map[string]db.Any{"User": user})
}

func saveRotaPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	rota, have := service.GetRota()
	return c.Render(http.StatusOK, "rota", map[string]db.Any{"User": user, "First": !have, "Rota": rota})
}

func indexPage(c echo.Context) error {
	rota, _ := service.GetRota()
	return c.Render(http.StatusOK, "index", rota)
}

func directoryPage(c echo.Context) error {
	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.Render(http.StatusOK, "directory", map[string]db.Any{"Deps": deps})
}

func loginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

func addUserPage(c echo.Context) error {
	return c.Render(http.StatusOK, "adduser", nil)
}

func addDirectoryPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)

	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		log.Println(err)
		return err
	}

	return c.Render(http.StatusOK, "adddirectory", map[string]db.Any{"User": user, "Update": false, "Deps": deps})
}

func directoryListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.Render(http.StatusOK, "directorylist", map[string]db.Any{"User": user, "Deps": deps})
}

func modifyDirectoryPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	var deps []db.Department = make([]db.Department, 10)
	if err := db.GetAll("department", &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.Render(http.StatusOK, "adddirectory", map[string]db.Any{"Id": c.QueryParam("id"), "User": user, "Update": true, "Deps": deps, "Name": c.QueryParam("name"), "DepName": c.QueryParam("dep"), "Job": c.QueryParam("job"), "Tel": c.QueryParam("tel"), "Phone": c.QueryParam("phone")})
}

func depListPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	return c.Render(http.StatusOK, "deplist", map[string]db.Any{"User": user})
}

func addDepPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	return c.Render(http.StatusOK, "adddep", map[string]db.Any{"User": user, "Update": false})
}

func modifyDepPage(c echo.Context) error {
	user := c.(*CustomContext).GetSession("user").(*db.User)
	return c.Render(http.StatusOK, "adddep", map[string]db.Any{"Id": c.QueryParam("id"), "User": user, "Update": true, "Name": c.QueryParam("name")})
}

func modifyUserPage(c echo.Context) error {

	data := make(map[string]db.Any, 4)
	data["Update"] = true
	data["Dep"] = c.QueryParam("dep")
	data["Id"] = c.QueryParam("id")
	switch c.QueryParam("role") {
	case "大队":
		data["DD"] = true
	case "中队":
		data["ZD"] = true
	case "部门":
		data["BM"] = true
	}

	return c.Render(http.StatusOK, "adduser", data)
}

func modifyUser(c echo.Context) error {
	if err := db.UpdateById("user", c.FormValue("id"), bson.M{"$set": bson.M{"dep": c.FormValue("dep"), "role": c.FormValue("role")}}); err != nil {
		log.Println(err)
		return err
	} else {
		return userListPage(c)
	}
}

func modifyDep(c echo.Context) error {

	if order, err := strconv.Atoi(c.FormValue("sort")); err != nil {
		log.Println(err)
		return errors.New("排序必需为数字")
	} else {

		if err := db.UpdateById("department", c.FormValue("id"), bson.M{"$set": bson.M{"name": c.FormValue("depname")}}); err != nil {
			log.Println(err)
			return err
		}

		count := service.GetDepsCount()
		var oldDep db.Department
		if err := db.GetById("department", c.FormValue("id"), &oldDep); err != nil {
			log.Println(err)
			return err
		}

		if order >= count {
			db.UpdateByCond("department", bson.M{"order": bson.M{"$gt": oldDep.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateById("department", c.FormValue("id"), bson.M{"$set": bson.M{"order": count}})
		} else if order > 0 { //
			log.Println("oldordder: ", oldDep.Order, ";order: ", order)
			db.UpdateByCond("department", bson.M{"order": bson.M{"$gt": oldDep.Order}}, bson.M{"$inc": bson.M{"order": -1}})
			db.UpdateByCond("department", bson.M{"order": bson.M{"$gte": order}}, bson.M{"$inc": bson.M{"order": 1}})
			db.UpdateById("department", c.FormValue("id"), bson.M{"$set": bson.M{"order": order}})
		}

		return directoryListPage(c)

	}
}

func modifyDirectory(c echo.Context) error {
	if order, err := strconv.Atoi(c.FormValue("sort")); err != nil {
		log.Println(err)
		return errors.New("排序必需为数字")
	} else {

		if err := db.UpdateById("directory", c.FormValue("id"), bson.M{"$set": bson.M{"name": c.FormValue("directoryName"), "dep": c.FormValue("dep"), "job": c.FormValue("job"), "tel": c.FormValue("tel"), "phone": c.FormValue("phone")}}); err != nil {
			log.Println(err)
			return err
		}

		count := service.GetDirectorysCount(c.FormValue("dep"))
		var oldDir db.Directory
		if err := db.GetById("directory", c.FormValue("id"), &oldDir); err != nil {
			log.Println(err)
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

		return MyRedirect(c, "/admin/page/dep_list")

	}
}

func userListPage(c echo.Context) error {
	users := make([]db.User, 10)
	if err := db.FindMany("user", nil, &users); err != nil {
		log.Println(err)
		return err
	} else {
		return c.Render(http.StatusOK, "userlist", users)
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

func getDirectoryList(c echo.Context) error {

	return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetDirectorysByName(c.QueryParam("searchValue"))})

}

func getDirectoryListByDepName(c echo.Context) error {
	if page, err := strconv.Atoi(c.QueryParam("page")); err != nil {
		return err
	} else {
		return c.JSON(http.StatusOK, map[string]db.Any{"data": service.GetDirectorysByDep(c.QueryParam("depName"), page), "count": service.GetDirectorysCount(c.QueryParam("depName"))})
	}
}

func getDirectoryByName(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder1("directory", bson.M{"dep": c.QueryParam("depName")}, "order", 0, &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByName(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder1("directory", bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByJob(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder1("directory", bson.M{"job": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByTel(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder1("directory", bson.M{"tel": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
}

func searchDirectoryByPhone(c echo.Context) error {
	var deps []db.Directory = make([]db.Directory, 5)
	if err := db.FindManyOrder1("directory", bson.M{"phone": bson.M{"$regex": bson.RegEx{Pattern: c.QueryParam("searchKeyWord"), Options: "ixs"}}}, "order", 0, &deps); err != nil {
		log.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, map[string]db.Any{"data": deps})
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
	return c.JSON(http.StatusOK, map[string]string{"state": "SUCCESS", "url": fileName, "title": file.Filename, "origin": file.Filename})
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

func Two(a int) bool {
	return a > 0 && (a%2 != 0)
}
