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
	"jjdd/session"
	"math"
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
			return c.JSON(http.StatusUnauthorized, Resp{Error: "没有权限"})
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

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = t
	e.Static("/", "public")

	//日志
	e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost"},
		AllowCredentials: true,
	}))

	//Session控制
	e.Use(sessionHandler)
	//最大失效时间相当于没有限制
	session.InitContext(math.MaxInt32)

	//检查RESTful权限
	admin := e.Group("/admin")
	//admin.Use(apiGroup)
	admin.GET("/page/publish", publishPage)
	admin.GET("/page/preview", previewPage)

	admin.Any("/publish", publishArticle)
	admin.POST("/preview", previewArticle)

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

func previewArticle(c echo.Context) error {
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

	data := map[string]db.Any{
		"subject":   c.FormValue("subject"),
		"title":     c.FormValue("title"),
		"creator":   c.FormValue("creator"),
		"assessor":  c.FormValue("assessor"),
		"signature": c.FormValue("signature"),
		"from":      c.FormValue("from"),
		"content":   template.HTML(c.FormValue("content")),
		"category":  c.FormValue("category"),
		"pic":       fileName,
		"id":        bson.NewObjectId().Hex(),
	}
	article := db.ArticleFromMap(data)
	c.(*CustomContext).SetSession("article", article)
	return c.Render(http.StatusOK, "preview", data)
}

func publishArticle(c echo.Context) error {
	article := c.(*CustomContext).GetSession("article")
	if err := db.Add("article", &article); err != nil {
		log.Println(err)
		return c.Redirect(http.StatusOK, "/error.html")
	} else {
		return c.String(http.StatusOK, "add ok")
	}
}

func publishPage(c echo.Context) error {
	return c.Render(http.StatusOK, "publish", c.(*CustomContext).GetSession("article"))
}

func previewPage(c echo.Context) error {
	return c.Render(http.StatusOK, "preview", nil)
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
