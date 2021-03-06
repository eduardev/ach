// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moov-io/ach"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func TestFiles__createFileEndpoint(t *testing.T) {
	repo := NewRepositoryInMemory(testTTLDuration, nil)
	svc := NewService(repo)

	body := strings.NewReader(`{"random":"json"}`)

	resp, err := createFileEndpoint(svc, repo, nil)(context.TODO(), body)
	r, ok := resp.(createFileResponse)
	if !ok {
		t.Errorf("got %#v", resp)
	}
	if err == nil || r.Err == nil {
		t.Errorf("expected error: err=%v resp.Err=%v", err, r.Err)
	}

}

func TestFiles__getFilesEndpoint(t *testing.T) {
	repo := NewRepositoryInMemory(testTTLDuration, nil)
	svc := NewService(repo)

	body := strings.NewReader(`{"random":"json"}`)

	resp, err := getFilesEndpoint(svc)(context.TODO(), body)
	_, ok := resp.(getFilesResponse)
	if !ok || err != nil {
		t.Errorf("got %#v : err=%v", resp, err)
	}
}

func TestFiles__getFileEndpoint(t *testing.T) {
	repo := NewRepositoryInMemory(testTTLDuration, nil)
	svc := NewService(repo)

	body := strings.NewReader(`{"random":"json"}`)

	resp, err := getFileEndpoint(svc, nil)(context.TODO(), body)
	r, ok := resp.(getFileResponse)
	if !ok {
		t.Errorf("got %#v", resp)
	}
	if err == nil || r.Err == nil {
		t.Errorf("expected error: err=%v resp.Err=%v", err, r.Err)
	}

}

func TestFiles__getFileContentsEndpoint(t *testing.T) {
	repo := NewRepositoryInMemory(testTTLDuration, nil)
	svc := NewService(repo)

	body := strings.NewReader(`{"random":"json"}`)

	resp, err := getFileContentsEndpoint(svc, nil)(context.TODO(), body)
	_, ok := resp.(getFileContentsResponse)
	if !ok {
		t.Errorf("got %#v", resp)
	}
	if err == nil {
		t.Errorf("expected error: err=%v", err)
	}

}

func TestFiles__validateFileEndpoint(t *testing.T) {
	repo := NewRepositoryInMemory(testTTLDuration, nil)
	svc := NewService(repo)

	rawBody := `{"random":"json"}`

	resp, err := validateFileEndpoint(svc, nil)(context.TODO(), strings.NewReader(rawBody))
	r, ok := resp.(validateFileResponse)
	if !ok {
		t.Errorf("got %#v", resp)
	}
	if err == nil || r.Err == nil {
		t.Errorf("expected error: err=%v resp.Err=%v", err, r.Err)
	}

	// write an ACH file into repository
	fd, err := os.Open(filepath.Join("..", "test", "testdata", "ppd-valid.json"))
	if fd == nil {
		t.Fatalf("empty ACH file: %v", err)
	}
	defer fd.Close()
	bs, _ := ioutil.ReadAll(fd)
	file, _ := ach.FileFromJSON(bs)
	file.Header.ImmediateDestination = "" // invalid routing number
	repo.StoreFile(file)

	// test status code
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/files/%s/validate", file.ID), strings.NewReader(rawBody))

	router := mux.NewRouter()
	router.Methods("GET").Path("/files/{id}/validate").Handler(
		httptransport.NewServer(validateFileEndpoint(svc, nil), decodeValidateFileRequest, encodeResponse),
	)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
	if !strings.HasPrefix(w.Body.String(), `{"error":"invalid ACH file: ImmediateDestination`) {
		t.Errorf("unknown error: %v", err)
	}
}
