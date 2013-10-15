package main

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
)

const (
	NSSM_REGISTRY              = "SYSTEM\\CurrentControlSet\\Services\\%s\\Parameters"
	NSSM_REG_EXE               = "Application"
	NSSM_REG_FLAGS             = "AppParameters"
	NSSM_REG_DIR               = "AppDirectory"
	NSSM_REG_ENV               = "AppEnvironment"
	NSSM_REG_EXIT              = "AppExit"
	NSSM_REG_THROTTLE          = "AppThrottle"
	NSSM_REG_STDIN             = "AppStdin"
	NSSM_REG_STDOUT            = "AppStdout"
	NSSM_REG_STDERR            = "AppStderr"
	NSSM_REG_STDIO_SHARING     = "ShareMode"
	NSSM_REG_STDIO_DISPOSITION = "CreationDisposition"
	NSSM_REG_STDIO_FLAGS       = "FlagsAndAttributes"
	NSSM_STDIO_LENGTH          = 29
)

func switch_to(name, old_exe, new_exe, params string) error {
	appkey, err := OpenKey(syscall.HKEY_LOCAL_MACHINE, fmt.Sprint(NSSM_REGISTRY, name))
	if nil != err {
		return errors.New("open registry of '" + name + "' failed, " + err.Error())
	}
	if 0 != len(old_exe) {
		s, e := appkey.GetString(NSSM_REG_EXE)
		if nil != e {
			return errors.New("load '" + NSSM_REG_EXE + "' of '" + name + "' failed, " + e.Error())
		}

		if old_exe != s {
			return errors.New("old exec is not equal '" + old_exe + "', actual is '" + s + "'")
		}
	}

	if err = appkey.SetString(NSSM_REG_EXE, new_exe); nil != err {
		return errors.New("set '" + NSSM_REG_EXE + "' of '" + name + "' failed, " + err.Error())
	}

	if err = appkey.SetString(NSSM_REG_FLAGS, params); nil != err {
		return errors.New("set '" + NSSM_REG_FLAGS + "' of '" + name + "' failed, " + err.Error())
	}
	return nil
}

func restart(name string) {
	exec.Command("cmd", "/c \"start sc stop "+name+" && sc start "+name+"\"")
}
