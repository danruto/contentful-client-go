package contenfulclientgo

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrContentfulInvalidRequest = errors.New("Invalid `ContentfulRequest`")

type ContentfulVariables = map[string]any

type ContentfulRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type ContentfulCacheItem[T any] struct {
	Payload   T         `json:"payload"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ContentfulSys struct {
	ID string `json:"id"`
}

// The base struct for a single contentful item
// that you can then inherit for your own objects
type ContentfulItem struct {
	Sys ContentfulSys
}

type ContentfulItemIDExtract interface {
	GetID() string
}

func (ci ContentfulItem) GetID() string {
	return ci.Sys.ID
}

func ContentfulItemSliceToIDSlice[T ContentfulItemIDExtract](slice []T) []string {
	ret := make([]string, len(slice))

	for ii, item := range slice {
		ret[ii] = item.GetID()
	}

	return ret
}

// The contentful collection response that wraps your collection of items
type ContentfulCollection[T any] struct {
	Data T `json:"data"`
}

// The contenful collection response nested item convenience
type ContentfulCollectionItem[T any] struct {
	Items []T `json:"items"`
}

type ContentfulCacher interface {
	generateKey(prefix string, req *ContentfulRequest) string
	get(ctx context.Context, key string, target any) error
	put(ctx context.Context, key string, src any) error
}

type BaseContentfulCacher struct {
	ContentfulCacher
}

func (bcc BaseContentfulCacher) generateKey(prefix string, req *ContentfulRequest) string {
	var variables string

	if len(req.Variables) > 0 {
		variables = fmt.Sprintf("%v", req.Variables)
	}

	return fmt.Sprintf("%s-%s", prefix, variables)
}
