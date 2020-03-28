package transmission

import (
	"errors"
	"strconv"
	"time"

	"github.com/hekmon/transmissionrpc"

	"github.com/JosephSalisbury/staler/stale"
)

// Torrent is a Staler for finished Transmission torrents.
type Torrent struct {
	expiry time.Duration

	client *transmissionrpc.Client
}

// NewTransmission returns a new Staler for finished Transmission torrents.
func NewTorrent(host string, user string, password string, expiry time.Duration) (*Torrent, error) {
	if host == "" {
		return nil, errors.New("transmission host cannot be empty")
	}

	client, err := transmissionrpc.New(host, user, password, &transmissionrpc.AdvancedConfig{})
	if err != nil {
		return nil, err
	}

	torrent := &Torrent{
		expiry: expiry,

		client: client,
	}

	return torrent, nil
}

// Expiry returns the maximum age of finished Transmission torrents.
func (t *Torrent) Expiry() time.Duration {
	return t.expiry
}

// List returns all finished Transmission torrents.
func (t *Torrent) List() ([]stale.Item, error) {
	torrents, err := t.client.TorrentGetAll()
	if err != nil {
		return nil, err
	}

	items := []stale.Item{}
	for _, torrent := range torrents {
		if !*torrent.IsFinished {
			continue
		}

		item := stale.Item{
			ID:  strconv.FormatInt(*torrent.ID, 10),
			Age: *torrent.ActivityDate,
		}

		items = append(items, item)
	}

	return items, nil
}

// Delete deletes the specified Transmission torrent.
func (t *Torrent) Delete(item stale.Item) error {
	id, err := strconv.ParseInt(item.ID, 10, 64)
	if err != nil {
		return err
	}

	return t.client.TorrentRemove(
		&transmissionrpc.TorrentRemovePayload{
			IDs:             []int64{id},
			DeleteLocalData: true,
		},
	)
}

func (t *Torrent) String() string {
	return "Transmission torrents"
}
