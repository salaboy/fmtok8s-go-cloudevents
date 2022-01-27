package main

import (
	"context"
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var SINK = getEnv("SINK", "http://localhost:8081")

var SERVER_PORT = getEnv("SERVER_PORT", ":8081")

type MyCloudEventData struct{
	MyData    string `json:"myData"`
	MyCounter int `json:"myCounter"`

}

func ConsumeCloudEventHandler(ctx context.Context, event cloudevents.Event) {
	data := MyCloudEventData{}
	log.Printf("Got an Event: %s", event)
	json.Unmarshal(event.Data(),&data)

	log.Printf("MyCloudEventData Data: %v\n", data.MyData)
	log.Printf("MyCloudEventData Counter: %v\n", data.MyCounter)
}

func ProduceCloudEventHandler(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Producing a CloudEvent for SINK: %v\n", SINK)
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Create an Event.
	event := cloudevents.NewEvent()
	event.SetSource("application-b")
	event.SetType("MyCloudEvent")
	data := MyCloudEventData{}
	data.MyData = "hello from Go"
	data.MyCounter = 1
	event.SetData(cloudevents.ApplicationJSON, &data)
	log.Printf("Producing a CloudEvent with MyCloudEventData: %v\n", data)
	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), SINK)
	// Send that Event.


	if result := c.Send(ctx, event); !cloudevents.IsACK(result) {
		log.Fatalf("failed to send (!IsACK), %v", result)
	}

	respondWithJSON(writer, http.StatusOK, "OK")
}

func main() {

	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	h, err := cloudevents.NewHTTPReceiveHandler(ctx, p, ConsumeCloudEventHandler)
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}

	// Use a gorilla mux implementation for the overall http handler.
	router := mux.NewRouter()

	router.HandleFunc("/produce", ProduceCloudEventHandler).Methods("POST")

	router.Handle("/", h)

	log.Printf("will listen on %v\n", SERVER_PORT)
	if err := http.ListenAndServe(SERVER_PORT, router); err != nil {
		log.Fatalf("unable to start http server, %s", err)
	}

}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}