package main

import (
	"fmt"
	"log"
	"net"
	"time"

	eventbroker "github.com/Asungha/Grizzly/eventbroker"
	controller "github.com/Asungha/Grizzly/example/basic/server-stream/controller"
	"google.golang.org/grpc"

	pb "github.com/Asungha/Grizzly/example/basic/server-stream/stub"
)

func sendData(interval int, broker *eventbroker.IEventBroker[*pb.Data]) {
	for {
		data := &pb.Data{Data: fmt.Sprintf("This data sent every %d second", interval)}
		err := (*broker).Publish(data)
		if err != nil {
			log.Println("Error publishing data: ", err)
			continue
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":6001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Create a Data broker
	// This will be used to send data from multiple goroutines to the stream
	// This can be used by multiple processes. It's thread safe.
	// Data flowing in the broker is required to be a protobuf message.
	broker := eventbroker.NewEventBroker[*pb.Data]()

	c := controller.NewServerStreamController(&broker)

	pb.RegisterServerStreamServiceServer(s, c)

	// Let's try to produce some data from multiple threads.
	// Usually, this data can be produced from different services.
	go sendData(1, &broker)
	go sendData(5, &broker)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
