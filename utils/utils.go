package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"errors"

	"github.com/PuerkitoBio/goquery"
)

var ERROR_NOT_FOUND = errors.New("not found")
var ERROR_TIMEOUT = errors.New("timeout")
var DUB_REGEX = "dublado|portugues"

func NewError(fn string, err error) error {
	return fmt.Errorf("%s: %w", fn, err)
}

func NewRequest(url, method string, params url.Values, headers map[string]string) (*http.Response, error) {
	encoded := params.Encode()

	req, _ := http.NewRequest(method, url, strings.NewReader(encoded))

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := http.Client{}

	return client.Do(req)
}

func GetOcurrencesByPattern(text, pattern string) []string {
	regex := regexp.MustCompile(pattern)
	return regex.FindAllString(text, -1)
}

func FilterByMatchPattern(texts []string, pattern string) (filtered []string) {
	for _, text := range texts {
		if regexp.MustCompile(pattern).MatchString(text) {
			filtered = append(filtered, text)
		}
	}
	return
}
func FilterByNotMatchPattern(texts []string, pattern string) (filtered []string) {
	for _, text := range texts {
		if !regexp.MustCompile(pattern).MatchString(text) {
			filtered = append(filtered, text)
		}
	}
	return
}

func ReplaceAllString(text string, new string, patterns []string) string {
	for _, pattern := range patterns {
		text = regexp.MustCompile(pattern).ReplaceAllString(text, new)
	}
	return text
}

func GetPageDocument(res *http.Response, err error) (*goquery.Document, error) {
	if err != nil {
		return nil, NewError("GetPageDocument", err)
	}
	defer res.Body.Close()
	return goquery.NewDocumentFromReader(res.Body)
}

func GetAttrByElements(doc *goquery.Document, selector, attr string) (occurences []string) {
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		value, exists := s.Attr(attr)

		if exists {
			occurences = append(occurences, value)
		}
	})
	return
}

func RemoveDuplicateStrings(slice []string) (list []string) {
	appended := make(map[string]bool)
	for _, item := range slice {
		if appended[item] {
			continue
		}
		appended[item] = true
		list = append(list, item)
	}
	return
}

func GetFileExtensionFromUrl(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", NewError("GetFileExtensionFromUrl", err)
	}
	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", NewError("GetFileExtensionFromUrl", err)
	}
	return u.Path[pos+1 : len(u.Path)], nil
}

func TimeoutRoutine(timeout chan error) {
	seconds, _ := strconv.Atoi(os.Getenv("GOANIME_TIMEOUT"))
	time.Sleep(time.Duration(seconds) * time.Second)
	timeout <- ERROR_TIMEOUT
}
