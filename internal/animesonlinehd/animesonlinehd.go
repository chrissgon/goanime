package animesonlinehd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/chrissgon/goanime/utils"
)

type animeResponse [][]string

var REPLACE_PATTERNS = []string{
	"https://animesonlinehd.vip/",
	"/",
}

var REQUEST_HEADERS = map[string]string{
	"Referer": "https://animesonlinehd.vip/",
}

func internalError(err error) error {
	return utils.NewError("ANIMESONLINEHD", err)
}

func Search(anime, episode string, dub bool) (*http.Response, error) {
	animesID, err := getAnimesID(anime)

	if err != nil {
		return nil, internalError(err)
	}

	if dub {
		animesID = utils.FilterByMatchPattern(animesID, utils.DUB_REGEX)
	} else {
		animesID = utils.FilterByNotMatchPattern(animesID, utils.DUB_REGEX)

	}

	html, err := getReleasedEpisodeHTML(animesID, episode)

	if err != nil {
		return nil, internalError(err)
	}

	mp4URL, err := getVideoURL(html)

	if err != nil {
		return nil, internalError(err)
	}

	res, err := utils.NewRequest(mp4URL, http.MethodGet, nil, REQUEST_HEADERS)

	if err != nil {
		return nil, internalError(err)
	}

	mp4URL = res.Request.Response.Header.Get("Location")

	return utils.NewRequest(mp4URL, http.MethodGet, url.Values{}, REQUEST_HEADERS)
}

func getAnimesID(anime string) ([]string, error) {
	query := strings.Join(strings.Split(anime, " "), "+")
	url := fmt.Sprintf("https://animesonlinehd.vip/?s=%s", strings.ToLower(query))

	searchPageDocument, err := utils.GetPageDocument(http.Get(url))

	if err != nil {
		return nil, utils.NewError("getAnimesID", err)
	}

	duplicateUrls := utils.GetAttrByElements(searchPageDocument, "a[itemprop=URL]", "href")
	urls := utils.RemoveDuplicateStrings(duplicateUrls)

	animesID := []string{}

	for _, url := range urls {
		animeID := utils.ReplaceAllString(url, "", REPLACE_PATTERNS)
		animesID = append(animesID, animeID)
	}

	return animesID, nil
}

func getReleasedEpisodeHTML(animesID []string, episode string) (string, error) {
	timeout := make(chan error)
	go utils.TimeoutRoutine(timeout)

	response := make(chan string)
	for _, animeID := range animesID {
		go getPageOK(animeID, episode, response)
	}

	select {
	case html := <-response:
		return html, nil
	case err := <-timeout:
		return "", err
	}
}

func getPageOK(animeID, episode string, response chan string) {
	url := fmt.Sprintf("https://animesonlinehd.vip/episodio/%s-episodio-%s", animeID, episode)
	res, err := http.Get(url)

	if err == nil && res.StatusCode == http.StatusOK {
		str, _ := ioutil.ReadAll(res.Body)
		response <- string(str)
	}
}

func getVideoURL(html string) (string, error) {
	url := regexp.MustCompile("https.*.mp4").FindString(html)

	if url == "" {
		return url, utils.ERROR_NOT_FOUND
	}

	return url, nil
}
