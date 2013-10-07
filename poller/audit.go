package poller

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	LicenseUrl *string
)

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
