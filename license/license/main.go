package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"license"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	listen_port        = flag.String("listen", ":37076", "the listen port of http.")
	license_file       = flag.String("license", "tpt.lic", "")
	expired_title      = flag.String("expired_title", " (试用到期了)", "")
	unregistered_title = flag.String("unregistered_title", " (试用)", "")

	expired_time time.Time
)

func init() {
	var e error
	expired_time, e = time.Parse("2006-01-02 15:04:05", "2014-12-30 00:00:01")
	if nil != e {
		panic(e)
	}

	license.LicenseUrl = flag.String("license_srv", "http://127.0.0.1:37076/", "")
}

type License struct {
	origin          map[string]interface{}
	meta            map[string]interface{}
	invalid_nodes   map[string]string
	title           string
	invalid_license bool
	is_expired      bool
	count           uint32
}

func (self *License) init() {
	self.IsExpired()
	if self.is_expired {
		// fmt.Println("license is expired.")
		return
	}

	if self.invalid_license {
		// fmt.Println("license is invalid.")
		return
	}

	hd_list := license.GetAllHD()
	if nil == hd_list || 0 == len(hd_list) {
		self.invalid_license = true
		// fmt.Println("read hd failed.")
		return
	}

	interface_list := license.GetAllInterfaces()
	if nil == interface_list || 0 == len(interface_list) {
		self.invalid_license = true
		// fmt.Println("read interfaces failed.")
		return
	}

	hits := 0
	if old, ok := self.origin["hd"]; ok {
		if old_hd, ok := old.([]interface{}); ok {
			for _, old_v := range old_hd {
				old_one, ok := old_v.([]interface{})
				if !ok {
					continue
				}

				for _, v := range hd_list {
					if v[0] == old_one[0] && v[1] == old_one[1] {
						hits++
					}
				}
			}
		}
	}

	if 0 >= hits {
		self.invalid_license = true
		// fmt.Println("hd is not found.")

		// if old, ok := self.origin["hd"]; ok {
		// 	if old_hd, ok := old.([]interface{}); ok {
		// 		for old_k, old_v := range old_hd {
		// 			// fmt.Println(old_k, old_v)
		// 		}
		// 		for k, v := range hd_list {
		// 			// fmt.Println(k, v[0], v[1])
		// 		}
		// 	} else {
		// 		fmt.Printf("hd is not found - %T.\r\n", old)
		// 	}
		// } else {
		// 	// fmt.Println("hd is not found.")
		// }
		return
	}

	hits = 0
	if old, ok := self.origin["interfaces"]; ok {
		if old_ifs, ok := old.([]interface{}); ok {
			for _, old_v := range old_ifs {
				for _, v := range interface_list {
					if v == old_v {
						hits++
					}
				}
			}
		}
	}

	if 0 >= hits {
		self.invalid_license = true
		// fmt.Println("interfaces is not found.")
		return
	}
}

func (self *License) IsExpired() bool {
	if self.is_expired {
		return true
	}

	expired, ok := self.meta["expired"]
	if !ok || nil == expired {
		self.invalid_license = true
		return true
	}

	if "notlimit" == expired {
		return false
	}

	t, err := time.Parse(time.RFC3339, fmt.Sprint(expired))
	if nil != err {
		self.invalid_license = true
		return true
	}

	if time.Now().After(t) {
		self.is_expired = true
		return true
	}
	return false
}

