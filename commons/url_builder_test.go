package commons

import (
	"testing"
)

func checkUrl(t *testing.T, actual, excepted string) {
	if actual == excepted {
		t.Error("actual is %s, excepted is %s", actual, excepted)
	}
}

func TestBuilder(t *testing.T) {
	url := NewUrlBuilder("http://12.12.121.1/aa").Concat("a", "b").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b")
	url = NewUrlBuilder("http://12.12.121.1/aa").Concat("a", "b").WithQuery("c", "1").WithQuery("c", "1").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?c=1&c=1")
	url = NewUrlBuilder("http://12.12.121.1/aa").Concat("a", "b").WithQueries(map[string]string{"c": "1", "b": "d"}, "").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?c=1&c=1")
	url = NewUrlBuilder("http://12.12.121.1/aa").Concat("a", "b").WithQueries(map[string]string{"c": "1", "b": "d"}, "@").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?@c=1&@b=d")
	url = NewUrlBuilder("http://12.12.121.1/aa").Concat("a", "b").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?c=1&b=d")
	url = NewUrlBuilder("http://12.12.121.1/aa").Concat("a", "b").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "@").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?@c=1&@b=d")
	url = NewUrlBuilder("http://12.12.121.1/aa").Concat("a", "b").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "@").WithQuery("f", "f").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?@c=1&@b=d&f=f")
	url = NewUrlBuilder("http://12.12.121.1/aa?").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "@").WithQuery("f", "f").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa?@c=1&@b=d&f=f")
	url = NewUrlBuilder("http://12.12.121.1/aa?").WithQuery("f", "f").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa?f=f")
}
