package pkg

import (
	"net/http"

	"github.com/chrissgon/goanime/internal/animesonlinehd"
)

type animesOnlineHD anime

// search anime
func (a *animesOnlineHD) Search() (*http.Response, error) {
	return animesonlinehd.Search(a.anime, a.episode, a.dub)
}

func NewScraperAnimesOnlineHD(anime, episode string, dub bool) Scraper {
	return &animesOnlineHD{
		anime:   anime,
		episode: episode,
		dub:     dub,
	}
}
