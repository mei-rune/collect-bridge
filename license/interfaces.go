package license

import (
	"net"
)

var (
	public = "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAs9NBYPEj+TpsTlSesfV/P6feyVT6p4dd0XticQiD/TMoVbOpLctoePH7l53ypg5vxQIevGJ++Fi5pHeOeJS4i2h8AeSr9vMvAnFoN2MpmUrUzYKXHCd1Q3IAzb0iYf4tepmM3ka8dZtgc65w/tOnRyWOE/oZJoKClmMShwcdIo/0aIDxaE+eu0C/wO853iCd8+vFLmbRL+QMjQ1e/FENWndXdchkClmjO79iEOgxLhnHRdQ3/Z7Vm9JVkUQ0tdxZ3OOgOIC/kgQUaGKGCNNx5o3ZVo0rYpAJ2xL5hun3Pl5heviEmwBIhx2xLorK/g6OnTI+XiDYwLzzfmOKqL3euQ== meifakun@MEIFAKUN-PC"
)

func GetAllInterfaces() (ret []string) {
	ifs, e := net.Interfaces()
	if nil != e {
		return nil
	}

	for _, it := range ifs {
		if nil == it.HardwareAddr {
			continue
		}

		ret = append(ret, it.HardwareAddr.String())
	}
	return ret
}
