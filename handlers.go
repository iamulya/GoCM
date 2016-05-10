package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type gcmRequest struct {
	Tokens  []string               `json:"tokens"`
	Payload map[string]interface{} `json:"payload"`
}

// Send a message to GCM
func send(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errText := "Couldn't read body"
		log.Println(errText)
	}
	log.Println(string(body))
	var gcmReq gcmRequest
	unmarshallerr := json.Unmarshal(body, &gcmReq)
	if unmarshallerr != nil {
		errText := "Couldn't decode json"
		log.Println(errText)
	}

	log.Println("Extracted tokens", gcmReq.Tokens)
	log.Println("Extracted payload", gcmReq.Payload)

	tokens := gcmReq.Tokens
	payload := gcmReq.Payload

	go func() {
		incrementPending()
		sendMessageToGCM(tokens, payload)
	}()

	// Return immediately
	output := "ok\n"
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(output)))
	io.WriteString(w, output)
}

// Return a run report for this process
func getReport(w http.ResponseWriter, r *http.Request) {
	runReportMutex.Lock()
	a, _ := json.Marshal(runReport)
	runReportMutex.Unlock()
	b := string(a)
	io.WriteString(w, b)
}

// Return all currently collected canonical reports from GCM
func getCanonicalReport(w http.ResponseWriter, r *http.Request) {
	canonicalReplacementsMutex.Lock()
	ids := map[string][]canonicalReplacement{"canonical_replacements": canonicalReplacements}
	a, _ := json.Marshal(ids)
	canonicalReplacementsMutex.Unlock()

	b := string(a)
	io.WriteString(w, b)

	// Clear out canonicals
	go func() {
		canonicalReplacementsMutex.Lock()
		defer canonicalReplacementsMutex.Unlock()
		canonicalReplacements = nil
	}()
}

// Return all tokens that need to be unregistered
func getNotRegisteredReport(w http.ResponseWriter, r *http.Request) {
	notRegisteredMutex.Lock()
	ids := map[string][]string{"tokens": notRegisteredKeys}
	a, _ := json.Marshal(ids)
	notRegisteredMutex.Unlock()

	b := string(a)
	io.WriteString(w, b)

	// Clear ids
	go func() {
		notRegisteredMutex.Lock()
		defer notRegisteredMutex.Unlock()
		notRegisteredKeys = nil
	}()
}
