package main

import (
	"github.com/gorilla/mux"
	"gopkg.in/olivere/elastic.v2"
	"log"
	"io"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var elasticClient *elastic.Client

func init() {
	if err := establishElasticConnection(); err != nil {
		log.Panicln(err)
	}
}

func respondWithJSON(w http.ResponseWriter, rootKey string, body interface{}, statusCode int) {
	data := make(map[string]interface{}, 1)
	data[rootKey] = body
	jsonString, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	io.WriteString(w, string(jsonString))
}

func createElasticClient() (*elastic.Client, error) {
	// Check for an elastic URL, which is probably missing in development
	// elasticURL := os.Getenv("ELASTIC_URL")
	// if elasticURL == "" {
	// 	elasticURL = "http://localhost:9200"
	// }
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	if res, err := client.ClusterHealth().WaitForStatus("red").Timeout("15s").Do(); err != nil {
		return nil, err
	} else if res.TimedOut {
		return nil, errors.New("time out waiting for cluster status red")
	}
	return client, nil
}

func establishElasticConnection() error {
	var err error
	for i := 0; i < 10; i++ {
		elasticClient, err = createElasticClient()
		if err != nil {
			log.Println("Sleeping one second to wait for elasticsearch")
			time.Sleep(time.Second)
		} else {
			return nil
		}
	}
	return err
}

func HelloServer(w http.ResponseWriter, req *http.Request) {

	vehicles, err := searchForVehicles(elasticClient, "fiat")
	if (err != nil) {
		log.Println(err)
		io.WriteString(w, "Shit went wrong")
		return
	}

	respondWithJSON(w, "vehicles", vehicles, http.StatusOK)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", HelloServer)
	// r.HandleFunc("/products", ProductsHandler)
	// r.HandleFunc("/articles", ArticlesHandler)

	log.Printf("Running server on 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
