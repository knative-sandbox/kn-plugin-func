package invoker

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/google/uuid"
)

const (
	DefaultEventSource = "/boson/fn"
	DefaultEventType   = "boson.fn"
)

type Invoker struct {
	Endpoint    string
	Source      string
	Type        string
	Id          string
	Data        string
	ContentType string
}

func NewInvoker() *Invoker {
	return &Invoker{
		Source:      DefaultEventSource,
		Type:        DefaultEventType,
		Id:          uuid.NewString(),
		Data:        "",
		ContentType: event.TextPlain,
	}
}

func (e *Invoker) Send(ctx context.Context, endpoint string) (err error) {
	c, err := newClient(endpoint)
	if err != nil {
		return
	}
	evt := event.Event{
		Context: event.EventContextV1{
			Type:   e.Type,
			Source: *types.ParseURIRef(e.Source),
			ID:     e.Id,
		}.AsV1(),
	}
	if err = evt.SetData(e.ContentType, e.Data); err != nil {
		return
	}
	event, result := c.Request(ctx, evt)
	if !cloudevents.IsACK(result) {
		return fmt.Errorf(result.Error())
	}
	if event != nil {
		fmt.Printf("%v", event)
	}
	return nil
}

func newClient(target string) (c client.Client, err error) {
	p, err := http.New(http.WithTarget(target))
	if err != nil {
		return
	}
	return client.New(p)
}
