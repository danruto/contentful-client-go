# Contentful Client Go

This is just a simple contentful client to make the graphql calls a bit easier to manage since the official SDK
is in limbo.

# Usage

## Without the custom cacher
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
  if err = client.Get(req, &collection); err != nil {
    return
  }

  fmt.Println(collection)
}

```

## With the custom cacher

```go

package main

import (
  "fmt"
  "context"
  ccgo "github.com/danruto/contentful-client-go"
)

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

func (c *dbCacher) get(ctx context.Context, key string, target any) error{
  // Connect to something like a db and query from the table for `key`
  // then check if the entry has expired (say a 1 hour sliding window)
  // if it hasn't then save it to target using
  // json.Unmarshal()
  return nil
}

func (c *dbCacher) put(ctx context.Context, key string, src any) error{
  // Similar to above, evict anything expired with this key and then
  // insert this new entry with the current timestamp
  return nil
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
  if err = client.GetOrFetch(dbc, "GetSomething", req, &collection); err != nil {
    return
  }

  fmt.Println(collection)
}
```
