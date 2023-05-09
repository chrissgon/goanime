package goanime

import (
	"fmt"
	"testing"

	"github.com/machinebox/progress"
)

func TestAnimesGames(t *testing.T) {
	status := make(chan progress.Progress)
	go func() {
		for s := range status {
			fmt.Println(int(s.Percent()))
		}
	}()
	res, err := DownloadByScraper(NewScraper(ANIMESONLINEHD, "one punch man", "12", false), status)

	fmt.Println(res, err)
}
