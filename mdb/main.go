package mdb

import (
	"encoding/json"
	"errors"
	"flag"
	//"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"strings"
	"web"
)

var (
	address   = flag.String("http", ":7071", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
	mgoUrl    = flag.String("mgo", "127.0.0.1", "the address of mongo server")
	mgoDB     = flag.String("db", "test", "the db of mongo server")
)

type MdbServer struct {
	driver      Driver
	definitions ClassDefinitions
}

func splitUrl(driver *MdbServer, t, s string) ([]ObjectId, error) {
	ss := strings.Split(s, "/")
	if len(ss)%2 != 1 {
		return nil, errors.New("url format is error, it must is 'type/id(type/id)*'")
	}

	parents := make([]ObjectId, 0, 4)
	parents = append(parents, ObjectId{definition: driver.definitions.Find(ss[1]), id: ss[0]})
	for i := 2; i < len(ss); i += 2 {
		parents = append(parents, ObjectId{definition: driver.definitions.Find(ss[1]), id: ss[i]})
	}
	return parents, nil
}

func objectFindById(driver *MdbServer, ctx *web.Context, objectType, url string) {
	var result map[string]interface{}
	parents, err := splitUrl(driver, objectType, url)
	if nil != err {
		log.Panicln(err.Error())
	}
	definition := parents[len(parents)-1].definition
	id := parents[len(parents)-1].id

	// definition := driver.definitions.Find(t)
	// if nil == definition {
	// 	log.Panicln("class '" + t + "' is not found")
	// }

	driver.driver.FindById(definition, id, parents[0:len(parents)-1])
	if err != nil {
		log.Panicln("query result from db, " + err.Error())
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Panicln("convert result to json failed, " + err.Error())
	}

	ctx.Write(bytes)
}

func objectUpdateById(driver *MdbServer, ctx *web.Context, objectType, url string) {
	// var result map[string]interface{}
	// definition := driver.definitions.Find(objectType)
	// if nil == definition {
	// 	log.Panicln("class '" + objectType + "' is not found")
	// }

	// bytes, err := ioutil.ReadAll(ctx.Request.Body)
	// if err != nil {
	// 	log.Panicln("read data from request failed, " + err.Error())
	// }

	// err = json.Unmarshal(bytes, &result)
	// if err != nil {
	// 	log.Panicln("unmarshal object from request failed, " + err.Error())
	// }

	// err = definition.ValidatePartials(result)
	// if nil != err {
	// 	log.Panicln("validate input data failed, " + err.Error())
	// }

	// err = driver.session.DB(*mgoDB).C(definition.CollectionName()).UpdateId(id, &result)
	// if err != nil {
	// 	log.Panicln("update object to db, " + err.Error())
	// }

	// ctx.WriteString("ok")
}

func objectDeleteById(driver *MdbServer, ctx *web.Context, objectType, id string) {
	// definition := driver.definitions.Find(objectType)
	// if nil == definition {
	// 	log.Panicln("class '" + objectType + "' is not found")
	// }

	// err := driver.session.DB(*mgoDB).C(definition.CollectionName()).RemoveId(id)
	// if err != nil {
	// 	log.Panicln("insert object to db, " + err.Error())
	// }

	// ctx.WriteString("ok")
}

func objectCreate(driver *MdbServer, ctx *web.Context, objectType string) {
	// var result map[string]interface{}
	// definition := driver.definitions.Find(objectType)
	// if nil == definition {
	// 	log.Panicln("class '" + objectType + "' is not found")
	// }

	// bytes, err := ioutil.ReadAll(ctx.Request.Body)
	// if err != nil {
	// 	log.Panicln("read data from request failed, " + err.Error())
	// }

	// err = json.Unmarshal(bytes, &result)
	// if err != nil {
	// 	log.Panicln("unmarshal object from request failed, " + err.Error())
	// }

	// instance, err := definition.CreateIt(result)
	// if nil != err {
	// 	log.Panicln("validate input data failed, " + err.Error())
	// }

	// err = driver.session.DB(*mgoDB).C(definition.CollectionName()).Insert(&result)
	// if err != nil {
	// 	log.Panicln("insert object to db, " + err.Error())
	// }

	// ctx.WriteString("ok")
}

func mdb_get(driver *MdbServer, ctx *web.Context, url string) {
}

func mdb_post(driver *MdbServer, ctx *web.Context, url string) {

}

func main() {
	flag.Parse()
	svr := web.NewServer()
	svr.Config.Name = "meijing-mdb v1.0"
	svr.Config.Address = *address
	svr.Config.StaticDirectory = *directory
	svr.Config.CookieSecret = *cookies
	sess, err := mgo.Dial(*mgoUrl)
	if nil != err {
		log.Printf("connect to mongo server failed, %s", err.Error())
		return
	}

	driver := &MdbServer{driver: &mgo_driver{session: sess.DB(*mgoDB)}}

	svr.Get("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectFindById(driver, ctx, objectType, id) })
	svr.Put("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectUpdateById(driver, ctx, objectType, id) })
	svr.Delete("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectDeleteById(driver, ctx, objectType, id) })
	svr.Post("/mdb/(.*)", func(ctx *web.Context, objectType string) { objectCreate(driver, ctx, objectType) })

	svr.Run()
}
