package manager

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	statusSentry "github.com/karlsburg87/statusSentry/pkg/configuration"
	bolt "go.etcd.io/bbolt"
)

//Manager is the management type for the config database
type Manager struct {
	tickr       time.Ticker //interval in which servers are told to update configeration if there have been changes (see NeedsReload)
	Datastore   *bolt.DB
	client      *http.Client
	Server      *http.Server
	NeedsReload bool //NeedsReload is whether the statusSentry servers need to reload the configeration at the next interval
}

//openDb creates or opens a bolt kv database for use.
//
//Must close the database after this call
func openDb() (*bolt.DB, error) {
	db, err := bolt.Open("configs.db", 0666, nil)
	if err != nil {
		return nil, err
	}
	//defer db.Close()
	return db, nil
}

//NewManager initializes a new Manager and starts the goroutine that triggers updates in statusSentry instances
func NewManager(client *http.Client) (*Manager, error) {
	db, err := openDb()
	if err != nil {
		return nil, err
	}
	m := &Manager{
		Datastore:   db,
		tickr:       *time.NewTicker(5 * time.Minute),
		NeedsReload: false,
		client:      client,
	}
	//launch goroutine to ping
	go m.nudge()

	return m, nil
}

//nudge is a goroutine method that calls the http endpoints of statusSentry instances to nudge them
// to call for configuration updates
//
//A list of comma seperated URL endpoints of statusSentry instances useful for this task are under
// the environment variable 'STATUSSENTRY_INSTANCES'
//
func (man Manager) nudge() {
	for range man.tickr.C {
		if !man.NeedsReload {
			continue
		}
		instances := os.Getenv("STATUSSENTRY_INSTANCES")
		if instances == "" {
			continue
		}
		instancesUnits := strings.Split(instances, ",")
		for _, ins := range instancesUnits {
			res, err := man.client.Get(ins)
			if err != nil {
				log.Printf("issue making nudge call to :'%s'\n", ins)
				continue
			}
			res.Body.Close()
		}
	}
}

//AddToDb adds a config to the database
func (man *Manager) AddToDb(toAdd statusSentry.Config) error {
	// Start a write transaction.
	if err := man.Datastore.Update(func(tx *bolt.Tx) error {
		// Create a bucket.
		b, err := tx.CreateBucket([]byte("configs"))
		if err != nil {
			return err
		}

		payload, err := json.Marshal(toAdd)
		if err != nil {
			return err
		}
		if err := b.Put([]byte(toAdd.ServiceName), payload); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	man.NeedsReload = true
	return nil
}

//DeleteFromDb deletes a config from the database by serviceName key.
func (man *Manager) DeleteFromDb(toDelete string) error {
	// Start a write transaction.
	if err := man.Datastore.Update(func(tx *bolt.Tx) error {
		// Create a bucket.
		b, err := tx.CreateBucket([]byte("configs"))
		if err != nil {
			return err
		}
		if err := b.Delete([]byte(toDelete)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	man.NeedsReload = true
	return nil
}

//GetFromDb fetches a single Config from the configeration by service name
func (man Manager) GetFromDb(toGet string) (statusSentry.Config, error) {
	out := statusSentry.Config{}
	if err := man.Datastore.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("configs"))
		v := b.Get([]byte(toGet))
		if err := json.Unmarshal(v, &out); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return statusSentry.Config{}, nil
	}
	return out, nil
}

//DumpToConfigeration dumps out the db as a statusSentry.Configeration structured JSON file.
// Returns output file location
func (man Manager) DumpToConfigeration(w http.ResponseWriter) error {
	//start stream out
	enc := json.NewEncoder(w)
	//stream out from database
	return man.Datastore.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("configs"))

		//create configeration with all configs in the db
		b.ForEach(func(k, v []byte) error {
			//decode singular object
			out := statusSentry.Config{}
			if err := json.Unmarshal(v, &out); err != nil {
				return err
			}
			if err := enc.Encode(out); err != nil {
				return err
			}
			return nil
		})
		return nil
	})
}

//LoadConfigeration loads a JSON file in statusSentry.Configeration structure to the bolt datastore
func (man Manager) LoadConfigeration(configLocation *url.URL) error {
	res, err := http.DefaultClient.Get(configLocation.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)
	for {
		out := statusSentry.Config{}
		err := dec.Decode(&out)
		if err != nil {
			return err
		}
		man.AddToDb(out)
	}
}
