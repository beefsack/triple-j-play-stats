package tjps

import (
	"testing"
)

func checkSongParse(text, expectedTitle, expectedArtist string, t *testing.T) {
	s, err := ParseSong(text)
	if err != nil {
		t.Fatal(err.Error())
	}
	if s.Title != expectedTitle {
		t.Fatalf(`Title does not match expected
text: %s
expected: %s
actual: %s`, text, expectedTitle, s.Title)
	}
	if s.Artist != expectedArtist {
		t.Fatalf(`Artist does not match expected
text: %s
expected: %s
actual: %s`, text, expectedArtist, s.Artist)
	}
}

func TestParseSong(t *testing.T) {
	checkSongParse(".@thekooksmusic - Sofa Song [03:47]",
		"Sofa Song",
		".@thekooksmusic",
		t)
}
