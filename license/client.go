package license

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get(url string) ([]byte, int, error) {
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
