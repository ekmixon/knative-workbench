package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type Timestamp struct {
	Time string `json:"msg,omitempty"`
}

type eventSender struct {
	client cloudevents.Client
	target   string
}

func (s *eventSender) send(ctx context.Context, ts Timestamp) {
	newEvent := cloudevents.NewEvent()
	newEvent.SetID(uuid.New().String())
	newEvent.SetSource("eventing/time-emitter")
	newEvent.SetType("github.com/vladimirvivien/knative/eventing/time-emitter")
	if err := newEvent.SetData(cloudevents.ApplicationJSON, ts); err != nil {
		log.Printf("Error: %s", err)
	}
	log.Printf("Sending timestamp: %s", newEvent)
	ctx = cloudevents.ContextWithTarget(ctx, s.target)
	result := s.client.Send(ctx, newEvent)
	if result.Error() != ""{
		log.Printf("Error: %s", result.Error())
	}
}

func main() {
	target := os.Getenv("KN_TARGET")
	if target == "" {
		log.Fatal("env variable KN_TARGET not set")
	}

	client, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("Cloud-event client failed: %v", err)
	}
	sender := &eventSender{client: client, target: target}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	delay := 5 * time.Second
	log.Println("Starting time emitter (ctrl+c to stop)")
	log.Printf("Sending every %s", delay)

	for {
		select {
		case <-time.After(delay):
			sender.send(ctx, Timestamp{Time: time.Now().String()})
		case <-ctx.Done():
			log.Println("\nStopping...")
			stop()
			return
		}
	}
}
