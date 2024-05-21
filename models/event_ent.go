package models

import "github.com/nkust-monitor-iot-project-2024/central/ent"

type eventRepositoryEnt struct {
	client *ent.Client
}

func NewEventRepositoryEnt(client *ent.Client) EventRepository {
	return &eventRepositoryEnt{client: client}
}

func (r *eventRepositoryEnt) GetEvent(id string) (*Event, error) {
	r.client.Event.Get()
	panic("not implemented")
}