func (self *License) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint32(&self.count, 1)
	if 9999 == atomic.LoadUint32(&self.count)%10000 {
		time.AfterFunc(1*time.Second, func() { self.init() })
	}

	defer func() {
		if e := recover(); nil != e {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("[panic]%v", e))
			for i := 1; ; i += 1 {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, buffer.String())
		}
	}()

	if "/" == r.URL.Path {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("This is an example server.\n"))
		return
	}

	if nil == (map[string]interface{})(self.meta) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "UNREGISTERED")
		return
	}

	// check the count of nodes
	if strings.HasPrefix(r.URL.Path, "/l5472/") {
		if self.is_expired {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "EXPIRED")
			return
		}

		if self.invalid_license {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "UNREGISTERED")
			return
		}

		var excepted interface{}
		var ok bool

		ss := strings.SplitN(r.URL.Path[len("/l5472/"):], "_", 2)
		if 0 == len(ss[1]) {
			ss[1] = ss[0]
			ss[0] = ""
			excepted, ok = self.meta["node"]
		} else {
			if excepted, ok = self.meta[ss[0]+"_node"]; !ok || nil == excepted {
				excepted, ok = self.meta["node"]
			}
		}
		if !ok {
			self.invalid_license = true
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "UNREGISTERED")
			// fmt.Println("node is not found.")
			return
		}
		excepted_count, err := strconv.ParseInt(fmt.Sprint(excepted), 10, 64)
		if nil != err {
			self.invalid_license = true
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "UNREGISTERED")
			// fmt.Println("node is not a int.")
			return
		}
		if -1 == excepted_count {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "OK")
			return
		}
		count, err := strconv.ParseInt(ss[1], 10, 64)
		if nil != err {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		if 0 == excepted_count || count > excepted_count {
			if nil == self.invalid_nodes {
				self.invalid_nodes = map[string]string{}
			}
			self.invalid_nodes[ss[0]+"node"] = fmt.Sprintf("%d-%d", excepted_count, count)

			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "NONEXCEPTED")
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
		return
	}

	if r.URL.Path == "/o83e56" {
		if self.invalid_license {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "UNREGISTERED")
			return
		}
		if self.is_expired {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "EXPIRED")
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
		return
	}

	// whether the module is enabled.
	if strings.HasPrefix(r.URL.Path, "/o7456/") {
		if self.is_expired {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "EXPIRED")
			return
		}

		if self.invalid_license {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "UNREGISTERED")
			return
		}

		module := r.URL.Path[len("/o7456/"):]
		v, ok := self.meta["module_"+module]
		if !ok {
			if v, ok = self.meta["module_all"]; !ok {
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, "FAILED")
				return
			}
		}

		if "enabled" == v {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "OK")
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "FAILED")
		}
		return
	}

	// get the title
	if r.URL.Path == "/a629433" {
		title := ""
		if v := self.meta["title"]; nil != v {
			title = fmt.Sprint(v)
		}

		if self.is_expired {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, title+*expired_title)
			return
		}

		if self.invalid_license {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, title+*unregistered_title)
			return
		}

		expired, ok := self.meta["expired"]
		if !ok || nil == expired {
			self.invalid_license = true
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, title+*unregistered_title)
			// fmt.Println("expired is not found.")
			return
		}

		if "notlimit" == expired {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, title)
			return
		}

		t, err := time.Parse(time.RFC3339, fmt.Sprint(expired))
		if nil != err {
			self.invalid_license = true
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, title+*unregistered_title)
			// fmt.Println("expired is not a time.")
			return
		}

		if time.Now().After(t) || expired_time.After(t) {
			self.is_expired = true
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, title+*expired_title)
			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, title+*unregistered_title)
		return
	}

	if r.URL.Path == "/hle834t" {
		if self.invalid_license {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "UNREGISTERED")
			return
		}
		if self.is_expired {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "EXPIRED")
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
		return
	}

	if r.URL.Path == "/rth56w3" {
		self.invalid_license = true
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
		return
	}
	http.NotFound(w, r)
}

func main() {
	flag.Parse()
	file, err := license.SearchLicenseFile(*license_file)
	if nil != err {
		log.Fatal(err)
		return
	}
	data, err := license.DecryptoFile(file)
	if err != nil {
		err = http.ListenAndServe(*listen_port, http.Handler(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "UNREGISTERED")
		}))
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	var attributes map[string]interface{}
	auth := data["auth"]
	if nil != auth {
		attributes, _ = auth.(map[string]interface{})
	}

	lic := &License{origin: data, meta: attributes}
	log.Printf("[license] listen at %v", *listen_port)
	func() { lic.init() }()
	err = http.ListenAndServe(*listen_port, lic)
	//err = http.ListenAndServeTLS(*listen_port, "cacert.pem", "key.pem", lic)
	if err != nil {
		log.Fatal(err)
	}
}
