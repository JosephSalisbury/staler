package stale

import (
	"log"
	"sync"
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

// RemoveStale takes a list of Stalers, and deletes all stale items.
func RemoveStale(stalers []Staler) {
	log.Printf("checking for and removing stale items")

	var wg sync.WaitGroup

	for _, staler := range stalers {
		wg.Add(1)

		go func(staler Staler) {
			defer wg.Done()

			items, err := staler.List()
			if err != nil {
				log.Printf("could not list items for %v: %v", staler, err)
				return
			}

			for _, item := range items {
				if time.Since(item.Age) > staler.Expiry() {
					log.Printf("deleting stale %v item with id '%v'", staler, item)

					if err := staler.Delete(item); err != nil {
						log.Printf("could not delete stale %v item with id '%v': %v", staler, item, err)
					}
				}

			}
		}(staler)
	}

	wg.Wait()

	log.Printf("checked for and removed any stale items")
}
