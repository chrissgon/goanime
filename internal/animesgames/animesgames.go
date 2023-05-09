package animesgames

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/chrissgon/goanime/utils"
)

type anime struct {
	URL string `json:"url"`
}

type animesResponse map[string]anime

var REQUEST_HEADERS = map[string]string{
	"Referer": "https://animesgames.net/",
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

	url := utils.GetTitleWithGreatestSimilarity(anime, urls)
	fmt.Println(url)
	html, err := getReleasedEpisodeHTML([]string{url}, episode)

	if err != nil {
		return nil, internalError(err)
	}

	return getVideoRequest(html)

	// if err != nil {
	// 	return nil, internalError(err)
	// }

	// _, err = utils.NewRequest(videoURL, http.MethodGet, nil, REQUEST_HEADERS)

	// if err != nil {
	// 	return nil, internalError(err)
	// }

	// fmt.Println(videoURL)

	// mp4URL = res.Request.Response.Header.Get("Location")

	// return utils.NewRequest(videoURL, http.MethodGet, nil, nil)
}

func getAnimesEpisodesURL(anime string) (urls []string, err error) {
	query := strings.Join(strings.Split(strings.TrimSpace(anime), " "), "+")
	url := fmt.Sprintf("https://animesgames.net/wp-json/animesonline/search/?nonce=1e8d73f99e&keyword=%s", strings.ToLower(query))
	res, err := utils.NewRequest(url, http.MethodGet, nil, nil)

	if err != nil {
		return nil, utils.NewError("getAnimesID", err)
	}

	animesResponse := animesResponse{}
	json.NewDecoder(res.Body).Decode(&animesResponse)

	for _, animeResponse := range animesResponse {
		urls = append(urls, animeResponse.URL)
	}

	return
}

func getReleasedEpisodeHTML(urls []string, episode string) (string, error) {
	timeout := make(chan error)
	go utils.TimeoutRoutine(timeout)

	response := make(chan string)
	for _, url := range urls {
		go getPageOK(url, episode, response)
	}

	select {
	case html := <-response:
		return html, nil
	case err := <-timeout:
		return "", err
	}
}

func getPageOK(url, episode string, response chan string) {
	res, err := http.Get(url)

	if err != nil {
		return
	}

	str, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	html := string(str)

	regexMatchEpisodesContainer := regexp.MustCompile(`<a href="https://animesgames.net/video/[^>]*>.*?</a>`)

	episodesStrHTML := regexMatchEpisodesContainer.FindAllString(html, -1)

	start, end := getRangePossibleEpisodes(episode)

	if len(episodesStrHTML) < end {
		return
	}

	url = getEpisodePageURL(episode, episodesStrHTML[start:end])

	res, err = http.Get(url)

	if err == nil && res.StatusCode == http.StatusOK {
		str, _ := ioutil.ReadAll(res.Body)
		response <- string(str)
	}
}

func getEpisodePageURL(episode string, episodesStrHTML []string) (url string) {
	for _, episodeStrHTML := range episodesStrHTML {
		pattern := fmt.Sprintf("Episódio %s ", episode)
		isTheEpisode := regexp.MustCompile(pattern).MatchString(episodeStrHTML)

		if isTheEpisode {
			url = regexp.MustCompile("https://animesgames.net/video/[0-9-]*").FindString(episodeStrHTML)
			break
		}
	}

	return
}

func getRangePossibleEpisodes(episode string) (int, int) {
	intEpisode, _ := utils.StrToInt(episode)

	if intEpisode <= 3 {
		return 0, 3
	}

	return intEpisode - 3, intEpisode
}

func getVideoRequest(html string) (*http.Response, error) {
	url := regexp.MustCompile(`https://animesgames.net/player.*?"`).FindString(html)

	if url == "" {
		return nil, utils.ERROR_NOT_FOUND
	}

	url = url[:len(url)-1]
	res, err := http.Get(url)
	bytesHTML, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	videoURL := regexp.MustCompile("(?m)https:.*?.(?i)(.mp4)").FindString(string(bytesHTML))
	videoURL = utils.ReplaceAllString(videoURL, "", []string{`\\`})
	res, err = utils.NewRequest(videoURL, http.MethodGet, nil, REQUEST_HEADERS)

	if err != nil {
		return nil, err
	}

	contentType := res.Header.Get("Content-type")

	if contentType == "video/mp4" {
		return utils.NewRequest(videoURL, http.MethodGet, nil, REQUEST_HEADERS)
	}

	// return nil, utils.ERROR_NOT_FOUND

	// fmt.Println(videoURL)
	// return nil, nil
	videoURL = res.Request.Response.Header.Get("Location")
	res, err = utils.NewRequest(videoURL, http.MethodGet, nil, REQUEST_HEADERS)

	if err != nil {
		return nil, utils.ERROR_NOT_FOUND
	}

	bytesHTML, err = ioutil.ReadAll(res.Body)
	videoURL = regexp.MustCompile(`(?m)play_url.*?.(?i)(.mp4).*?"`).FindString(string(bytesHTML))
	videoURL = videoURL[11 : len(videoURL)-1]

	return utils.NewRequest(videoURL, http.MethodGet, nil, nil)
}

func internalError(err error) error {
	return utils.NewError("ANIMESGAMES", err)
}
