package contenfulclientgo_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	ccgo "github.com/danruto/contentful-client-go"
)

const (
    testToken = "9d5de88248563ebc0d2ad688d0473f56fcd31c600e419d6c8962f6aed0150599"
    testUrl = "https://graphql.contentful.com/content/v1/spaces/f8bqpb154z8p/environments/master"
)

type testcacher struct {
	ccgo.BaseContentfulCacher

	cache        map[string]*ccgo.ContentfulAny
	expectExists bool
}

func newTestCacher() testcacher {
	return testcacher{
		BaseContentfulCacher: ccgo.BaseContentfulCacher{},
		cache:                map[string]*ccgo.ContentfulAny{},
		expectExists:         false,
	}
}

func (c testcacher) GenerateKey(prefix string, req *ccgo.ContentfulRequest) string {
	return c.BaseContentfulCacher.GenerateKey(prefix, req)
}

func (c testcacher) Get(ctx context.Context, key string) (*ccgo.ContentfulAny, error) {
	v, ok := c.cache[key]
	if ok {
		fmt.Printf("Returning from cache for %s: %v\n", key, v)
		return v, nil
	}

	if c.expectExists {
		return nil, errors.New("Expected key to exist in cache")
	}

	return nil, errors.New("Not found in cache")
}

func (c testcacher) Put(ctx context.Context, key string, src any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	// Convert from src to any type
	v := ccgo.ContentfulAny{
		Data: data,
	}
	c.cache[key] = &v

	return nil
}

func TestSimpleQuery(t *testing.T) {
	c := ccgo.NewContentfulClient().WithToken(testToken).WithUrl(testUrl)

	req := ccgo.ContentfulRequest{
		Query: `
		query GetLessonCopys($limit:Int!) {
		  lessonCopyCollection(limit:$limit) {
		    items {
		      sys {
		        id
		      }
		    }
		  }
		}
		`,
		Variables: ccgo.ContentfulVariables{"limit": 2},
	}

	var target ccgo.ContentfulCollection[struct {
		LessionCopyCollection ccgo.ContentfulCollectionItem[ccgo.ContentfulItem] `json:"lessonCopyCollection"`
	}]
	if err := c.Get(req, &target); err != nil {
		t.Fatal(err)
	}
	if len(target.Data.LessionCopyCollection.Items) != 2 {
		t.Logf("target: %v", target)
		t.Fatal("Got an invalid number of items in return")
	}

	expectedOne := "5jR4ciJ8Y8m2KKqmMKOkg4"
	expectedTwo := "3k6uoYm9i8MycCm42IsY62"

	if target.Data.LessionCopyCollection.Items[0].GetID() != expectedOne {
		t.Fatalf("First item ID was incorrect. Got %s -> Expected %s", target.Data.LessionCopyCollection.Items[0], expectedOne)
	}

	if target.Data.LessionCopyCollection.Items[1].GetID() != expectedTwo {
		t.Fatalf("Second item ID was incorrect. Got %s -> Expected %s", target.Data.LessionCopyCollection.Items[1], expectedTwo)
	}
}

func TestSimpleQueryWithCacher(t *testing.T) {
	c := ccgo.NewContentfulClient().WithToken(testToken).WithUrl(testUrl)
	cacher := newTestCacher()

	req := ccgo.ContentfulRequest{
		Query: `
		query GetLessonCopys($limit:Int!) {
		  lessonCopyCollection(limit:$limit) {
		    items {
		      sys {
		        id
		      }
		    }
		  }
		}
		`,
		Variables: ccgo.ContentfulVariables{"limit": 2},
	}

	var target ccgo.ContentfulCollection[struct {
		LessionCopyCollection ccgo.ContentfulCollectionItem[ccgo.ContentfulItem] `json:"lessonCopyCollection"`
	}]
	if err := c.GetOrFetch(cacher, "TestSimpleWithCacher", req, &target); err != nil {
		t.Fatal(err)
	}
	if len(target.Data.LessionCopyCollection.Items) != 2 {
		t.Logf("target: %v", target)
		t.Fatal("Got an invalid number of items in return")
	}

	expectedOne := "5jR4ciJ8Y8m2KKqmMKOkg4"
	expectedTwo := "3k6uoYm9i8MycCm42IsY62"

	if target.Data.LessionCopyCollection.Items[0].GetID() != expectedOne {
		t.Fatalf("First item ID was incorrect. Got %s -> Expected %s", target.Data.LessionCopyCollection.Items[0], expectedOne)
	}

	if target.Data.LessionCopyCollection.Items[1].GetID() != expectedTwo {
		t.Fatalf("Second item ID was incorrect. Got %s -> Expected %s", target.Data.LessionCopyCollection.Items[1], expectedTwo)
	}

	// The second time should hit the cache
	cacher.expectExists = true
	var target2 ccgo.ContentfulCollection[struct {
		LessionCopyCollection ccgo.ContentfulCollectionItem[ccgo.ContentfulItem] `json:"lessonCopyCollection"`
	}]
	if err := c.GetOrFetch(cacher, "TestSimpleWithCacher", req, &target2); err != nil {
		t.Fatal(err)
	}
	if len(target.Data.LessionCopyCollection.Items) != 2 {
		t.Logf("target: %v", target2)
		t.Fatal("Got an invalid number of items in return")
	}

	if target2.Data.LessionCopyCollection.Items[0].GetID() != expectedOne {
		t.Fatalf("First item ID was incorrect. Got %s -> Expected %s", target2.Data.LessionCopyCollection.Items[0], expectedOne)
	}

	if target2.Data.LessionCopyCollection.Items[1].GetID() != expectedTwo {
		t.Fatalf("Second item ID was incorrect. Got %s -> Expected %s", target2.Data.LessionCopyCollection.Items[1], expectedTwo)
	}
}
