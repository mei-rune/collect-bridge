package mdb

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"web"
)

var (
	address   = flag.String("http", ":7071", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
	mgoUrl    = flag.String("mgo", "127.0.0.1", "the address of mongo server")
	mgoDB     = flag.String("db", "test", "the db of mongo server")
)

type MdbDriver struct {
	session     *mgo.Session
	definitions ClassDefinitions
}

func objectFindById(driver MdbDriver, ctx *web.Context, objectType, id string) {
	var result map[string]interface{}
	definition = driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	err := driver.session.DB(mgoDB).C(definition.CollectionName()).FindId(id).One(&result)
	if err != nil {
		log.Panicln("query result from db, " + err)
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Panicln("convert result to json failed, " + err)
	}

	ctx.Write(bytes)
}

func objectUpdateById(driver MdbDriver, ctx *web.Context, objectType, id string) {
	var result map[string]interface{}
	definition = driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	bytes, err := ioutil.ReadAll(ctx)
	if err != nil {
		log.Panicln("read data from request failed, " + err)
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Panicln("unmarshal object from request failed, " + err)
	}

	err = driver.session.DB(mgoDB).C(definition.CollectionName()).UpdateId(id, &result)
	if err != nil {
		log.Panicln("update object to db, " + err)
	}

	ctx.WriteString("ok")
}

func objectDeleteById(driver MdbDriver, ctx *web.Context, objectType, id string) {
	var result map[string]interface{}
	definition = driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	err := driver.session.DB(mgoDB).C(definition.CollectionName()).RemoveId(id)
	if err != nil {
		log.Panicln("insert object to db, " + err)
	}

	ctx.WriteString("ok")
}

func objectCreate(driver MdbDriver, ctx *web.Context, objectType string) {
	var result map[string]interface{}
	definition = driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	bytes, err := ioutil.ReadAll(ctx)
	if err != nil {
		log.Panicln("read data from request failed, " + err)
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Panicln("unmarshal object from request failed, " + err)
	}

	err = driver.session.DB(mgoDB).C(definition.CollectionName()).Insert(&result)
	if err != nil {
		log.Panicln("insert object to db, " + err)
	}

	ctx.WriteString("ok")
}

func main() {
	flag.Parse()
	svr := web.NewServer()
	svr.Config.Name = "meijing-mdb v1.0"
	svr.Config.Address = *address
	svr.Config.StaticDirectory = *directory
	svr.Config.CookieSecret = *cookies

	driver := &MdbDriver{session: mgo.Dial(mgoUrl)}

	svr.Get("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectFindById(driver, ctx, objectType, id) })
	svr.Put("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectUpdateById(driver, ctx, objectType, id) })
	svr.Delete("/mdb/(.*)/(.*)", func(ctx *web.Context, objectType, id string) { objectDeleteById(driver, ctx, objectType, id) })
	svr.Post("/mdb/(.*)", func(ctx *web.Context, objectType string) { objectCreate(driver, ctx, objectType) })

	svr.Run()
}
