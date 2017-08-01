package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var timeout float64
var postgresURL string
var keys = make(map[string]bool)
var waiting = make(map[string]bool)

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
		return true

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
		return true
	} else {
		return keys[key]
	}
}

func waitFor(key string) bool {
	waiting[key] = true
	fmt.Printf("'%s' : Waiting ...\n", key)
	t := time.Now()
	for {
		if keyExists(key) {
			fmt.Printf("'%s' : Waiting - OK\n", key)
			delete(waiting, key)
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
	keys[key] = true
	return true
}

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	if strings.ToLower(r.Method) == "get" {
		if key == "/" {
			fmt.Fprintln(w, "web-rendezvous")
			fmt.Fprintln(w, "\nCurrently waiting for + recently failed:")
			for k := range waiting {
				fmt.Fprintln(w, "\t", k)
			}
			fmt.Fprintln(w, "Total :", len(waiting))
		} else if key == "/favicon.ico" {
			w.WriteHeader(404)
		} else {
			if !waitFor(key) {
				w.WriteHeader(404)
				fmt.Fprintf(w, "Timeout : '%s'", key)
			} else {
				w.WriteHeader(200)
				fmt.Fprint(w, "OK")
			}
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
	flag.StringVar(&postgresURL, "postgresURL", "postgres://postgres:pgpass@postgres:5432/{{.}}?sslmode=disable", "Postgres URL of the form 'postgres://<username>:<password>@<host>:<port>/{{.}}[?<connection-string-params]'. {{.}} will be replaced with the db name")
	flag.Parse()

	fmt.Println("web-rendezvous")
	fmt.Printf("GET will timeout after: %.0f seconds\n", timeout)
	fmt.Printf("Listening on port: %s\n", port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
