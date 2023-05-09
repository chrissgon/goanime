package utils

import (
	"fmt"
	"math"
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

func ReplaceAllString(old string, new string, patterns []string) string {
	for _, pattern := range patterns {
		old = regexp.MustCompile(pattern).ReplaceAllString(old, new)
	}
	return old
}

func GetAttrByElements(doc *goquery.Document, selector, attr string) (occurences []string) {
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
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

func StrToInt(str string) (int, error) {
	part := strings.Split(str, ".")[0]
	return strconv.Atoi(part)
}

func GetTitleWithGreatestSimilarity(text string, titles []string) (title string) {
	var greatest float64 = 0

	for _, value := range titles {
		similarity := cosineSimilarity(text, value)

		if similarity > greatest {
			greatest = similarity
			title = value
		}
	}
	return
}

func cosineSimilarity(s1 string, s2 string) float64 {
	set1 := make(map[string]int)
	set2 := make(map[string]int)

	for _, char := range strings.Split(s1, "") {
		set1[char]++
	}

	for _, char := range strings.Split(s2, "") {
		set2[char]++
	}

	numerator := 0
	for key, value := range set1 {
		numerator += value * set2[key]
	}

	sum1 := 0
	for _, value := range set1 {
		sum1 += value * value
	}
	norm1 := math.Sqrt(float64(sum1))

	sum2 := 0
	for _, value := range set2 {
		sum2 += value * value
	}
	norm2 := math.Sqrt(float64(sum2))

	return float64(numerator) / (norm1 * norm2)
}
