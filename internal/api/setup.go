package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/karlsburg87/statusConfig/internal/manager"
)

func Setup() manager.Manager {
	//_ = make(statusSentry.Configuration, 0)

	//Client - for getting base configeration and calling the config refetch endpoint on statusSentry services
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	//get manager
	manger, err := manager.NewManager(client)
	if err != nil {
		log.Fatalln(err)
	}

	//Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//Mux
	mux := http.NewServeMux()

	//healthcheck default endpoint -TODO: admin UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // default for Cloud Run health check
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "hello visitor. Try the `/get` endpoint"}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	//statusSentry fetch configeration endpoint
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		Get(manger, w, r)
	})

	//admin API for managing the configuration list
	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		Add(manger, w, r)
	})

	mux.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		Remove(manger, w, r)
	})

	mux.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		Update(manger, w, r)
	})

	mux.HandleFunc("/configuration", func(w http.ResponseWriter, r *http.Request) {
		Update(manger, w, r)
	})

	//Server setup
	server := &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           mux,
		Addr:              fmt.Sprintf(":%s", port),
	}

	manger.Server = server

	return *manger
}
