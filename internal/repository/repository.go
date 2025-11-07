package repository

import "context"

type SettingRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

type VolumeRepository interface {
	Add(ctx context.Context, volume Volume) error
	Get(ctx context.Context, id string) (Volume, error)
	List(ctx context.Context) ([]Volume, error)
}

type TaskRepository interface {
	Add(ctx context.Context, task Task) error
	Get(ctx context.Context, id string) (Task, error)
	List(ctx context.Context) ([]Task, error)
}

type EventRepository interface {
	Add(ctx context.Context, event Event) error
	Get(ctx context.Context, id string) (Event, error)
	List(ctx context.Context) ([]Event, error)
}

type ConnectionRepository interface {
	Add(ctx context.Context, connection Connection) error
	Get(ctx context.Context, id string) (Connection, error)
	List(ctx context.Context) ([]Connection, error)
}

type FilterRepository interface {
	Add(ctx context.Context, filter Filter) error
	Get(ctx context.Context, id string) (Filter, error)
	List(ctx context.Context) ([]Filter, error)
}

type SourceRepository interface {
	Add(ctx context.Context, source Source) error
	Get(ctx context.Context, id string) (Source, error)
	List(ctx context.Context) ([]Source, error)
}

type Volume interface{}

type Filter interface{}

type Source interface{}

type Connection interface{}

type Event interface{}

type Task interface{}
