package twitter

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/JosephSalisbury/staler/stale"
)

// Following is a Staler for Twitter Followings.
type Following struct {
	expiry time.Duration

	client *anaconda.TwitterApi
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

	client := anaconda.NewTwitterApiWithCredentials(
		accessToken,
		accessTokenSecret,
		consumerKey,
		consumerSecret,
	)

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

	for page := range f.client.GetFriendsListAll(v) {
		if page.Error != nil {
			return nil, page.Error
		}

		for _, user := range page.Friends {
			wg.Add(1)

			go func(user anaconda.User) {
				defer wg.Done()

				v := url.Values{}
				v.Set("user_id", user.IdStr)
				v.Set("count", "1")

				timeline, err := f.client.GetUserTimeline(v)
				if err != nil {
					errChan <- err
				}

				if len(timeline) == 0 {
					item := stale.Item{
						ID:  user.ScreenName,
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
					ID:  user.ScreenName,
					Age: latestTweetTime,
				}
				itemChan <- item
			}(user)
		}
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
	_, err := f.client.UnfollowUser(item.ID)

	return err
}

func (f *Following) String() string {
	return "Twitter followings"
}
