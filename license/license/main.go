package main

import (
	"fmt"
	"license"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func main() {
	for i := 0; i < 256; i++ {
		m, s, r := license.GetHD(i)
		if 0 != r {
			break
		}
		fmt.Println(i, m, s)
	}

	http.HandleFunc("/", handler)
	log.Printf("About to listen on 10443. Go to https://127.0.0.1:10443/")
	err := http.ListenAndServeTLS(":10443", "cacert.pem", "key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}
