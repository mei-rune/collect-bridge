package license

import (
	//"crypto/tls"
	//"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	LicenseUrl    *string
	https_key_pem = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDULGIb5iRx3ro3YLPvKDBgYXR4WtS5jRAr+JLUmN2/Qrykmugk
0nexfWNAeL4+c8PHfQUcIDPxG9fcMcbZhOMoqJPbpW6gb6lDWwJc+1071uJyXLfx
o52VyHRmdgbTqbVuFwpruO2PN1GLgosZFTsAzv7SIWXlf7iEYbXDQvzJrQIDAQAB
AoGABC0jeT0cdolVzQVARXLbEOqaKkxPZ5+ZM+Rer4SayMr9f5f0/dSOupWnQHuf
8wbAUcrqMNWJcDOGGjV4not5FSie0PqxNF+hFrg9IdO6ulaeewrSaqav7LkDax95
zupomKwPYZv2qgjbslTg44MwzztDB3lkgJmnv/fGaIxWWoECQQDu1GAGg3hWtVjg
xYFe/xiQ+Kv1tly4RnDErc6RE6gBKmpEUdWw1LhnMcoZ2hK0Duogrt+g3ZpyITDy
ktuSsS59AkEA421nefhwkIazC4hBcWR5h6LMGOymj4rbzJQo+2NqU054xzNLJc9p
wq7Mzg43+eoems//6bcg81OH60qH41D+8QJAWALNbDkQpKtpmFNQTIinLe1luUO9
wW676c6/G7lppRxTUt/xZpvNZMH1Xzd8wvvoDalD4c0oODzBA/NYlSNUJQJAKJPD
i5qNEuxFk8Aq1P11RYMBYU0P5rqCvvyMV1YEiXqNyBTZypQ4LXkcp4MX76oa7cpA
wcVfxqpXrN5uYlt4MQJBAIoJzgiJMMVrWGbW2mgU4i4Evlhuv0GZLHrE8LtrJr4I
sTwMyIIVeSpCvbUFSiYQ2B8ROk1QF5MqrHn0fpzdziI=
-----END RSA PRIVATE KEY-----`

	https_cacert_pem = `-----BEGIN CERTIFICATE-----
MIIC5DCCAk2gAwIBAgIJAIlYEins3yUHMA0GCSqGSIb3DQEBBQUAMIGKMQswCQYD
VQQGEwJjbjEOMAwGA1UECAwFY2hpbmExETAPBgNVBAcMCHNoYW5naGFpMQwwCgYD
VQQKDAN0cHQxEDAOBgNVBAsMB2RldmVsb3AxEzARBgNVBAMMCnJ1bm5lci1tZWkx
IzAhBgkqhkiG9w0BCQEWFHJ1bm5lci5tZWlAZ21haWwuY29tMB4XDTEzMDkyNzAz
MzIyOFoXDTE2MDkyNjAzMzIyOFowgYoxCzAJBgNVBAYTAmNuMQ4wDAYDVQQIDAVj
aGluYTERMA8GA1UEBwwIc2hhbmdoYWkxDDAKBgNVBAoMA3RwdDEQMA4GA1UECwwH
ZGV2ZWxvcDETMBEGA1UEAwwKcnVubmVyLW1laTEjMCEGCSqGSIb3DQEJARYUcnVu
bmVyLm1laUBnbWFpbC5jb20wgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANQs
YhvmJHHeujdgs+8oMGBhdHha1LmNECv4ktSY3b9CvKSa6CTSd7F9Y0B4vj5zw8d9
BRwgM/Eb19wxxtmE4yiok9ulbqBvqUNbAlz7XTvW4nJct/GjnZXIdGZ2BtOptW4X
Cmu47Y83UYuCixkVOwDO/tIhZeV/uIRhtcNC/MmtAgMBAAGjUDBOMB0GA1UdDgQW
BBQQdlJ+MjJCBmKJLHmzZXzArFTJnzAfBgNVHSMEGDAWgBQQdlJ+MjJCBmKJLHmz
ZXzArFTJnzAMBgNVHRMEBTADAQH/MA0GCSqGSIb3DQEBBQUAA4GBABbYXEeZVK6F
XV/RhABmE0gKIiHhM4ZNiQQUw8XJMdckKB1hrPTPfy8HvWTRU1C69URLe4OK+9ip
8oHlN9cK/bIMW0qvBjZbjIblnYEevNkm6dmCa0O5ZFYxA5Hz25LormD9fJYpibyC
wr2JK9s/CWEk81dSpCD2UNAOc/wLbJRA
-----END CERTIFICATE-----`
)

func init() {
	s := "http://127.0.0.1:37076/"
	if nil == LicenseUrl {
		LicenseUrl = &s
	}

	// if nil == http.DefaultTransport.(*http.Transport).TLSClientConfig {
	// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: x509.NewCertPool()}
	// } else if nil == http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs {
	// 	http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs = &x509.CertPool{}
	// }

	// http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs.AppendCertsFromPEM([]byte(https_cacert_pem))

	// if nil == http.DefaultTransport.(*http.Transport).TLSClientConfig {
	// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{}
	// }
	// certificates := http.DefaultTransport.(*http.Transport).TLSClientConfig.Certificates
	// certificate, e := tls.X509KeyPair([]byte(https_cacert_pem), []byte(https_key_pem))
	// if nil != e {
	// 	panic(e.Error())
	// }

	// certificates = append(certificates, certificate)
	// http.DefaultTransport.(*http.Transport).TLSClientConfig.Certificates = certificates
}

func SetInvalidLicense() error {
	_, _, e := Get(*LicenseUrl + "rth56w3")
	if nil != e {
		return e
	}
	return nil
}

func IsMatchNode(t string, count int64) (bool, error) {
	_, code, e := Get(*LicenseUrl + "l5472/" + t + "_" + fmt.Sprint(count))
	if nil != e {
		return true, e
	}
	if http.StatusUnauthorized == code {
		return false, e
	}
	return true, nil
}

func IsEnabledModule(nm string) (bool, error) {
	_, code, e := Get(*LicenseUrl + "o7456/" + nm)
	if nil != e {
		return true, e
	}
	if http.StatusUnauthorized == code {
		return false, e
	}
	return true, nil
}

func Get(url string) ([]byte, int, error) {
	// if nil == http.DefaultTransport.(*http.Transport).TLSClientConfig {
	// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: x509.NewCertPool()}
	// } else if nil == http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs {
	// 	http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs = &x509.CertPool{}
	// }

	// certPEMBlock, err := ioutil.ReadFile("../license/key.pem")
	// if err != nil {
	// 	return nil, 0, err
	// }

	// http.DefaultTransport.(*http.Transport).TLSClientConfig.RootCAs.AppendCertsFromPEM(certPEMBlock)
	resp, err := http.DefaultClient.Get(url)
	if nil != err {
		return nil, http.StatusOK, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return nil, http.StatusInternalServerError, err
	}
	if http.StatusOK != resp.StatusCode {
		return nil, resp.StatusCode, fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}
	return body, resp.StatusCode, nil
}
