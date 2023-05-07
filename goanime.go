package goanime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/chrissgon/goanime/pkg"
	"github.com/chrissgon/goanime/utils"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/machinebox/progress"
)

type Providers string

const (
	ANIMESONLINEHD Providers = "ANIMESONLINEHD"
)

var factories = map[string]pkg.NewScraperFunction{
	"ANIMESONLINEHD": pkg.NewScraperAnimesOnlineHD,
}

func init() {
	godotenv.Load()
}

// generate scraper by provider
func NewScraper(provider Providers, anime, episode string, dub bool) pkg.Scraper {
	return factories[string(provider)](anime, episode, dub)
}

// generate scrapers by providers
func NewScrapers(anime, episode string, dub bool) (all []pkg.Scraper) {
	for _, fs := range factories {
		all = append(all, fs(anime, episode, dub))
	}
	return
}

// download anime by especific provider
func DownloadByScraper(scraper pkg.Scraper, status chan progress.Progress) (string, error) {
	res, err := scraper.Search()

	if err != nil {
		return "", utils.NewError("DownloadByProvider", err)
	}

	return download(res, status)
}

// download anime by all providers
func DownloadByScrapers(scrapers []pkg.Scraper, status chan progress.Progress) (string, error) {
	timeout := make(chan error)
	go utils.TimeoutRoutine(timeout)

	response := make(chan interface{})
	for _, scraper := range scrapers {
		go asyncSearch(scraper, response)
	}

	select {
	case err := <-timeout:
		return "", utils.NewError("DownloadByProviders", err)

	case value := <-response:
		switch value.(type) {
		case *http.Response:
			return download(value.(*http.Response), status)
		default:
			return "", utils.NewError("DownloadByProviders", value.(error))
		}
	}
}

func asyncSearch(scraper pkg.Scraper, response chan interface{}) {
	url, err := scraper.Search()

	if err != nil {
		response <- err
		return
	}

	response <- url
}

func download(res *http.Response, status chan progress.Progress) (string, error) {
	base := os.Getenv("GOANIME_FOLDER")
	rd := progress.NewReader(res.Body)

	extension, err := utils.GetFileExtensionFromUrl(res.Request.URL.String())

	if err != nil {
		return "", utils.NewError("Download", err)
	}

	lastChar := base[len(base)-1:]
	if lastChar != "/" {
		base = fmt.Sprintf("%s/", base)
	}

	filepath := fmt.Sprintf("%s%s.%s", base, uuid.New().String(), extension)
	out, err := os.Create(filepath)

	if err != nil {
		return "", utils.NewError("Download", err)
	}

	go func() {
		p := progress.NewTicker(context.Background(), rd, res.ContentLength, 1*time.Second)

		for v := range p {
			status <- v
		}
	}()

	_, err = io.Copy(out, rd)

	return filepath, err
}
