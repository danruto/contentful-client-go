package contenfulclientgo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	ErrUrlNotSet   = errors.New("The client url is not set by either the env var CONTENTFUL_URL or the builder method WithUrl")
	ErrTokenNotSet = errors.New("The client token is not set by either the env var CONTENTFUL_TOKEN or the builder method WithToken")
)

type ContentfulClient struct {
	url, token string
}

// Creates a new instance of a `ContentfulClient`
//
// It can load the opts from env vars named
//
//	CONTENTFUL_URL   - for the contentful base url
//	CONTENTFUL_TOKEN - for the published api token
//
// otherwise you can set them via the builder utils `.WithUrl` or `.WithToken`
// which will take priority over the env vars.
//
// This can be useful to switch to the preview token
func NewContentfulClient() *ContentfulClient {
	url := os.Getenv("CONTENTFUL_URL")
	token := os.Getenv("CONTENTFUL_TOKEN")

	return &ContentfulClient{
		url, token,
	}
}

func (c *ContentfulClient) WithUrl(url string) *ContentfulClient {
	c.url = url
	return c
}

func (c *ContentfulClient) WithToken(token string) *ContentfulClient {
	c.token = token
	return c
}

func (c *ContentfulClient) validate() error {
	if len(c.url) == 0 {
		return ErrUrlNotSet
	}
	if len(c.token) == 0 {
		return ErrTokenNotSet
	}

	return nil
}

func (c *ContentfulClient) Get(req ContentfulRequest, target any) error {
	if err := c.validate(); err != nil {
		return err
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r, err := http.NewRequest(http.MethodPost, c.url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", fmt.Sprintf("Bearer %s", c.token))

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return err
	}

	return nil
}

func (c *ContentfulClient) GetOrFetch(cacher ContentfulCacher, prefix string, req ContentfulRequest, target any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	key := cacher.GenerateKey(prefix, &req)
	payload, err := cacher.Get(ctx, key)
	if err != nil {
		log.Printf("[ccgo] Failed to find %s in cache", prefix)
	} else {
		if err = payload.Decode(target); err != nil {
			log.Printf("[ccgo] Failed to decode payload %s", payload)
		}
	}

	// Cache was found
	if err == nil && target != nil {
		return nil
	}

	err = c.Get(req, target)
	if err != nil {
		return err
	}

	err = cacher.Put(ctx, key, target)
	if err != nil {
		return err
	}

	return nil
}
