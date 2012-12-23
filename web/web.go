package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"reflect"
	"regexp"
	"runtime"
	rpprof "runtime/pprof"
	"strconv"
	"strings"
	"time"
)

type Context struct {
	Request     *http.Request
	Params      map[string]string
	QueryParams map[string]string
	Server      *Server
	Writer      http.ResponseWriter
}

func (ctx *Context) WriteString(content string) {
	ctx.Writer.Write([]byte(content))
}

func (ctx *Context) Write(content []byte) (int, error) {
	return ctx.Writer.Write(content)
}

func (ctx *Context) Abort(status int, body string) {
	ctx.Writer.WriteHeader(status)
	ctx.Writer.Write([]byte(body))
}

func (ctx *Context) Redirect(status int, url_ string) {
	ctx.Writer.Header().Set("Location", url_)
	ctx.Writer.WriteHeader(status)
	ctx.Writer.Write([]byte("Redirecting to: " + url_))
}

func (ctx *Context) NotModified() {
	ctx.Writer.WriteHeader(304)
}

func (ctx *Context) NotFound(message string) {
	ctx.Writer.WriteHeader(404)
	ctx.Writer.Write([]byte(message))
}

//Sets the content type by extension, as defined in the mime package. 
//For example, ctx.ContentType("json") sets the content-type to "application/json"
func (ctx *Context) ContentType(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		ctx.Writer.Header().Set("Content-Type", ctype)
	}
}

func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
	if unique {
		ctx.Writer.Header().Set(hdr, val)
	} else {
		ctx.Writer.Header().Add(hdr, val)
	}
}

//Sets a cookie -- duration is the amount of time in seconds. 0 = forever
func (ctx *Context) SetCookie(name string, value string, age int64) {
	var utctime time.Time
	if age == 0 {
		// 2^31 - 1 seconds (roughly 2038)
		utctime = time.Unix(2147483647, 0)
	} else {
		utctime = time.Unix(time.Now().Unix()+age, 0)
	}
	cookie := fmt.Sprintf("%s=%s; expires=%s", name, value, webTime(utctime))
	ctx.SetHeader("Set-Cookie", cookie, false)
}

func getCookieSig(key string, val []byte, timestamp string) string {
	hm := hmac.New(sha1.New, []byte(key))

	hm.Write(val)
	hm.Write([]byte(timestamp))

	hex := fmt.Sprintf("%02x", hm.Sum(nil))
	return hex
}

func (ctx *Context) SetSecureCookie(name string, val string, age int64) {
	//base64 encode the val
	if len(ctx.Server.Config.CookieSecret) == 0 {
		ctx.Server.Logger.Println("Secret Key for secure cookies has not been set. Please assign a cookie secret to web.Config.CookieSecret.")
		return
	}
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(val))
	encoder.Close()
	vs := buf.String()
	vb := buf.Bytes()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := getCookieSig(ctx.Server.Config.CookieSecret, vb, timestamp)
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	ctx.SetCookie(name, cookie, age)
}

func (ctx *Context) GetSecureCookie(name string) (string, bool) {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name != name {
			continue
		}

		parts := strings.SplitN(cookie.Value, "|", 3)

		val := parts[0]
		timestamp := parts[1]
		sig := parts[2]

		if getCookieSig(ctx.Server.Config.CookieSecret, []byte(val), timestamp) != sig {
			return "", false
		}

		ts, _ := strconv.ParseInt(timestamp, 0, 64)

		if time.Now().Unix()-31*86400 > ts {
			return "", false
		}

		buf := bytes.NewBufferString(val)
		encoder := base64.NewDecoder(base64.StdEncoding, buf)

		res, _ := ioutil.ReadAll(encoder)
		return string(res), true
	}
	return "", false
}

func (c *Context) Close() error {
	rwc, buf, _ := c.Writer.(http.Hijacker).Hijack()
	if buf != nil {
		buf.Flush()
	}

	if rwc != nil {
		return rwc.Close()
	}
	return nil
}

