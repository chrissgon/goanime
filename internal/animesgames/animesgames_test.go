package animesgames

import (
	"os"
)

func init() {
	os.Setenv("GOANIME_TIMEOUT", "10")
	os.Setenv("GOANIME_FOLDER", "./animes")
}
