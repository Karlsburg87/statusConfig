package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	statusSentry "github.com/karlsburg87/statusSentry/pkg/configuration"
)

//parseIncomingConfig parses a config from the body of a request query
func parseIncomingConfig(w http.ResponseWriter, r *http.Request) (statusSentry.Config, error) {
	body := statusSentry.Config{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not decode request body as JSON: %v)", err)}); err != nil {
			log.Println(err)
		}
		return statusSentry.Config{}, err
	}
	return body, nil
}

//checkMethod checks whether the request is of the required method and if not returns an error after sending BadRequest
func checkMethod(method string, w http.ResponseWriter, r *http.Request) error {
	if r.Method != method {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "endpoint will only accept POST http requests"}); err != nil {
			log.Println(err)
		}
		return fmt.Errorf("incorrect http method")
	}
	return nil
}

//useNonZeroVal checks if value 1 is its zero value - if it is returns val2. Otherwise returns val1
//
//Uses generics
func useNonZeroVal[v comparable](val1, val2 v) v {
	var zeroValue v
	if val1 == zeroValue {
		return val2
	}
	return val1
}

//parseAndCheckServiceNameExists parses the response to Config and checks that the Config has a non null ServiceName field
func parseAndCheckServiceNameExists(w http.ResponseWriter, r *http.Request) (statusSentry.Config, error) {
	body, err := parseIncomingConfig(w, r)
	if err != nil {
		return body, err
	}
	//Must have service name
	if body.ServiceName == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "must include Service Name in the payload to update"}); err != nil {
			log.Println(err)
		}
		return body, err
	}
	return body, nil
}

//returnJson adds the content type application/json to responseWriter
func returnJson(w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json")
}