// small optimization: cache the context type instead of repeteadly calling reflect.Typeof
var contextType reflect.Type
var exeFile string

// default
func defaultStaticDir() string {
	root, _ := path.Split(exeFile)
	return path.Join(root, "static")
}

func init() {
	contextType = reflect.TypeOf(Context{})
	//find the location of the exe file
	arg0 := path.Clean(os.Args[0])
	wd, _ := os.Getwd()
	if path.IsAbs(arg0) {
		exeFile = arg0
	} else {
		//TODO for robustness, search each directory in $PATH
		exeFile = path.Join(wd, arg0)
	}
}

type route_t struct {
	regexString     string
	compiledRegex   *regexp.Regexp
	method          string
	requiresContext bool
	handler         reflect.Value
}

type ServerConfig struct {
	Name            string
	StaticDirectory string
	Address         string
	CookieSecret    string
	RecoverPanic    bool
}

type Server struct {
	Config *ServerConfig
	routes [HTTP_METHOD_END][]route_t
	Logger *log.Logger
	Env    map[string]interface{}
	//save the listener so it can be closed
	l net.Listener
}

//should the context be passed to the handler?
func requiresContext(handlerType reflect.Type) bool {
	//if the method doesn't take arguments, no
	if handlerType.NumIn() == 0 {
		return false
	}

	//if the first argument is not a pointer, no
	a0 := handlerType.In(0)
	if a0.Kind() != reflect.Ptr {
		return false
	}
	//if the first argument is a context, yes
	if a0.Elem() == contextType {
		return true
	}

	return false
}

func (s *Server) AddRoute(regexString string, method string, handler interface{}) {
	compiledRegex, err := regexp.Compile(regexString)
	if err != nil {
		s.Logger.Printf("Error in route regex %q\n", regexString)
		return
	}
	fv, ok := handler.(reflect.Value)
	if !ok {
		fv = reflect.ValueOf(handler)
	}

	m := HashMethod(method)
	if -1 == m {
		m = HTTP_OTHER_METHOD
	}

	s.routes[m] = append(s.routes[m], route_t{regexString, compiledRegex, method, requiresContext(fv.Type()), fv})
}

//Adds a handler for the 'GET' http method.
func (s *Server) Get(route string, handler interface{}) {
	s.AddRoute(route, "GET", handler)
}

//Adds a handler for the 'POST' http method.
func (s *Server) Post(route string, handler interface{}) {
	s.AddRoute(route, "POST", handler)
}

//Adds a handler for the 'PUT' http method.
func (s *Server) Put(route string, handler interface{}) {
	s.AddRoute(route, "PUT", handler)
}

//Adds a handler for the 'DELETE' http method.
func (s *Server) Delete(route string, handler interface{}) {
	s.AddRoute(route, "DELETE", handler)
}

func (s *Server) ServeHTTP(c http.ResponseWriter, req *http.Request) {
	s.routeHandler(c, req)
}

//Calls a function with recover block
func (s *Server) safelyCall(function reflect.Value, args []reflect.Value) (resp []reflect.Value, e interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if !s.Config.RecoverPanic {

				// go back to panic
				panic(err)
			} else {
				e = err
				resp = nil
				var buffer bytes.Buffer
				buffer.WriteString(fmt.Sprintf("Handler crashed with error - %s\r\n", err))
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
				}
				msg := buffer.String()
				resp = make([]reflect.Value, 1)
				resp[0] = reflect.ValueOf(msg)
				s.Logger.Println(msg)
			}
		}
	}()

	return function.Call(args), nil
}

