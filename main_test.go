// main_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLockAndWait(t *testing.T) {

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

func TestInvalidRequests(t *testing.T) {

	// Create a request to POST key '_abc'
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
