package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/karlsburg87/statusConfig/internal/manager"
	statusSentry "github.com/karlsburg87/statusSentry/pkg/configuration"
)

//Get handles requests to get a single Config within the Configeration list
//
//The Service Name to get is in the `q` url query key of the GET request
func Get(manger *manager.Manager, w http.ResponseWriter, r *http.Request) {
	returnJson(w)
	if err := checkMethod(http.MethodGet, w, r); err != nil {
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "query key `q` must contain a Service Name of Config object to get"}); err != nil {
			log.Println(err)
		}
		return
	}
	//serve Configerations to statusSentry services
	payload, err := manger.GetFromDb(q)
	if err != nil {
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not fetch Config from datastore: %v (q: %+v", err, q)}); err != nil {
			log.Println(err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(&payload); err != nil {
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not encode Config to response body: %v ", err)}); err != nil {
			log.Println(err)
		}
		return
	}
}

//Add adds the POST body JSON representation of a Config to the Configeration list
func Add(manger *manager.Manager, w http.ResponseWriter, r *http.Request) {
	returnJson(w)
	if err := checkMethod(http.MethodPost, w, r); err != nil {
		return
	}
	body, err := parseAndCheckServiceNameExists(w, r)
	if err != nil {
		return
	}

	//add
	if err := manger.AddToDb(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not add Config to datastore: %v (config: %+v", err, body)}); err != nil {
			log.Println(err)
		}
		return
	}
}

//Remove handles GET requests with a Service Name as the value on the  `q`
// query string key pointing to a Service that should be deleted from the list
func Remove(manger *manager.Manager, w http.ResponseWriter, r *http.Request) {
	returnJson(w)
	if err := checkMethod(http.MethodDelete, w, r); err != nil {
		return
	}
	body, err := parseAndCheckServiceNameExists(w, r)
	if err != nil {
		return
	}

	if err := manger.DeleteFromDb(body.ServiceName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not delete Config from datastore: %v (q: %+v)", err, body.ServiceName)}); err != nil {
			log.Println(err)
		}
		return
	}
}

//Update handles calls to update an existing Config with new values via a Config update payload.
// Only update fields that are present in the payload
//
//Operates by deleting the original and re adding the merged copy
func Update(manger *manager.Manager, w http.ResponseWriter, r *http.Request) {
	returnJson(w)
	//Must be post
	if err := checkMethod(http.MethodPatch, w, r); err != nil {
		return
	}
	//Get payload
	body, err := parseAndCheckServiceNameExists(w, r)
	if err != nil {
		return
	}

	//get current
	payload, err := manger.GetFromDb(body.ServiceName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not fetch Config from datastore: %v (q: %+v", err, body.ServiceName)}); err != nil {
			log.Println(err)
		}
		return
	}
	//merge body from request and payload from datastore
	nw := statusSentry.Config{
		ServiceName:   body.ServiceName,
		DisplayDomain: useNonZeroVal(payload.DisplayDomain, body.DisplayDomain),
		StatusPage:    useNonZeroVal(payload.StatusPage, body.StatusPage),
		TargetHook:    useNonZeroVal(payload.TargetHook, body.TargetHook),
		PollFrequency: useNonZeroVal(payload.PollFrequency, body.PollFrequency),
	}
	nw.PollPages = body.PollPages
	if payload.PollPages != nil {
		nw.PollPages = payload.PollPages
	}
	//delete current
	if err := manger.DeleteFromDb(payload.ServiceName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not delete Config from datastore: %v (q: %+v)", err, payload.ServiceName)}); err != nil {
			log.Println(err)
		}
		return
	}
	//add new updated item to datastore
	if err := manger.AddToDb(nw); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not add Config to datastore: %v (config: %+v)", err, nw)}); err != nil {
			log.Println(err)
		}
		return
	}

}

//FetchConfig handles incoming requests from statusSentry instances wanting an updated Configeration JSON file
func FetchConfig(manger *manager.Manager, w http.ResponseWriter, r *http.Request) {
	returnJson(w)
	if err := checkMethod(http.MethodGet, w, r); err != nil {
		return
	}

	if err := manger.DumpToConfigeration(w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("could not return config : %v", err)}); err != nil {
			log.Println(err)
		}
		return
	}

}
