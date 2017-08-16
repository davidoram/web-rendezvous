package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var timeout float64
var postgresURL string

// Map of keys that have been 'marked' as existing
var keys = make(map[string]bool)
var keysLock = sync.RWMutex{}

// Map of keys that we are waiting to be 'marked' currently, or waiting for and failed
var waiting = make(map[string]bool)
var waitingLock = sync.RWMutex{}

// Check if a key exists.
// key:  _postgres/dbname - Will check if database 'dbname' exists
// key:  _port/host/port  - Will check if can open a TCP connection to that host/port
// All other keys will simply check if keys[key] exists
func keyExists(key string) bool {
	if strings.HasPrefix(key, "/_postgres/") {
		// Postgres connection?
		split := strings.Split(key, "/")
		if len(split) != 3 {
			fmt.Printf("'%s' : Error: unable to parse postgres db from : '%s' ...\n", key, key)
			return false
		}
		database := split[2]
		fmt.Printf("'%s' : Testing for postgres db: '%s' ...\n", key, database)
		t := template.Must(template.New("dbConnection").Parse(postgresURL))
		var tpl bytes.Buffer
		if err := t.Execute(&tpl, database); err != nil {
			fmt.Printf("'%s' : Error generating postgres connection template: '%s', error: '%s' ...\n", key, postgresURL, err)
			return false
		}

		db, err := sql.Open("postgres", tpl.String())
		defer db.Close()
		if err == nil {
			if _, err := db.Query("SELECT 1"); err != nil {
				// fmt.Printf("'%s' : Can't run query 'select 1' in postgres db: '%s', error: '%s' ...\n", key, database, err)
				return false
			}
		} else {
			fmt.Printf("'%s' : Can't connect to postgres db: '%s', error: '%s' ...\n", key, database, err)
			return false
		}
		fmt.Printf("'%s' : postgres db '%s' created - OK\n", key, database)
		return mark(key)

	} else if strings.HasPrefix(key, "/_port/") {
		// Port connection?
		split := strings.Split(key, "/")
		if len(split) != 4 {
			fmt.Printf("'%s' : Error: unable to parse host/port from : '%s' ...\n", key, key)
			return false
		}
		host := split[2]
		port := split[3]
		fmt.Printf("'%s' : Connecting to host:port: '%s:%s' ...\n", key, host, port)
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
		if err != nil {
			// fmt.Printf("'%s' : Can't connect to host/port: '%s:%s', error: '%s\n", key, host, port, err)
			return false
		}
		conn.Close()
		fmt.Printf("'%s' : host:port: '%s:%s' listening - OK\n", key, host, port)
		return mark(key)
	} else {
		keysLock.RLock()
		defer keysLock.RUnlock()
		return keys[key]
	}
}

func deleteFromWaiting(key string) {
	waitingLock.Lock()
	delete(waiting, key)
	waitingLock.Unlock()
}

func waitFor(key string) bool {
	waitingLock.Lock()
	waiting[key] = true
	waitingLock.Unlock()
	fmt.Printf("'%s' : Waiting ...\n", key)
	t := time.Now()
	for {
		if keyExists(key) {
			fmt.Printf("'%s' : Waiting - OK\n", key)
			deleteFromWaiting(key)
			return true
		}
		time.Sleep(500 * time.Millisecond)
		if time.Since(t).Seconds() > timeout {
			fmt.Printf("'%s' : Waiting - Timeout\n", key)
			return false
		}
	}
}

func mark(key string) bool {
	fmt.Printf("'%s' : Marked - OK\n", key)
	deleteFromWaiting(key)
	keysLock.Lock()
	defer keysLock.Unlock()
	keys[key] = true
	return true
}

type StandardResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type RootResponse struct {
	StandardResponse
	WaitKeys   []string `json:"waiting_and_failed"`
	MarkedKeys []string `json:"marked"`
}

func NewRootResponse() RootResponse {
	r := new(RootResponse)
	r.Ok = true
	r.Error = ""
	r.WaitKeys = make([]string, 0)
	r.MarkedKeys = make([]string, 0)
	return *r
}

func Handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	if strings.ToLower(r.Method) == "get" {
		if key == "/" {
			response := NewRootResponse()
			waitingLock.RLock()
			for k := range waiting {
				response.WaitKeys = append(response.WaitKeys, k)
			}
			waitingLock.RUnlock()
			keysLock.RLock()
			for k := range keys {
				response.MarkedKeys = append(response.MarkedKeys, k)
			}
			keysLock.RUnlock()
			jResponse, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Cant marshall response: %v", response)
			}
			fmt.Fprint(w, string(jResponse))
		} else if key == "/favicon.ico" {
			w.WriteHeader(404)
		} else {
			if !waitFor(key) {
				w.WriteHeader(404)
				response := StandardResponse{Ok: false, Error: "Timeout"}
				jResponse, err := json.Marshal(response)
				if err != nil {
					log.Fatalf("Cant marshall response: %v", response)
				}
				fmt.Fprint(w, string(jResponse))
			} else {
				w.WriteHeader(200)
				response := StandardResponse{Ok: true}
				jResponse, err := json.Marshal(response)
				if err != nil {
					log.Fatalf("Cant marshall response: %v", response)
				}
				fmt.Fprint(w, string(jResponse))
			}
		}
	} else if strings.ToLower(r.Method) == "put" || strings.ToLower(r.Method) == "post" {
		if strings.HasPrefix(key, "/_") {
			w.WriteHeader(404)
			response := StandardResponse{Ok: false, Error: "Keys starting with '_' are reserved"}
			jResponse, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Cant marshall response: %v", response)
			}
			fmt.Fprint(w, string(jResponse))
		} else {
			mark(key)
			w.WriteHeader(200)
			response := StandardResponse{Ok: true}
			jResponse, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Cant marshall response: %v", response)
			}
			fmt.Fprint(w, string(jResponse))
		}
	} else {
		response := StandardResponse{Ok: false, Error: "I dont understand that HTTP method, try 'put', 'post' or 'get'"}
		jResponse, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Cant marshall response: %v", response)
		}
		fmt.Fprint(w, string(jResponse))
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "Port to listen on")
	flag.Float64Var(&timeout, "timeout", 30.0, "Timeout in seconds")
	flag.StringVar(&postgresURL, "postgresURL", "postgres://postgres:pgpass@postgres:5432/{{.}}?sslmode=disable", "Postgres URL of the form 'postgres://<username>:<password>@<host>:<port>/{{.}}[?<connection-string-params]'. {{.}} will be replaced with the db name")
	flag.Parse()

	fmt.Println("web-rendezvous")
	fmt.Printf("GET will timeout after: %.0f seconds\n", timeout)
	fmt.Printf("Listening on port: %s\n", port)
	http.HandleFunc("/", Handler)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
