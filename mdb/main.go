package mdb

import (
	"encoding/json"
	"errors"
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

type MdbServer struct {
	driver      Driver
	definitions ClassDefinitions
}

func (self *MdbServer) validate(cls *ClassDefinition, attributes map[string]interface{}) (map[string]interface{}, error) {
	//new_attributes := make(map[string]interface{}, len(attributes))
	return nil, errors.New("not implemented")
}

func (self *MdbServer) Create(cls *ClassDefinition, attributes map[string]interface{}) (interface{}, error) {
	//attributes, errs := self.validate(cls, attributes)
	return nil, errors.New("not implemented")

}
func (self *MdbServer) FindById(cls *ClassDefinition, id interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (self *MdbServer) Update(cls *ClassDefinition, id interface{}, attributes map[string]interface{}) error {
	return errors.New("not implemented")
}

func (self *MdbServer) RemoveById(cls *ClassDefinition, id interface{}) error {
	return errors.New("not implemented")
}

// func splitUrl(driver *MdbServer, t, s string) ([]ObjectId, error) {
// 	ss := strings.Split(s, "/")
// 	if len(ss)%2 != 1 {
// 		return nil, errors.New("url format is error, it must is 'type/id(type/id)*'")
// 	}

// 	parents := make([]ObjectId, 0, 4)
// 	parents = append(parents, ObjectId{definition: driver.definitions.Find(ss[1]), id: ss[0]})
// 	for i := 2; i < len(ss); i += 2 {
// 		parents = append(parents, ObjectId{definition: driver.definitions.Find(ss[1]), id: ss[i]})
// 	}
// 	return parents, nil
// }

func objectCreate(driver *MdbServer, ctx *web.Context, objectType string) {
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

func objectFindById(driver *MdbServer, ctx *web.Context, objectType, id string) {

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

func objectUpdateById(driver *MdbServer, ctx *web.Context, objectType, id string) {
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

func objectDeleteById(driver *MdbServer, ctx *web.Context, objectType, id string) {
	definition := driver.definitions.Find(objectType)
	if nil == definition {
		log.Panicln("class '" + objectType + "' is not found")
	}

	err := driver.RemoveById(definition, id)
	if err != nil {
		log.Panicln("insert object to db, " + err.Error())
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
