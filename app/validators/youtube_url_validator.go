package validators

import (
	"fmt"
	"strings"
)

func YoutubeURLValidator(url string) bool {
	if url == "" {
		fmt.Printf("Invalid URL: %s", url)
		return false
	}

	var domain = "youtube"

	return strings.Contains(domain, url)
}
