package commons

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

type LogCallback func(bytes []byte)

type Writer interface {
	IsEnabled() bool
	Output(calldepth int, s string)
	Printf(format string, v ...interface{})
	Print(v ...interface{})
}

type Logger struct {
	DEBUG Writer
	INFO  Writer
	ERROR Writer
	WARN  Writer
	FATAL PanicWriter

	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix to write at beginning of each line
	flag   int        // properties
	out    io.Writer  // destination for output
	log    LogCallback
	buf    []byte // for accumulating text to write
}

func (self *Logger) InitLoggerWithCallback(cb LogCallback, prefix string, flag int) {
	self.initLogger(nil, cb, prefix, flag)
}

func (self *Logger) InitLoggerWithWriter(wr io.Writer, prefix string, flag int) {
	self.initLogger(wr, nil, prefix, flag)
}

func (self *Logger) initLogger(wr io.Writer, cb LogCallback, prefix string, flag int) {
	self.DEBUG = &nullWriter{super: self, level: "DEBUG"}
	self.INFO = &internalWriter{super: self, level: "INFO"}
	self.WARN = &internalWriter{super: self, level: "WARN"}
	self.ERROR = &internalWriter{super: self, level: "ERROR"}
	self.FATAL.super = self
	self.FATAL.level = "FATAL"

	self.prefix = prefix
	self.flag = flag
	self.log = cb
	self.out = wr
}

func (l *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int, level string) {
	*buf = append(*buf, l.prefix...)
	*buf = append(*buf, ' ')
	*buf = append(*buf, level...)
	*buf = append(*buf, ' ')
	if l.flag&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		if l.flag&log.Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(log.Ltime|log.Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&log.Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(log.Lshortfile|log.Llongfile) != 0 {
		if l.flag&log.Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

// Output writes the output for a logging event.  The string s contains
// the text to print after the prefix specified by the flags of the
// Writer.  A newline is appended if the last character of s is not
// already a newline.  Calldepth is used to recover the PC and is
// provided for generality, although at the moment on all pre-defined
// paths it will be 2.
func (l *Logger) Output(calldepth int, level, s string) {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(log.Lshortfile|log.Llongfile) != 0 {
		// release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line, level)
	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	if nil != l.out {
		l.out.Write(l.buf)
	} else if nil != l.log {
		l.log(l.buf)
	}
}

func (self *Logger) LogInitialized() bool {
	return self.DEBUG != nil
}

func (self *Logger) LogFlags() int {
	return self.flag
}

func (self *Logger) LogPrefix() string {
	return self.prefix
}

func (self *Logger) SetLogFlags(flag int) {
	self.flag = flag
}

func (self *Logger) SetLogPrefix(prefix string) {
	self.prefix = prefix
}

func (self *Logger) switchWriter(wr Writer, new_wr Writer) {
	if self.DEBUG == wr {
		self.DEBUG = new_wr
	}
	if self.INFO == wr {
		self.INFO = new_wr
	}
	if self.ERROR == wr {
		self.ERROR = new_wr
	}
	if self.WARN == wr {
		self.WARN = new_wr
	}
}

type internalWriter struct {
	super *Logger
	level string
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
// Knows the buffer has capacity.
func itoa(buf *[]byte, i int, wid int) {
	var u uint = uint(i)
	if u == 0 && wid <= 1 {
		*buf = append(*buf, '0')
		return
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	*buf = append(*buf, b[bp:]...)
}

func (l *internalWriter) IsEnabled() bool { return true }

func (l *internalWriter) Switch() {
	l.super.switchWriter(l, &nullWriter{super: l.super, level: l.level})
}

func (l *internalWriter) Output(calldepth int, s string) {
	l.super.Output(calldepth+1, l.level, s)
}

func (l *internalWriter) Printf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
}

func (l *internalWriter) Print(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

type nullWriter struct {
	super *Logger
	level string
}

func (l *nullWriter) IsEnabled() bool { return false }

func (l *nullWriter) Switch() {
	l.super.switchWriter(l, &internalWriter{super: l.super, level: l.level})
}

func (l *nullWriter) Output(calldepth int, s string) {}

func (l *nullWriter) Printf(format string, v ...interface{}) {}

func (l *nullWriter) Print(v ...interface{}) {}

type PanicWriter struct {
	super *Logger
	level string
}

func (l *PanicWriter) IsEnabled() bool { return true }

func (l *PanicWriter) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.super.Output(2, l.level, s)
	panic(s)
}

func (l *PanicWriter) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.super.Output(2, l.level, s)
	panic(s)
}

var (
	Log = &Logger{}
)

func init() {
	Log.InitLoggerWithWriter(os.Stdout, "", log.Ltime)
}
