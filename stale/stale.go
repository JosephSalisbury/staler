package stale

import (
	"time"
)

type Staler interface {
	Expiry() time.Duration
	List() ([]Item, error)
	Delete(Item) error
}

type Item struct {
	ID  string
	Age time.Time
}

func (i Item) String() string {
	return i.ID
}
