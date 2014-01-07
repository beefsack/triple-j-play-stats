package main

import (
	"fmt"
	"github.com/beefsack/triple-j-play-stats"
	"github.com/codegangsta/cli"
	"github.com/mrjones/oauth"
	"os"
	"time"
)

const (
	TIME_FORMAT = "2006-01-02"
)

func main() {
	app := cli.NewApp()
	app.Name = "tjps"
	app.Usage = "Triple J Play Stats - play statistics from the Triple J feed"
	app.Flags = []cli.Flag{
		cli.StringFlag{"key", "", "twitter consumer key"},
		cli.StringFlag{"secret", "", "twitter consumer secret"},
		cli.StringFlag{"from", "", "when to fetch plays from"},
		cli.StringFlag{"until", "", "when to fetch plays until"},
	}
	app.Action = func(c *cli.Context) {
		key := c.String("key")
		if key == "" {
			fmt.Fprintln(os.Stderr,
				"You must specify a key, please run help for more info")
			os.Exit(1)
		}
		secret := c.String("secret")
		if secret == "" {
			fmt.Fprintln(os.Stderr,
				"You must specify a secret, please run help for more info")
			os.Exit(1)
		}
		from, err := time.Parse(TIME_FORMAT, c.String("from"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing from: %s\n", err.Error())
			os.Exit(1)
		}
		until, err := time.Parse(TIME_FORMAT, c.String("until"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing until: %s\n", err.Error())
			os.Exit(1)
		}
		consumer := oauth.NewConsumer(key, secret, oauth.ServiceProvider{
			RequestTokenUrl:   "http://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})
		requestToken, url, err := consumer.GetRequestTokenAndUrl("oob")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(3)
		}
		fmt.Fprintf(os.Stderr,
			"Go to %s, grant access, and enter the verification code: ", url)
		verificationCode := ""
		fmt.Scanln(&verificationCode)
		accessToken, err := consumer.AuthorizeToken(requestToken,
			verificationCode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(4)
		}
		plays, err := tjps.GetPlays(consumer, accessToken, from, until)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching plays: %s\n", err.Error())
		}
		playsByCount := tjps.PlaysByCount(plays)
		sorted := tjps.SortPlayCounts(playsByCount)
		fmt.Println("Count\tTitle\tArtist")
		for _, p := range sorted {
			fmt.Printf("%d\t%s\t%s\n", p.Count, p.Song.Title, p.Song.Artist)
		}
	}
	app.Run(os.Args)
}
