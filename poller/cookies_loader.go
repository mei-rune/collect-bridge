package poller

import (
	"commons"
	"errors"
	"strconv"
	"strings"
)

type cookiesLoader interface {
	persistCookiesWithAcitonId(id int64, ctx map[string]interface{}, cookie map[string]interface{}) error
	loadCookiesWithAcitonId(id int64, ctx map[string]interface{}) (map[string]interface{}, error)
}

type cookiesLoaderImpl struct {
	isPersist                bool
	loadFromWebWhileNotFound bool
	client                   *commons.Client
	id2cookies               map[int64]map[string]interface{}
}

func (self *cookiesLoaderImpl) loadCookiesByAcitonId(id int64) (map[string]interface{}, error) {
	res := self.client.Get(map[string]string{"id": "@" + strconv.FormatInt(id, 10)})
	if res.HasError() {
		if 404 == res.ErrorCode() {
			return nil, nil
		}
		return nil, errors.New(res.ErrorMessage())
	}

	return res.Value().AsObject()
}

func (self *cookiesLoaderImpl) loadCookies(query map[string]string) (int, error) {
	res := self.client.Get(query)
	if res.HasError() {
		return 0, errors.New("load cookies failed, " + res.ErrorMessage())
	}

	cookies, e := res.Value().AsObjects()
	if nil != e {
		return 0, errors.New("load cookies failed, results is not a []map[string]interface{}, " + e.Error())
	}

	if nil == cookies || 0 == len(cookies) {
		return 0, nil
	}

	for _, attributes := range cookies {
		action_id := commons.GetInt64WithDefault(attributes, "action_id", 0)
		self.id2cookies[action_id] = attributes
	}
	return len(cookies), nil
}

func (self *cookiesLoaderImpl) init() error {
	self.id2cookies = map[int64]map[string]interface{}{}

	if *not_limit {
		_, e := self.loadCookies(map[string]string{})
		return e
	} else {
		for offset := 0; ; offset += 100 {
			count, e := self.loadCookies(map[string]string{"limit": "100", "offset": strconv.FormatInt(int64(offset), 10)})
			if nil != e {
				return e
			}

			if 100 != count {
				break
			}
		}
	}
	return nil
}

func (self *cookiesLoaderImpl) initWithIds(id_list []string) error {
	self.id2cookies = map[int64]map[string]interface{}{}
	offset := 0
	for ; (offset + 100) < len(id_list); offset += 100 {
		_, e := self.loadCookies(map[string]string{"@id": "[in]" + strings.Join(id_list[offset:offset+100], ",")})
		if nil != e {
			return e
		}
	}

	_, e := self.loadCookies(map[string]string{"@id": "[in]" + strings.Join(id_list[offset:], ",")})
	if nil != e {
		return e
	}

	return nil
}

func (self *cookiesLoaderImpl) loadCookiesWithAcitonId(id int64, ctx map[string]interface{}) (map[string]interface{}, error) {
	if c, ok := self.id2cookies[id]; ok {
		delete(self.id2cookies, id)
		return c, nil
	}
	if self.loadFromWebWhileNotFound {
		return self.loadCookiesByAcitonId(id)
	}
	return nil, nil
}

func (self *cookiesLoaderImpl) persistCookiesWithAcitonId(id int64, ctx map[string]interface{}, cookie map[string]interface{}) error {
	if !self.isPersist {
		return nil
	}

	if nil == self.id2cookies {
		self.id2cookies = map[int64]map[string]interface{}{}
	}
	self.id2cookies[id] = cookie
	return nil
}

type mockCookiesLoader struct {
	e       error
	cookies map[string]interface{}
}

func (self *mockCookiesLoader) loadCookiesWithAcitonId(id int64, ctx map[string]interface{}) (map[string]interface{}, error) {
	return self.cookies, self.e
}

func (self *mockCookiesLoader) persistCookiesWithAcitonId(id int64, ctx map[string]interface{}, cookie map[string]interface{}) error {
	return nil
}
