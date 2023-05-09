package animesonlinehd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

func Search(anime, episode string, dub bool) (*http.Response, error) {
	urls, err := getAnimesEpisodesURL(anime)

	if err != nil {
		return nil, internalError(err)
	}

	if dub {
		urls = utils.FilterByMatchPattern(urls, utils.DUB_REGEX)
	} else {
		urls = utils.FilterByNotMatchPattern(urls, utils.DUB_REGEX)
	}

	urlPattern := fmt.Sprintf("https://animesonlinehd.vip/%s", utils.ReplaceAllString(anime, "-", []string{"-"}))
	url := utils.GetTitleWithGreatestSimilarity(urlPattern, urls)

	html, err := getReleasedEpisodeHTML(url, episode)

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

	return utils.NewRequest(mp4URL, http.MethodGet, nil, REQUEST_HEADERS)
}

func getAnimesEpisodesURL(anime string) (urls []string, err error) {
	query := strings.Join(strings.Split(anime, " "), "+")
	url := fmt.Sprintf("https://animesonlinehd.vip/?s=%s", strings.ToLower(query))

	searchPageDocument, err := goquery.NewDocument(url)

	if err != nil {
		return nil, utils.NewError("getAnimesID", err)
	}

	duplicateUrls := utils.GetAttrByElements(searchPageDocument, "a[itemprop=URL]", "href")
	urls = utils.RemoveDuplicateStrings(duplicateUrls)

	if len(urls) == 0 {
		return nil, internalError(utils.ERROR_NOT_FOUND)
	}

	return
}

func getReleasedEpisodeHTML(url, episode string) (string, error) {
	res, err := http.Get(url)

	if err != nil {
		return "", utils.NewError("getReleasedEpisodeHTML", err)
	}

	defer res.Body.Close()

	str, _ := ioutil.ReadAll(res.Body)
	episodesHTML := string(str)

	regexPattern := fmt.Sprintf("https://animesonlinehd.vip/episodio/.*?.%s.*?/", episode)
	regexMatchEpisodeURL := regexp.MustCompile(regexPattern)

	url = regexMatchEpisodeURL.FindString(episodesHTML)
	res, err = http.Get(url)

	if err != nil {
		return "", utils.NewError("getReleasedEpisodeHTML", err)
	}

	defer res.Body.Close()

	str, _ = ioutil.ReadAll(res.Body)

	return string(str), nil
}

func getVideoURL(html string) (string, error) {
	url := regexp.MustCompile("https.*.mp4").FindString(html)

	if url == "" {
		return url, utils.ERROR_NOT_FOUND
	}

	return url, nil
}

func internalError(err error) error {
	return utils.NewError("ANIMESONLINEHD", err)
}
