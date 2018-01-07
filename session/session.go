package session

import (
	"crypto/md5"
	"fmt"
	"github.com/labstack/echo"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var context map[string]*Session
var timeout int32
var reset chan string

type Session struct {
	Values    map[string]interface{}
	sessionId string
	mu        sync.RWMutex
	expire    int32
}

func GetSession(c echo.Context) *Session {
	cookie, err := c.Cookie("sessionId")
	if err != nil {
		return newSession(c)
	}

	session := context[cookie.Value]
	if nil == session {
		return newSession(c)
	} else {
		return session
	}
}

func GetSessionFromRequest(r *http.Request) *Session {
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return nil
	}
	return context[cookie.Value]
}

func (session *Session) GetId() string {
	return session.sessionId
}

func (session *Session) GetValue(key string) interface{} {
	session.mu.RLock()
	defer session.mu.RUnlock()
	value := session.Values[key]
	return value
}

func (session *Session) PutValue(key string, value interface{}) {
	session.mu.Lock()
	session.Values[key] = value
	defer session.mu.Unlock()
}

func newSession(c echo.Context) *Session {
	id := getSessionId()
	cookie := new(http.Cookie)
	cookie.Name = "sessionId"
	cookie.Value = id
	cookie.Path = "/"
	session := new(Session)
	session.Values = make(map[string]interface{}, 1)
	session.expire = timeout
	session.sessionId = id
	context[id] = session
	c.SetCookie(cookie)
	return session
}

func getSessionId() string {
	nano := time.Now().UnixNano()
	rand.Seed(nano)
	rndNum := rand.Int63()
	return Md5(Md5(strconv.FormatInt(nano, 10)) + Md5(strconv.FormatInt(rndNum, 10)))
}

func Md5(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

func start() {
	c := time.Tick(5 * time.Second)
	for {
		select {
		case id := <-reset:
			if nil != context[id] {
				context[id].expire = timeout
			}
		case <-c:
			for key, value := range context {
				value.expire -= 5
				if value.expire <= 0 {
					delete(context, key)
				}
			}
		}
	}
}

func Reset(c echo.Context) {
	cookie, err := c.Cookie("sessionId")
	if err != nil {
		return
	}
	reset <- cookie.Value
}

func InitContext(t int32) {
	context = make(map[string]*Session, 10)
	timeout = t
	reset = make(chan string)
	go start()
}
