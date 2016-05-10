package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

type gcmRequest struct {
	tokens  []string
	payload string
}

// Send a message to GCM
func send(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	var gcmReq gcmRequest
	err := decoder.Decode(&gcmReq)
	if err != nil {
		errText := "Couldn't decode json"
		log.Println(errText)
	}

	log.Println("Extracted tokens", gcmReq.tokens)

	tokens := gcmReq.tokens
	payload := gcmReq.payload

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
