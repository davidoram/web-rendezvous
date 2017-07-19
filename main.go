package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var timeout float64
var keys = make(map[string]bool)

func waitFor(key string) bool {
	fmt.Printf("'%s' : Waiting ...\n", key)
	t := time.Now()
	for {
		if keys[key] {
			fmt.Printf("'%s' : Waiting - OK\n", key)
			return true
		}
		time.Sleep(100 * time.Millisecond)
		if time.Since(t).Seconds() > timeout {
			fmt.Printf("'%s' : Waiting - Timeout\n", key)
			return false
		}
	}
}

func mark(key string) bool {
	fmt.Printf("'%s' : Marked - OK\n", key)
	keys[key] = true
	return true
}

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	if strings.ToLower(r.Method) == "get" {
		if !waitFor(key) {
			w.WriteHeader(404)
			fmt.Fprintf(w, "Timeout : '%s'", key)
		} else {
			w.WriteHeader(200)
			fmt.Fprint(w, "OK")
		}
	} else if strings.ToLower(r.Method) == "put" || strings.ToLower(r.Method) == "post" {
		mark(key)
		w.WriteHeader(200)
		fmt.Fprint(w, "OK")
	} else {
		fmt.Fprintf(w, "I dont understand that HTTP method: '%s', try 'put', 'post' or 'get'", r.Method)
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "Port to listen on")
	flag.Float64Var(&timeout, "timeout", 30.0, "Timeout in seconds")
	flag.Parse()

	fmt.Println("web-rendezvous")
	fmt.Printf("GET will timeout after: %.0f seconds\n", timeout)
	fmt.Printf("Listening on port: %s\n", port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
