package commons

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"sync"
	"time"
)

type LogCallback func(s string) error

type Logger interface {
	IsEnabled() bool
	Output(calldepth int, s string) error
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Flags() int
	SetFlags(flag int)
	Prefix() string
	SetPrefix(prefix string)
}

type Loggers struct {
	DEBUG Logger
	INFO  Logger
	ERROR Logger
	WARN  Logger
	FATAL Logger

	logFlags  int
	logPrefix string
}

func (self *Loggers) InitLoggers(out io.Writer, cb LogCallback, prefix string, flag int) error {
	self.logFlags = flag
	self.logPrefix = prefix
	self.DEBUG = &InternalLogger{out: out, log: cb, prefix: prefix + " - DEBUG ", flag: flag}
	self.INFO = &InternalLogger{out: out, log: cb, prefix: prefix + " - INFO ", flag: flag}
	self.WARN = &InternalLogger{out: out, log: cb, prefix: prefix + " - WARN ", flag: flag}
	self.ERROR = &InternalLogger{out: out, log: cb, prefix: prefix + " - ERROR ", flag: flag}
	self.FATAL = &InternalLogger{out: out, log: cb, prefix: prefix + " - FATAL ", flag: flag, is_panic: true}
	return nil
}

func (self *Loggers) LogInitialized() bool {
	return self.DEBUG != nil
}

func (self *Loggers) LogFlags() int {
	return self.logFlags
}

func (self *Loggers) LogPrefix() string {
	return self.logPrefix
}

func (self *Loggers) SetLogFlags(flag int) {
	self.logFlags = flag
	self.DEBUG.SetFlags(flag)
	self.INFO.SetFlags(flag)
	self.WARN.SetFlags(flag)
	self.ERROR.SetFlags(flag)
	self.FATAL.SetFlags(flag)
}

func (self *Loggers) SetLogPrefix(prefix string) {
	self.logPrefix = prefix
	self.DEBUG.SetPrefix(prefix + " - DEBUG ")
	self.INFO.SetPrefix(prefix + " - INFO ")
	self.WARN.SetPrefix(prefix + " - WARN ")
	self.ERROR.SetPrefix(prefix + " - ERROR ")
	self.FATAL.SetPrefix(prefix + " - FATAL ")
}

type InternalLogger struct {
	mu       sync.Mutex // ensures atomic writes; protects the following fields
	prefix   string     // prefix to write at beginning of each line
	flag     int        // properties
	out      io.Writer  // destination for output
	log      LogCallback
	buf      []byte // for accumulating text to write
	is_panic bool
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

func (l *InternalLogger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
	*buf = append(*buf, l.prefix...)
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
// Logger.  A newline is appended if the last character of s is not
// already a newline.  Calldepth is used to recover the PC and is
// provided for generality, although at the moment on all pre-defined
// paths it will be 2.
func (l *InternalLogger) Output(calldepth int, s string) error {
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
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	if nil != l.out {
		_, err := l.out.Write(l.buf)
		if l.is_panic {
			panic(string(l.buf))
		}
		return err
	} else if nil != l.log {
		s := string(l.buf)
		err := l.log(s)
		if l.is_panic {
			panic(s)
		}
		return err
	} else if l.is_panic {
		panic(string(l.buf))
	}
	return nil
}

func (l *InternalLogger) IsEnabled() bool { return true }

func (l *InternalLogger) Printf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
}

func (l *InternalLogger) Print(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

func (l *InternalLogger) Flags() int { return l.flag }

func (l *InternalLogger) SetFlags(flag int) { l.flag = flag }

func (l *InternalLogger) Prefix() string { return l.prefix }

func (l *InternalLogger) SetPrefix(prefix string) { l.prefix = prefix }

type NullLogger struct{}

func (l *NullLogger) IsEnabled() bool { return false }

func (l *NullLogger) Output(calldepth int, s string) error { return nil }

func (l *NullLogger) Printf(format string, v ...interface{}) {}

func (l *NullLogger) Print(v ...interface{}) {}

func (l *NullLogger) Flags() int { return 0 }

func (l *NullLogger) SetFlags(flag int) {}

func (l *NullLogger) Prefix() string { return "null - logger" }

func (l *NullLogger) SetPrefix(prefix string) {}
