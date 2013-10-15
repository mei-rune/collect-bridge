package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
)

var (
	listenAddress = flag.String("listen", ":80", "the address of http")
	run_mode      = flag.String("mode", "all", "init_db, init_postgresql, console, backend, all")

	postgresql_data_dir = flag.String("postgresql.data_dir", "", "the postgresql data dir")
	postgresql_host     = flag.String("postgresql.host", "127.0.0.1", "the postgresql host")
	postgresql_port     = flag.Int("postgresql.port", 35432, "the postgresql port")
	postgresql_password = flag.String("postgresql.password", "", "the postgresql password")

	cd_dir = ""

	retry_list = []*regexp.Regexp{regexp.MustCompile(`^/?[0-9]+/retry/?$`),
		regexp.MustCompile(`^/?initialize/[0-9]+/retry/?$`),
		regexp.MustCompile(`^/?initialize/initialize/[0-9]+/retry/?$`)}

	delete_by_id_list = []*regexp.Regexp{regexp.MustCompile(`^/?[0-9]+/delete/?$`),
		regexp.MustCompile(`^/?initialize/[0-9]+/delete/?$`),
		regexp.MustCompile(`^/?initialize/initialize/[0-9]+/delete/?$`)}

	job_id_list = []*regexp.Regexp{regexp.MustCompile(`^/?[0-9]+/?$`),
		regexp.MustCompile(`^/?initialize/[0-9]+/?$`),
		regexp.MustCompile(`^/?initialize/initialize/[0-9]+/?$`)}
)

func init() {
	var e error
	cd_dir, e = os.Getwd()
	if nil != e {
		panic(e)
	}
}

func main() {
	flag.Parse()
	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}

	if "" == *postgresql_password {
		flag.Set("postgresql.password", os.Getenv("PGPASSWORD"))
	}

	switch *run_mode {
	case "init_postgresql":
		e := init_postgresql(fmt.Sprintf("host=%s port=%d dbname=postgres user=postgres password=%s sslmode=disable",
			*postgresql_host, *postgresql_port, *postgresql_password))
		if nil != e {
			log.Print(e)
			os.Exit(1)
		}
		return
	case "install_postgresql":
		if "" == *postgresql_data_dir {
			log.Print("directory is empty.")
			os.Exit(1)
			return
		}

		e := install_postgresql(*postgresql_data_dir, *postgresql_password, "*", fmt.Sprint(*postgresql_port))
		if nil != e {
			log.Print(e)
			os.Exit(1)
		}
		return
	}

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				switch r.URL.Path {
				case "/", "/index.html", "/index.htm", "/initialize", "/initialize/":
					indexHandler(w, r)
					return
				case "/static/initialize/bootstrap.css":
					bootstrapCssHandler(w, r)
					return
				case "/static/initialize/bootstrap_modal.js":
					bootstrapModalJsHandler(w, r)
					return
				case "/static/initialize/bootstrap_popover.js":
					bootstrapPopoverJsHandler(w, r)
					return
				case "/static/initialize/bootstrap_tab.js":
					bootstrapTabJsHandler(w, r)
					return
				case "/static/initialize/bootstrap_tooltip.js":
					bootstrapTooltipJsHandler(w, r)
					return
				case "/static/initialize/dj_mon.css":
					djmonCssHandler(w, r)
					return
				case "/static/initialize/dj_mon.js":
					djmonJsHandler(w, r)
					return
				case "/static/initialize/jquery.min.js":
					jqueryJsHandler(w, r)
					return
				case "/static/initialize/mustache.js":
					mustascheJsHandler(w, r)
					return
				}
				// case "POST":
				// 	switch r.URL.Path {

				// 	}
			}

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})
	log.Println("[setup] serving at '" + *listenAddress + "'")
	http.ListenAndServe(*listenAddress, nil)
}

func fileExists(nm string) bool {
	fs, e := os.Stat(nm)
	if nil != e {
		return false
	}
	return !fs.IsDir()
}

func fileHandler(w http.ResponseWriter, r *http.Request, path, default_content string) {
	name := cd_dir + path
	if fileExists(name) {
		http.ServeFile(w, r, name)
		return
	}

	io.WriteString(w, default_content)
}

func bootstrapCssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/css; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/bootstrap.css", bootstrap_css)
}
func bootstrapModalJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/javascript; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/bootstrap_modal.js", bootstrap_modal_js)
}
func bootstrapPopoverJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/javascript; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/bootstrap_popover.js", bootstrap_popover_js)
}
func bootstrapTabJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/javascript; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/bootstrap_tab.js", bootstrap_tab_js)
}
func bootstrapTooltipJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/javascript; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/bootstrap_tooltip.js", bootstrap_tooltip_js)
}
func djmonCssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/css; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/dj_mon.css", dj_mon_css)
}
func djmonJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/javascript; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/dj_mon.js", dj_mon_js)
}
func jqueryJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/javascript; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/jquery.min.js", jquery_min_js)
}
func mustascheJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = []string{"text/javascript; charset=utf-8"}
	fileHandler(w, r, "/static/initialize/mustasche.js", mustasche_js)
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fileHandler(w, r, "/index.html", index_html)
}
