package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/JosephSalisbury/staler/docker"
	"github.com/JosephSalisbury/staler/stale"
	"github.com/JosephSalisbury/staler/transmission"
	"github.com/JosephSalisbury/staler/twitter"
)

const (
	month = time.Hour * 24 * 30
)

var (
	rootCmd = &cobra.Command{
		Use:   "staler",
		Short: "staler cleans up stale things",
		Run:   run,
	}
)

var (
	dockerContainerExpiry time.Duration
	dockerImageExpiry     time.Duration

	transmissionHost          string
	transmissionUser          string
	transmissionPassword      string
	transmissionTorrentExpiry time.Duration

	twitterAccessToken       string
	twitterAccessTokenSecret string
	twitterConsumerKey       string
	twitterConsumerSecret    string
	twitterFollowingExpiry   time.Duration
)

func init() {
	rootCmd.Flags().DurationVar(&dockerContainerExpiry, "docker-container-expiry", month, "duration for exited Docker containers until they become stale")
	rootCmd.Flags().DurationVar(&dockerImageExpiry, "docker-image-expiry", month, "duration for Docker images until they become stale")

	rootCmd.Flags().StringVar(&transmissionHost, "transmission-host", "", "host for Transmission server")
	rootCmd.Flags().StringVar(&transmissionUser, "transmission-user", "", "username for Transmission server")
	rootCmd.Flags().StringVar(&transmissionPassword, "transmission-password", "", "password for Transmission server")
	rootCmd.Flags().DurationVar(&transmissionTorrentExpiry, "transmission-torrent-expiry", month, "duration for finished Transmission torrents until they become stale")

	rootCmd.Flags().StringVar(&twitterAccessToken, "twitter-access-token", "", "access token for Twitter")
	rootCmd.Flags().StringVar(&twitterAccessTokenSecret, "twitter-access-token-secret", "", "access token secret for Twitter")
	rootCmd.Flags().StringVar(&twitterConsumerKey, "twitter-consumer-key", "", "consumer key for Twitter")
	rootCmd.Flags().StringVar(&twitterConsumerSecret, "twitter-consumer-secret", "", "consumer secret for Twitter")
	rootCmd.Flags().DurationVar(&twitterFollowingExpiry, "twitter-following-expiry", 6*month, "duration for Twitter Followings with no new Tweets until they become stale")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run(cmd *cobra.Command, args []string) {
	dockerContainer, err := docker.NewContainer(dockerContainerExpiry)
	if err != nil {
		log.Fatalf("%v", err)
	}

	dockerImage, err := docker.NewImage(dockerImageExpiry)
	if err != nil {
		log.Fatalf("%v", err)
	}

	transmissionTorrent, err := transmission.NewTorrent(
		transmissionHost,
		transmissionUser,
		transmissionPassword,
		transmissionTorrentExpiry,
	)
	if err != nil {
		log.Fatalf("%v", err)
	}

	twitterFollowing, err := twitter.NewFollowing(
		twitterAccessToken,
		twitterAccessTokenSecret,
		twitterConsumerKey,
		twitterConsumerSecret,
		twitterFollowingExpiry,
	)
	if err != nil {
		log.Fatalf("%v", err)
	}

	stalers := []stale.Staler{
		dockerContainer,
		dockerImage,
		transmissionTorrent,
		twitterFollowing,
	}

	stale.RemoveStale(stalers)
}
