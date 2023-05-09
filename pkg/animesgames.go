package pkg

import (
	"net/http"

	"github.com/chrissgon/goanime/internal/animesgames"
)

type animesGames anime

// search anime
func (a *animesGames) Search() (*http.Response, error) {
	return animesgames.Search(a.anime, a.episode, a.dub)
}

func NewScraperAnimesGames(anime, episode string, dub bool) Scraper {
	return &animesGames{
		anime:   anime,
		episode: episode,
		dub:     dub,
	}
}
