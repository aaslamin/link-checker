package linkworker

import (
	"errors"
	"fmt"
	"net/url"
)

func toAbsolute(link, base string) string {
	uri, err := url.Parse(link)
	if err != nil {
		return ""
	}

	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}

	return baseUrl.ResolveReference(uri).String()
}

func validateURL(input string) (*url.URL, error) {
	if u, err := url.Parse(input); err != nil {
		return nil, errors.New(fmt.Sprintf("\"%s\" is not a valid url", input))
	} else if u.Scheme == "" || u.Host == "" {
		return nil, errors.New(fmt.Sprintf("\"%s\" must be an absolute url", input))
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New(fmt.Sprintf("\"%s\" must use http or https", input))
	} else {
		return u, nil
	}
}
