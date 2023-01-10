package subscriber

import (
	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/core/typedef"
)

type genericSubscriber struct {
	id string
	api.EventHandler
}

// NewGenericSubscriber returns the generic subscriber entity
func NewGenericSubscriber(handler api.EventHandler, id string) *genericSubscriber {
	return &genericSubscriber{
		id:           id,
		EventHandler: handler,
	}
}

// ID returns the unique ID of the subscriber
func (pub *genericSubscriber) ID() string {
	return pub.id
}

// NotifyFunc notifys subscriber event
func (pub *genericSubscriber) NotifyFunc(eventType typedef.EventType, event typedef.Event) {
	pub.HandleEvent(eventType, event)
}

// TopicsFunc returns the topics that the subscriber is interested in
func (pub *genericSubscriber) TopicsFunc() []typedef.EventType {
	return pub.EventTypes()
}
