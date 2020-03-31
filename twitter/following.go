package twitter

import (
	"errors"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/JosephSalisbury/staler/stale"
)

// Following is a Staler for Twitter Followings.
type Following struct {
	expiry time.Duration

	client *twitter.Client
}

// NewFollowing returns a new Staler for Twitter Followings.
func NewFollowing(accessToken string, accessTokenSecret string, consumerKey string, consumerSecret string, expiry time.Duration) (*Following, error) {
	if accessToken == "" {
		return nil, errors.New("twitter access token cannot be empty")
	}
	if accessTokenSecret == "" {
		return nil, errors.New("twitter access token secret cannot be empty")
	}
	if consumerKey == "" {
		return nil, errors.New("twitter consumer key cannot be empty")
	}
	if consumerSecret == "" {
		return nil, errors.New("twitter consumer secret cannot be empty")
	}

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	following := &Following{
		expiry: expiry,

		client: client,
	}

	return following, nil
}

// Expiry returns the maximum age of Twitter Followings.
func (f *Following) Expiry() time.Duration {
	return f.expiry
}

// List returns all Twitter Followings.
func (f *Following) List() ([]stale.Item, error) {
	var wg sync.WaitGroup

	itemChan := make(chan stale.Item)
	errChan := make(chan error)
	quitChan := make(chan struct{})

	v := url.Values{}
	v.Set("count", "200")

	friends, _, err := f.client.Friends.IDs(nil)
	if err != nil {
		return nil, err
	}

	for _, userId := range friends.IDs {
		wg.Add(1)

		go func(id int64) {
			defer wg.Done()

			timeline, _, err := f.client.Timelines.UserTimeline(&twitter.UserTimelineParams{
				UserID: userId,
				Count:  1,
			})
			if err != nil {
				errChan <- err
			}

			if len(timeline) == 0 {
				item := stale.Item{
					ID:  strconv.FormatInt(userId, 10),
					Age: time.Time{},
				}
				itemChan <- item

				return
			}

			latestTweetTime, err := timeline[0].CreatedAtTime()
			if err != nil {
				errChan <- err
			}

			item := stale.Item{
				ID:  strconv.FormatInt(userId, 10),
				Age: latestTweetTime,
			}
			itemChan <- item
		}(userId)
	}

	go func() {
		wg.Wait()
		close(quitChan)
	}()

	items := []stale.Item{}

	waiting := true
	for waiting {
		select {
		case item := <-itemChan:
			items = append(items, item)
		case err := <-errChan:
			return nil, err
		case <-quitChan:
			waiting = false
		}
	}

	return items, nil
}

// Delete deletes the specific Twitter Following.
func (f *Following) Delete(item stale.Item) error {
	id, err := strconv.ParseInt(item.ID, 10, 64)
	if err != nil {
		return err
	}

	if _, _, err := f.client.Friendships.Destroy(&twitter.FriendshipDestroyParams{
		UserID: id,
	}); err != nil {
		return err
	}

	return nil
}

func (f *Following) String() string {
	return "Twitter followings"
}
