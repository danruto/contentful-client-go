# Contentful Client Go

This is just a simple contentful client to make the graphql calls a bit easier to manage since the official SDK
is in limbo.

# Usage

It provides two interfaces you can choose to use.

One is to not use a Cacher for the request and you can then handle it as you'd like, or you can implement the Cacher interface
and the library will use your Cacher implementation to cache the values

## No Cacher example
```go
package main

import ccgo "github.com/danruto/contentful-client-go"

type somethingItem struct {
  ccgo.ContentfulItem
  Name string `json:"name"`
}

type somethingData struct {
  Something ccgo.ContentfulCollectionItem[somethingItem] `json:"somethingCollection"`
}

func main() {
  client := ccgo.NewContentfulClient()

  req := ccgo.ContentfulRequest {
    Query: `
      query GetSomething($limit: Int) {
        somethingCollection(limit: $limit) {
          items {
            sys { id }
            name
          }
        }
      }
    `,
    Variables: ccgo.ContentfulVariables{"limit": 1},
  }

  var collection ccgo.ContentfulCollection[somethingData]
  if err := client.Get(req, &collection); err != nil {
    return
  }

  fmt.Println(collection)
}

```

## Implementing and using a custom Cacher
```go

package main

import (
  "fmt"
  "context"
  ccgo "github.com/danruto/contentful-client-go"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// Init the pool ...

type somethingItem struct {
  ccgo.ContentfulItem
  Name string `json:"name"`
}

type somethingData struct {
  Something ccgo.ContentfulCollectionItem[somethingItem] `json:"somethingCollection"`
}

type dbCacher struct {
  ccgo.BaseContentfulCacher
}

func (c dbCacher) GenerateKey(prefix string, req *ccgo.ContentfulRequest) string {
	return c.BaseContentfulCacher.GenerateKey(prefix, req)
}

func (c dbCacher) Get(ctx context.Context, key string) (*ccgo.ContentfulAny, error) {
	row := DB.QueryRow(ctx, `SELECT payload,updated_at FROM contentful_cache WHERE key = $1`, key)

	var cacheItem contentfulCacheItem[string]
	err := row.Scan(&cacheItem.Payload, &cacheItem.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Check cutoff times
	cutoff := time.Now()
	// 1 day back
	cutoff = cutoff.AddDate(0, 0, -1)
	diff := cutoff.Sub(cacheItem.UpdatedAt)
	if diff.Hours() < 24 {
		return &ccgo.ContentfulAny{Data: []byte(cacheItem.Payload)}, nil
	}

	// Invalid
	return nil, errors.New("Cache has expired")
}

func (c cacher) Put(ctx context.Context, key string, src any) error {
	tx, err := DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			// Log warning rolling back
		}
	}()

	_, err = tx.Exec(ctx,
		`INSERT INTO contentful_cache (key, payload)
        VALUES ($1, $2)
        ON CONFLICT (key)
        DO UPDATE
        SET payload = $2`, key, src)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func main() {
  client := ccgo.NewContentfulClient().WithToken("<read token from somewhere>")

  req := ccgo.ContentfulRequest {
    Query: `
      query GetSomething($limit: Int) {
        somethingCollection(limit: $limit) {
          items {
            sys { id }
            name
          }
        }
      }
    `,
    Variables: ccgo.ContentfulVariables{"limit": 1},
  }

  var collection ccgo.ContentfulCollection[somethingData]
  if err := client.GetOrFetch(dbc, "GetSomething", req, &collection); err != nil {
    return
  }

  fmt.Println(collection)
}
```
