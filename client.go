package contenfulclientgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
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

func (c *ContentfulClient) Get(req ContentfulRequest, target any) error {
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

	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}

	return nil
}

func (c *ContentfulClient) GetOrFetch(cacher ContentfulCacher, prefix string, req ContentfulRequest, target any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	key := cacher.generateKey(prefix, &req)
	err := cacher.get(ctx, key, target)
	if err != nil {
		return err
	}

	// Cache was found
	if target != nil {
		return nil
	}

	err = c.Get(req, target)
	if err != nil {
		return err
	}

	err = cacher.put(ctx, key, target)
	if err != nil {
		return err
	}

	return nil
}
