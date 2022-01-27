# Simple CloudEvents Application using Go and Gorilla MUX
Simple example in Go using CloudEvents SDK. This example exposes two HTTP REST endpoints, a producer and a consumer of CloudEvents. 

- `/produce` accepts a POST request and produce a CloudEvent to the `SINK` that by default is `http://localhost:8080/`

This example uses [Gorilla MUX](https://github.com/gorilla/mux) to handle different requests to different endpoints. 

The application by default runs on port 8081, this can be overriden by exporting an environment variable `SERVER_PORT`.


## Building and running the application

If you want to build from source and run it you need to have Go installed. 

To simple run the application: 

```
go run main.go
```

You can use Google `ko` to build an image and push a docker image, if you have `ko` installed you can run:

```
ko publish main.go
```

To build, publish and run a docker image with `ko`:

```
docker run -p 8081:8081 $(ko publish fmtok8s-go-cloudevents.go)
```

Or if you just want to use Docker:

```
docker run -e SINK=http://localhost:8080 -p 8081:8081 salaboy/fmtok8s-go-cloudevents.go
```

## Sending Requests and CloudEvents

Send a POST request to produce a CloudEvent, that will be sent to a downstream service. 

```
curl -X POST localhost:8081/produce -v
```

Consume a CloudEvent and print it out: 
```
curl -X POST http://localhost:8081/ -H "Content-Type: application/json" -H "ce-type: MyCloudEvent"  -H "ce-id: 123"  -H "ce-specversion: 1.0" -H "ce-source: curl-command" -d '{"myData" : "hello from curl", "myCounter" : 1 }'
```