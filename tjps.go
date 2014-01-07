package tjps

import (
	"encoding/json"
	"fmt"
	"github.com/mrjones/oauth"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

const (
	TWITTER_TIME_FORMAT  = time.RubyDate
	TWITTER_RATE_WINDOW  = 900
	TWITTER_RATE_AMOUNT  = 180
	TWITTER_REQUEST_WAIT = TWITTER_RATE_WINDOW / TWITTER_RATE_AMOUNT
)

var songRegexp = regexp.MustCompile(`^\s*(.+) - (.+?)\s+\[\d+:\d+\]\s*$`)

type searchResult struct {
	Statuses []tweet `json:"statuses"`
}

type tweet struct {
	Id        int       `json:"id"`
	CreatedAt string    `json:"created_at"`
	Text      string    `json:"text"`
	User      tweetUser `json:"user"`
}

type tweetUser struct {
	ScreenName string `json:"screen_name"`
}

type Song struct {
	Title, Artist string
}

type Play struct {
	Song Song
	At   time.Time
}

type PlayCount struct {
	Song  Song
	Count int
}

type PlayCountSorter struct {
	PlayCounts []PlayCount
}

func GetPlays(c *oauth.Consumer, token *oauth.AccessToken,
	from, until time.Time) (plays []Play,
	err error) {
	maxIdStr := ""
	maxId := 0
	untilStr := until.Format("2006-01-02")
	lastCall := int64(0)
	searching := true
	for searching {
		time.Sleep(time.Duration((TWITTER_REQUEST_WAIT-
			(time.Now().Unix()-lastCall))+1) * time.Second)
		fmt.Fprintf(os.Stderr, "Fetching tweets up to id %s\n", maxIdStr)
		params := map[string]string{
			"q":           "from:triplejplays",
			"result_type": "recent",
			"count":       "100",
		}
		if maxId != 0 {
			params["max_id"] = maxIdStr
		} else {
			params["until"] = untilStr
		}
		response, err := c.Get(
			"https://api.twitter.com/1.1/search/tweets.json", params, token)
		if err != nil {
			return plays, err
		}
		lastCall = time.Now().Unix()
		res := searchResult{}
		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&res)
		if err != nil {
			response.Body.Close()
			return plays, err
		}
		if len(res.Statuses) == 0 {
			break
		}
		lastStatusAt := ""
		for _, t := range res.Statuses {
			lastStatusAt = t.CreatedAt
			if maxId == 0 || maxId > t.Id-1 {
				maxId = t.Id
				maxIdStr = strconv.Itoa(t.Id - 1)
			}
			if t.User.ScreenName != "triplejplays" {
				continue
			}
			p, err := TweetToPlay(t)
			if err != nil {
				continue
			}
			if p.At.Before(from) {
				searching = false
				break
			}

			plays = append(plays, p)
		}
		fmt.Fprintf(os.Stderr, "Last tweet fetched dated %s\n", lastStatusAt)
		response.Body.Close()
	}
	return
}

func ParseSong(s string) (song Song, err error) {
	matches := songRegexp.FindStringSubmatch(s)
	if matches == nil {
		err = fmt.Errorf("Could not extract song from %s", s)
		return
	}
	song.Artist = matches[1]
	song.Title = matches[2]
	return
}

func TweetToPlay(t tweet) (play Play, err error) {
	s, err := ParseSong(t.Text)
	if err != nil {
		return
	}
	play.Song = s
	at, err := time.Parse(TWITTER_TIME_FORMAT, t.CreatedAt)
	if err != nil {
		return
	}
	play.At = at
	return
}

func PlaysByCount(plays []Play) (pc []PlayCount) {
	titleMap := map[string]*PlayCount{}
	for _, p := range plays {
		c := titleMap[p.Song.Title]
		if c == nil {
			c = &PlayCount{
				Song: p.Song,
			}
			titleMap[p.Song.Title] = c
		}
		c.Count++
	}
	for _, c := range titleMap {
		pc = append(pc, *c)
	}
	return
}

func SortPlayCounts(playCounts []PlayCount) []PlayCount {
	sorter := PlayCountSorter{playCounts}
	sort.Sort(sort.Reverse(&sorter))
	return sorter.PlayCounts
}

func (p *PlayCountSorter) Len() int {
	return len(p.PlayCounts)
}

func (p *PlayCountSorter) Swap(i, j int) {
	p.PlayCounts[i], p.PlayCounts[j] = p.PlayCounts[j], p.PlayCounts[i]
}

func (p *PlayCountSorter) Less(i, j int) bool {
	return p.PlayCounts[i].Count < p.PlayCounts[j].Count
}
