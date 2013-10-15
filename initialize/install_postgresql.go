package main

import (
	"bufio"
	"code.google.com/p/winsvc/mgr"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

var (

	//"tools/sed.exe" -e "s/^\([ \t]*\)#*\([ \t]*listen_addresses[ \t]*=[ \t]*'\)[^']*\(.*\)/\1\2*\3/" "%data_dir%\postgresql.conf" >"%data_dir%\postgresql.conf.new"
	//"tools/sed.exe" -e "s/^\([ \t]*\)#*\([ \t]*port[ \t]*=[ \t]*\)[0-9]*\(.*\)/\1\235432\3/" "%data_dir%\postgresql.conf.new" >"%data_dir%\postgresql.conf"

	address_match = regexp.MustCompile(`^([ \t]*)#*([ \t]*listen_addresses[ \t]*=[ \t]*')[^']*(.*)`)
	port_match    = regexp.MustCompile(`^([ \t]*)#*([ \t]*port[ \t]*=[ \t]*)[0-9]*(.*)`)
)

func install_postgresql(dir, pwd string) error {
	var cmd *exec.Cmd
	pwfile := filepath.Join(os.TempDir(), fmt.Sprintf("aa-%v", time.Now().Nanosecond()))

	st, e := os.Stat(dir)

	if nil != e {
		if !os.IsNotExist(e) {
			return e
		} else {
			os.MkdirAll(dir, 0)
			fmt.Println("directory '" + dir + "' is created.")
		}
	} else if !st.IsDir() {
		return errors.New("'" + dir + "' is not a directory.")
	}

	st, e = os.Stat(filepath.Join(dir, "postgresql.conf"))
	if nil == e {
		if st.IsDir() {
			return errors.New("'" + filepath.Join(dir, "postgresql.conf") + "' is not a directory.")
		}

		fmt.Println("data directory is already initialize...")
		goto install_nt
	} else if !os.IsNotExist(e) {
		return e
	} else {
		if is_empty, _ := dirEmpty(dir); !is_empty {
			os.RemoveAll(dir)
			fmt.Println("directory '" + dir + "' is removed.")
			os.MkdirAll(dir, 0)
			fmt.Println("directory '" + dir + "' is created.")
		}
	}

	cmd = exec.Command("icacls", dir, "/T", "/grant:r", os.Getenv("USERNAME")+":(OI)(CI)F")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	e = cmd.Run()
	if nil != e {
		return errors.New("set write and read perm for current user failed - " + e.Error())
	}

	cmd = exec.Command("icacls", dir, "/T", "/grant:r", "NT AUTHORITY\\NetworkService:(OI)(CI)F")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	e = cmd.Run()
	if nil != e {
		return errors.New("set write and read perm for \"NT AUTHORITY\\NetworkService\" failed - " + e.Error())
	}

	saveFile(pwfile, pwd)
	defer os.Remove(pwfile)

	cmd = exec.Command("runtime_env\\postgresql\\bin\\initdb.exe",
		"--encoding=UTF-8",
		"--auth=md5",
		"--username=postgres",
		"--pwfile="+pwfile,
		dir)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	e = cmd.Run()
	if nil != e {
		os.Remove(pwfile)
		return errors.New("init data directory failed - " + e.Error())
	}
	os.Remove(pwfile)
	e = replaceFile(filepath.Join(dir, "postgresql.conf"), "*", fmt.Sprint(*postgresql_port))
	if nil != e {
		os.Remove(filepath.Join(dir, "postgresql.conf"))
		return errors.New("change listen address and port failed - " + e.Error())
	}
	os.Remove(filepath.Join(dir, "postgresql.conf.bak"))
	fmt.Println("data directory is initialized...")

install_nt:
	lm, e := mgr.Connect()
	if nil != e {
		return errors.New("connect to scm failed" + e.Error())
	}

	svc, e := lm.OpenService("tpt_postgresql")
	if nil == e {
		svc.Close()
		fmt.Println("'tpt_postgresql' service is already installed.")
		return nil
	}

	// if errno, ok := e.(syscall.Errno); !ok || 1060 != errno {
	// 	return fmt.Errorf("connect to scm failed - %v", e)
	// }

	cmd = exec.Command("runtime_env\\postgresql\\bin\\pg_ctl.exe",
		"register",
		"-N", "tpt_postgresql",
		"-U", "NT AUTHORITY\\NetworkService",
		"-D", dir,
		"-S", "auto",
		"-w")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	e = cmd.Run()
	if nil != e {
		return errors.New("create 'tpt_postgresql' service failed - " + e.Error())
	}
	fmt.Println("'tpt_postgresql' service is installed.")
	return nil
}

func replaceFile(nm, host, port string) error {
	e := os.Rename(nm, nm+".bak")
	if nil != e {
		return e
	}

	in, e := os.Open(nm + ".bak")
	if nil != e {
		return e
	}
	defer in.Close()

	out, e := os.Create(nm)
	if nil != e {
		return e
	}
	defer out.Close()

	scaner := bufio.NewScanner(in)
	scaner.Split(bufio.ScanLines)

	for scaner.Scan() {
		s := scaner.Text()
		_, e = out.WriteString(replaceAddressAndPort(s, host, port))
		if nil != e {
			return e
		}
		out.WriteString("\r\n")
	}
	return nil
}

func replaceAddressAndPort(line, host, port string) string {
	if !port_match.MatchString(line) {
		if !address_match.MatchString(line) {
			return line
		}

		ss := address_match.FindStringSubmatch(line)
		return ss[1] + ss[2] + "*" + ss[3]
	}

	ss := port_match.FindStringSubmatch(line)
	return ss[1] + ss[2] + "80" + ss[3]
}

func dirEmpty(nm string) (bool, error) {
	f, err := os.Open(nm)
	if err != nil {
		return false, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return false, err
	}
	if nil != names && 0 != len(names) {
		return false, nil
	}
	return true, nil
}

func saveFile(pwfile, pwd string) {
	out, _ := os.Create(pwfile)
	defer out.Close()
	out.WriteString(pwd)
}
