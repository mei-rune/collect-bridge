package commons

import (
	"io"
	"log"
)

type Loggers struct {
	DEBUG Logger
	INFO  Logger
	ERROR Logger
	WARN  Logger
	FATAL Logger

	logFlags  int
	logPrefix string
}

func (self *Loggers) InitLoggers(out io.Writer, prefix string, flag int) error {
	self.logFlags = flag
	self.logPrefix = prefix
	self.DEBUG = &InternalLogger{internal: log.New(out, prefix+" - DEBUG -", flag)}
	self.INFO = &InternalLogger{internal: log.New(out, prefix+" - INFO -", flag)}
	self.WARN = &InternalLogger{internal: log.New(out, prefix+" - WARN -", flag)}
	self.ERROR = &InternalLogger{internal: log.New(out, prefix+" - ERROR -", flag)}
	self.FATAL = &InternalLogger{internal: log.New(out, prefix+" - FATAL -", flag)}
	return nil
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
	self.DEBUG.SetPrefix(prefix + " - DEBUG -")
	self.INFO.SetPrefix(prefix + " - INFO -")
	self.WARN.SetPrefix(prefix + " - WARN -")
	self.ERROR.SetPrefix(prefix + " - ERROR -")
	self.FATAL.SetPrefix(prefix + " - FATAL -")
}

type Logger interface {
	Output(calldepth int, s string) error
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Flags() int
	SetFlags(flag int)
	Prefix() string
	SetPrefix(prefix string)
}

type InternalLogger struct {
	internal *log.Logger
}

func (l *InternalLogger) Output(calldepth int, s string) error {
	return l.internal.Output(calldepth+1, s)
}

func (l *InternalLogger) Printf(format string, v ...interface{}) { l.internal.Printf(format, v) }

func (l *InternalLogger) Print(v ...interface{}) { l.internal.Print(v) }

func (l *InternalLogger) Println(v ...interface{}) { l.internal.Println(v) }

func (l *InternalLogger) Fatal(v ...interface{}) { l.internal.Fatal(v) }

func (l *InternalLogger) Fatalf(format string, v ...interface{}) { l.internal.Fatalf(format, v) }

func (l *InternalLogger) Fatalln(v ...interface{}) { l.internal.Fatalln(v) }

func (l *InternalLogger) Panic(v ...interface{}) { l.internal.Panic(v) }

func (l *InternalLogger) Panicf(format string, v ...interface{}) { l.internal.Panicf(format, v) }

func (l *InternalLogger) Panicln(v ...interface{}) { l.internal.Panicln(v) }

func (l *InternalLogger) Flags() int { return l.internal.Flags() }

func (l *InternalLogger) SetFlags(flag int) { l.internal.SetFlags(flag) }

func (l *InternalLogger) Prefix() string { return l.internal.Prefix() }

func (l *InternalLogger) SetPrefix(prefix string) { l.internal.SetPrefix(prefix) }

type NullLogger struct{}

func (l *NullLogger) Output(calldepth int, s string) error { return nil }

func (l *NullLogger) Printf(format string, v ...interface{}) {}

func (l *NullLogger) Print(v ...interface{}) {}

func (l *NullLogger) Println(v ...interface{}) {}

func (l *NullLogger) Fatal(v ...interface{}) {}

func (l *NullLogger) Fatalf(format string, v ...interface{}) {}

func (l *NullLogger) Fatalln(v ...interface{}) {}

func (l *NullLogger) Panic(v ...interface{}) {}

func (l *NullLogger) Panicf(format string, v ...interface{}) {}

func (l *NullLogger) Panicln(v ...interface{}) {}

func (l *NullLogger) Flags() int { return 0 }

func (l *NullLogger) SetFlags(flag int) {}

func (l *NullLogger) Prefix() string { return "null - logger" }

func (l *NullLogger) SetPrefix(prefix string) {}
