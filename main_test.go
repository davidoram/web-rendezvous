// main_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetState() {
	waiting.Range(func(k, v interface{}) bool {
		waiting.Delete(k)
		return true
	})
	marked.Range(func(k, v interface{}) bool {
		marked.Delete(k)
		return true
	})
}
func TestLockAndWait(t *testing.T) {
	resetState()
	timeout = 1.0

	// Create a request to GET key 'abc'
	getReq, err := http.NewRequest("GET", "/abc", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to POST key 'abc'
	postReq, err := http.NewRequest("POST", "/abc", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to GET the status
	statusReq, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check the initial status is ok
	statusResponse_1 := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)
	handler.ServeHTTP(statusResponse_1, statusReq)
	if status := statusResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"ok":true,"waiting_and_failed":[],"marked":[]}`
	if statusResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			statusResponse_1.Body.String(), expected)
	}

	// Perform a GET, and at this stage we haven't 'marked' the key, so should return a 404
	getResponse_1 := httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(getResponse_1, getReq)
	if status := getResponse_1.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
	expected = `{"ok":false,"error":"Timeout"}`
	if getResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			getResponse_1.Body.String(), expected)
	}

	// Check the status is ok
	statusResponse_1 = httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(statusResponse_1, statusReq)
	if status := statusResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"ok":true,"waiting_and_failed":["/abc"],"marked":[]}`
	if statusResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			statusResponse_1.Body.String(), expected)
	}

	// Perform a POST to 'mark' the key
	postResponse_1 := httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(postResponse_1, postReq)
	if status := postResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"ok":true}`
	if postResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			postResponse_1.Body.String(), expected)
	}

	// Check the status is ok
	statusResponse_1 = httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(statusResponse_1, statusReq)
	if status := statusResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"ok":true,"waiting_and_failed":[],"marked":["/abc"]}`
	if statusResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			statusResponse_1.Body.String(), expected)
	}

}

func TestInvalidRequestReservedUrl(t *testing.T) {
	resetState()
	timeout = 1.0

	// Create a request to POST key '_abc', a reserved URL because it starts with an _
	postReq, err := http.NewRequest("POST", "/_abc", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Perform a POST to 'mark' the key
	postResponse_1 := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)
	handler.ServeHTTP(postResponse_1, postReq)
	if status := postResponse_1.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
	expected := `{"ok":false,"error":"Keys starting with '_' are reserved"}`
	if postResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			postResponse_1.Body.String(), expected)
	}
}

func TestPostgresIntegration(t *testing.T) {

	resetState()
	// Setup connection details for postgres (running on travis)
	postgresURL = "postgres://postgres:@localhost:5432/{{.}}?sslmode=disable"
	timeout = 1.0

	// Create a request to GET key '_postgres/my_db'
	getReq_1, err := http.NewRequest("GET", "/_postgres/my_db", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to GET key '_postgres/no_such_db'
	getReq_2, err := http.NewRequest("GET", "/_postgres/no_such_db", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to GET the status
	statusReq, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Perform a GET to 'check' a db that exists
	getResponse_1 := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)
	handler.ServeHTTP(getResponse_1, getReq_1)
	if status := getResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"ok":true}`
	if getResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			getResponse_1.Body.String(), expected)
	}

	// Check the status is ok
	statusResponse_1 := httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(statusResponse_1, statusReq)
	if status := statusResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"ok":true,"waiting_and_failed":[],"marked":["/_postgres/my_db"]}`
	if statusResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			statusResponse_1.Body.String(), expected)
	}

	// Perform a GET to 'check' a db that doesn't exist
	getResponse_2 := httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(getResponse_2, getReq_2)
	if status := getResponse_2.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
	expected = `{"ok":false,"error":"Timeout"}`
	if getResponse_2.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			getResponse_2.Body.String(), expected)
	}

	// Check the status is ok
	statusResponse_1 = httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(statusResponse_1, statusReq)
	if status := statusResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"ok":true,"waiting_and_failed":["/_postgres/no_such_db"],"marked":["/_postgres/my_db"]}`
	if statusResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			statusResponse_1.Body.String(), expected)
	}

}

func TestServerListeningIntegration(t *testing.T) {

	resetState()
	timeout = 1.0

	// Create a request to GET key '_port/localhost/11211' (memcached)
	getReq_1, err := http.NewRequest("GET", "/_port/localhost/11211", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to GET key '_port/localhost/1' which wont be there
	getReq_2, err := http.NewRequest("GET", "/_port/localhost/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to GET the status
	statusReq, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Perform a GET to 'check' a port that has a server on it
	getResponse_1 := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)
	handler.ServeHTTP(getResponse_1, getReq_1)
	if status := getResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"ok":true}`
	if getResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			getResponse_1.Body.String(), expected)
	}

	// Check the status is ok
	statusResponse_1 := httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(statusResponse_1, statusReq)
	if status := statusResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"ok":true,"waiting_and_failed":[],"marked":["/_port/localhost/11211"]}`
	if statusResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			statusResponse_1.Body.String(), expected)
	}

	// Perform a GET to 'check' a db that doesn't exist
	getResponse_2 := httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(getResponse_2, getReq_2)
	if status := getResponse_2.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
	expected = `{"ok":false,"error":"Timeout"}`
	if getResponse_2.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			getResponse_2.Body.String(), expected)
	}

	// Check the status is ok
	statusResponse_1 = httptest.NewRecorder()
	handler = http.HandlerFunc(Handler)
	handler.ServeHTTP(statusResponse_1, statusReq)
	if status := statusResponse_1.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"ok":true,"waiting_and_failed":["/_port/localhost/1"],"marked":["/_port/localhost/11211"]}`
	if statusResponse_1.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			statusResponse_1.Body.String(), expected)
	}

}
