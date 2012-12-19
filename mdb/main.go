package mdb

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	//"strings"
	"web"
)

var (
	address   = flag.String("http", ":7071", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
	mgoUrl    = flag.String("mgo", "127.0.0.1", "the address of mongo server")
	mgoDB     = flag.String("db", "test", "the db of mongo server")
)

func objectCreate(ctx *web.Context, driver *mdb_server, objectType string) {
	definition := driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	bytes, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Panicln("read data from request failed, " + err.Error())
	}

	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Panicln("unmarshal object from request failed, " + err.Error())
	}

	instance_id, err := driver.Create(definition, result)
	if err != nil {
		log.Panicln("insert object to db, " + err.Error())
	}

	ctx.WriteString(fmt.Sprint(instance_id))
}

func objectFindById(ctx *web.Context, driver *mdb_server, objectType, id string) {

	definition := driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	result, err := driver.FindById(definition, id)
	if err != nil {
		log.Panicln("query result from db, " + err.Error())
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Panicln("convert result to json failed, " + err.Error())
	}

	ctx.Write(bytes)
}

func objectUpdateById(ctx *web.Context, driver *mdb_server, objectType, id string) {
	var result map[string]interface{}
	definition := driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	bytes, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Panicln("read data from request failed, " + err.Error())
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Panicln("unmarshal object from request failed, " + err.Error())
	}

	err = driver.Update(definition, id, result)
	if err != nil {
		log.Panicln("update object to db, " + err.Error())
	}

	ctx.WriteString("ok")
}

func objectDeleteById(ctx *web.Context, driver *mdb_server, objectType, id string) {
	definition := driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	ok, err := driver.RemoveById(definition, id)
	if !ok {
		log.Panicln("insert object to db, " + err.Error())
	}

	ctx.WriteString(err.Error())
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

	sess.SetSafe(&mgo.Safe{W: 1, FSync: true, J: true})

	driver := &mdb_server{driver: &mgo_driver{session: sess.DB(*mgoDB)}}

	svr.Get("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectFindById(ctx, driver, objectType, id) })
	svr.Put("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectUpdateById(ctx, driver, objectType, id) })
	svr.Delete("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectDeleteById(ctx, driver, objectType, id) })
	svr.Post("/mdb/(.*)", func(ctx *web.Context, objectType string) { objectCreate(ctx, driver, objectType) })

	svr.Run()
}
