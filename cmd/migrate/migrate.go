package migrate

import "net/url"

func cloneURLWithQuery(u *url.URL, newQuery url.Values) *url.URL {
	clone := *u
	query := clone.Query()
	for key, values := range newQuery {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	clone.RawQuery = query.Encode()
	return &clone
}

var supportedDrivers = map[string]struct{}{
	"postgres":   {},
	"postgresql": {},
}