func (s *Server) routeHandler(w http.ResponseWriter, req *http.Request) {
	requestPath := req.URL.Path
	ctx := Context{req, map[string]string{}, map[string]string{}, s, w}
	//log the request
	var logEntry bytes.Buffer
	fmt.Fprintf(&logEntry, "\033[32;1m%s %s\033[0m\n", req.Method, requestPath)

	//ignore errors from ParseForm because it's usually harmless.
	req.ParseForm()
	if len(req.Form) > 0 {
		for k, v := range req.Form {
			ctx.Params[k] = v[0]
		}
		fmt.Fprintf(&logEntry, "\033[37;1mForms: %v\033[0m\n", req.Form)
	}

	if len(req.URL.Query()) > 0 {
		for k, v := range req.URL.Query() {
			ctx.Params[k] = v[0]
			ctx.QueryParams[k] = v[0]
		}
		fmt.Fprintf(&logEntry, "\033[37;1mQuerys: %v\033[0m\n", req.URL.Query())
	}

	ctx.Server.Logger.Print(logEntry.String())

	//set some default headers
	ctx.SetHeader("Server", "web.go", true)
	tm := time.Now().UTC()
	ctx.SetHeader("Date", webTime(tm), true)

	staticFile := path.Join(s.Config.StaticDirectory, requestPath)
	if fileExists(staticFile) && (req.Method == "GET" || req.Method == "HEAD") {
		http.ServeFile(ctx.Writer, req, staticFile)
		return
	}

	//Set the default content-type
	ctx.SetHeader("Content-Type", "text/html; charset=utf-8", true)

	method := HashMethod(req.Method)
	routes := s.routes[method]
	for i := 0; i < len(routes); i++ {
		route := routes[i]
		cr := route.compiledRegex

		if !cr.MatchString(requestPath) {
			continue
		}
		match := cr.FindStringSubmatch(requestPath)

		if len(match[0]) != len(requestPath) {
			continue
		}

		var args []reflect.Value
		if route.requiresContext {
			args = append(args, reflect.ValueOf(&ctx))
		}
		for _, arg := range match[1:] {
			args = append(args, reflect.ValueOf(arg))
		}

		ret, err := s.safelyCall(route.handler, args)
		if err != nil {
			//there was an error or panic while calling the handler
			ctx.Abort(500, "Server Error")
		}
		if len(ret) == 0 {
			return
		}

		sval := ret[0]

		var content []byte

		if sval.Kind() == reflect.String {
			content = []byte(sval.String())
		} else if sval.Kind() == reflect.Slice && sval.Type().Elem().Kind() == reflect.Uint8 {
			content = sval.Interface().([]byte)
		}
		ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
		ctx.Write(content)
		return
	}

	//try to serve index.html || index.htm
	if indexPath := path.Join(path.Join(s.Config.StaticDirectory, requestPath), "index.html"); fileExists(indexPath) {
		http.ServeFile(ctx.Writer, ctx.Request, indexPath)
		return
	}

	if indexPath := path.Join(path.Join(s.Config.StaticDirectory, requestPath), "index.htm"); fileExists(indexPath) {
		http.ServeFile(ctx.Writer, ctx.Request, indexPath)
		return
	}

	ctx.Abort(404, "Page not found")
}

func NewServer() *Server {
	return &Server{
		Config: &ServerConfig{RecoverPanic: true},
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		Env:    map[string]interface{}{},
	}
}

func (s *Server) initServer() {

	if s.Logger == nil {
		s.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	}

	//try to serve a static file
	if "" == s.Config.StaticDirectory {
		s.Config.StaticDirectory = defaultStaticDir()
	}
}

//Runs the web application and serves http requests
func (s *Server) Run() {
	s.initServer()

	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	for _, pf := range rpprof.Profiles() {
		mux.Handle("/debug/pprof/"+pf.Name(), pprof.Handler(pf.Name()))
	}
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/", s)

	s.Logger.Printf("%s serving %s\n", s.Config.Name, s.Config.Address)

	l, err := net.Listen("tcp", s.Config.Address)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	s.l = l
	err = http.Serve(s.l, mux)
	s.l.Close()
}

//Stops the web server
func (s *Server) Close() {
	if s.l != nil {
		s.l.Close()
	}
}

func (s *Server) SetLogger(logger *log.Logger) {
	s.Logger = logger
}
