package pkg

import (
	"net/http"
)

type Scraper interface {
	Search() (*http.Response, error)
}

type NewScraperFunction func(anime, episode string, dub bool) Scraper

type anime struct {
	anime   string
	episode string
	dub     bool
}
