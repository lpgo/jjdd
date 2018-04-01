package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//

type Any interface{}

const URL = "127.0.0.1:27017" //mongodb连接字符串

var (
	mgoSession *mgo.Session
	dataBase   = "jjdd"
)

type Ids struct {
	Id bson.ObjectId `bson:"_id" json:"id"`
	N  int64         `bson:"n" json:"n"`
}

/**
 * 公共方法，获取session，如果存在则拷贝一份
 */
func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(URL)
		if err != nil {
			panic(err) //直接终止程序运行
		}
	}
	//最大连接池默认为4096
	return mgoSession.Clone()
}

//公共方法，获取collection对象
func WitchCollection(collection string, s func(*mgo.Collection) error) error {
	session := getSession()
	defer session.Close()
	c := session.DB(dataBase).C(collection)
	return s(c)
}

func GetAll(collection string, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		return c.Find(nil).All(dataPointer)
	}
	return WitchCollection(collection, query)
}

func GetAllByOrder(collection string, condition Any, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		return c.Find(condition).Sort("order").All(dataPointer)
	}
	return WitchCollection(collection, query)
}

func GetById(collection, id string, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		return c.FindId(bson.ObjectIdHex(id)).One(dataPointer)
	}
	return WitchCollection(collection, query)
}

func FindMany(collection string, condition Any, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		return c.Find(condition).All(dataPointer)
	}
	return WitchCollection(collection, query)
}

func FindPart(collection string, condition Any, skip, limit int, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		return c.Find(condition).Skip(skip).Limit(limit).All(dataPointer)
	}
	return WitchCollection(collection, query)
}

func FindPartOrder(collection string, condition Any, skip, limit int, dataPointer Any, order string) error {
	query := func(c *mgo.Collection) error {
		return c.Find(condition).Sort(order).Skip(skip).Limit(limit).All(dataPointer)
	}
	return WitchCollection(collection, query)
}

func FindManyOrder(collection string, condition Any, order string, limit int, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		if limit == 0 {
			return c.Find(condition).Sort(order).All(dataPointer)
		} else {
			return c.Find(condition).Sort(order).Limit(limit).All(dataPointer)
		}
	}
	return WitchCollection(collection, query)
}

func FindOne(collection string, condition Any, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		return c.Find(condition).One(dataPointer)
	}
	return WitchCollection(collection, query)
}

func Delete(collection, id string) error {
	remove := func(c *mgo.Collection) error {
		return c.RemoveId(bson.ObjectIdHex(id))
	}
	return WitchCollection(collection, remove)
}

func Add(collection string, dataPointer Any) error {
	query := func(c *mgo.Collection) error {
		return c.Insert(dataPointer)
	}
	return WitchCollection(collection, query)
}

func UpdateById(collection string, id string, data Any) error {
	update := func(c *mgo.Collection) error {
		return c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, data)
	}
	return WitchCollection(collection, update)
}

func UpdateByCond(collection string, cond Any, data Any) error {
	update := func(c *mgo.Collection) error {
		_, err := c.UpdateAll(cond, data)
		return err
	}
	return WitchCollection(collection, update)
}

func UpdateOneByCond(collection string, cond Any, data Any) error {
	update := func(c *mgo.Collection) error {
		return c.Update(cond, data)
	}
	return WitchCollection(collection, update)
}

func Upsert(collection string, data Any) error {
	update := func(c *mgo.Collection) error {
		_, err := c.Upsert(nil, data)
		return err
	}
	return WitchCollection(collection, update)
}

func GetCount(collection string, condition Any) int {
	result := 0
	query := func(c *mgo.Collection) error {
		var err error
		result, err = c.Find(condition).Count()
		return err
	}
	WitchCollection(collection, query)
	return result
}

func GetId() int64 {
	var doc Ids
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"n": 1}},
		ReturnNew: true,
	}
	get := func(c *mgo.Collection) error {
		_, err := c.Find(nil).Apply(change, &doc)
		return err
	}
	WitchCollection("ids", get)
	return doc.N
}
