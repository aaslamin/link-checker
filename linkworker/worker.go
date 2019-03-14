package linkworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"golang.org/x/net/html"
)

const (
	DefaultTimeoutValue = 5
	MaximumTimeoutValue = 20
)

type Worker struct {
	ID          string `json:"worker_id"`
	URL         string `json:"url"`
	LinkTimeout int    `json:"link_timeout"`

	httpClient  *http.Client
	urlBasePath string
}

type LinkError struct {
	URL     string `json:"url"`
	Status  int    `json:"http_status,omitempty"`
	Timeout bool   `json:"timeout,omitempty"`
}

func (w *Worker) ProcessURL(storage WorkerResultStorage) {
	linkErr, resp := w.verifyURL(w.URL)
	if linkErr != nil {
		storage.AddLinkError(context.Background(), w.ID, *linkErr)
		return
	}
	defer resp.Body.Close()

	var links []string
	var wg sync.WaitGroup
	tokenizer := html.NewTokenizer(resp.Body)

LOOP:
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			break LOOP
		case html.StartTagToken, html.EndTagToken:
			token := tokenizer.Token()
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					// sanitize parsed urls:
					//   - convert relative paths to absolute
					//   - a url is considered valid iff:
					//     - it has both a scheme and host
					//     - the scheme is either http or https
					if link, err := validateURL(toAbsolute(attr.Val, w.urlBasePath)); err == nil {
						links = append(links, link.String())
					}
				}
			}
		}
	}

	// using a wait group to create a counting semaphore in order to determine when all link worker routines
	// have finished - the link worker goroutines will decrement the waitgroup once they are donezo.
	wg.Add(len(links))
	// make a buffered channel of the length of the number of links because this is how many
	// results to get back
	resultPipe := make(chan LinkError, len(links))

	// start delegating work - spawn a worker for each link that we have scraped!
	// once all work has been delegated, close the result channel.
	// this is also done in a goroutine so we can start processing the results right away as they become ready
	go func() {
		for _, link := range links {
			go func(link string, wg *sync.WaitGroup, resultCh chan<- LinkError) {
				defer wg.Done()
				if linkErr, _ := w.verifyURL(link); linkErr != nil {
					resultCh <- *linkErr
				}
			}(link, &wg, resultPipe)
		}

		wg.Wait()
		close(resultPipe)
	}()

	for result := range resultPipe {
		storage.AddLinkError(context.Background(), w.ID, result)
	}
}

func (w *Worker) verifyURL(destination string) (*LinkError, *http.Response) {
	result := &LinkError{
		URL: destination,
	}

	resp, err := w.httpClient.Get(destination)
	if err != nil {
		if err, ok := err.(*url.Error); ok && err.Timeout() {
			result.Timeout = true
		}

		return result, resp
	}

	code := resp.StatusCode
	switch {
	case code >= 200 && code < 300:
		return nil, resp
	default:
		result.Status = code
	}

	return result, resp
}

func (w *Worker) ValidateWorkerRequest(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(w); err != nil {
		return errors.New(fmt.Sprintf("failed to parse JSON input - error: %s", err))
	}

	if w.LinkTimeout <= 0 || w.LinkTimeout > MaximumTimeoutValue {
		w.LinkTimeout = DefaultTimeoutValue
	}

	destinationURL, err := validateURL(w.URL)
	if err != nil {
		return err
	}

	w.ID = uuid.New()
	w.urlBasePath = fmt.Sprintf("%s://%s", destinationURL.Scheme, destinationURL.Host)
	w.httpClient = &http.Client{
		Timeout: time.Second * time.Duration(w.LinkTimeout),
	}

	return nil
}
