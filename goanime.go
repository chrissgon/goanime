package goanime

import (
	"github.com/chrissgon/goanime/cmd"
	"github.com/chrissgon/goanime/pkg"
	"github.com/joho/godotenv"
	"github.com/machinebox/progress"
)

func init() {
	godotenv.Load()
}

// generate scraper by provider
func NewScraper(provider cmd.Providers, anime, episode string, dub bool) pkg.Scraper {
	return cmd.NewScraper(provider, anime, episode, dub)
}

// generate scrapers by providers
func NewScrapers(anime, episode string, dub bool) (all []pkg.Scraper) {
	return cmd.NewScrapers(anime, episode, dub)
}

// download anime by especific provider
func DownloadByScraper(scraper pkg.Scraper, status chan progress.Progress) (string, error) {
	return cmd.DownloadByScraper(scraper, status)
}

// download anime by all providers
func DownloadByScrapers(scrapers []pkg.Scraper, status chan progress.Progress) (string, error) {
	return cmd.DownloadByScrapers(scrapers, status)
}
