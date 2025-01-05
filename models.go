package contenfulclientgo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrContentfulInvalidRequest = errors.New("Invalid `ContentfulRequest`")

type ContentfulVariables = map[string]any

type ContentfulRequest struct {
	Query     string              `json:"query"`
	Variables ContentfulVariables `json:"variables,omitempty"`
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
	Sys ContentfulSys `json:"sys"`
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

// The contenful collection response nested item convenience
type ContentfulCollectionItem[T any] struct {
	Items []T `json:"items"`
}

// The contentful collection response that wraps your collection of items
type ContentfulCollection[T any] struct {
	Data T `json:"data"`
}

type ContentfulAny struct {
	Data []byte
}

func (ca *ContentfulAny) Decode(out any) error {
	if err := json.Unmarshal(ca.Data, out); err != nil {
		return err
	}

	return nil
}

type ContentfulCacher interface {
	GenerateKey(prefix string, req *ContentfulRequest) string
	Get(ctx context.Context, key string) (*ContentfulAny, error)
	Put(ctx context.Context, key string, src any) error
}

type BaseContentfulCacher struct {
	ContentfulCacher
}

func (bcc BaseContentfulCacher) GenerateKey(prefix string, req *ContentfulRequest) string {
	var variables string

	if len(req.Variables) > 0 {
		variables = fmt.Sprintf("%v", req.Variables)
	}

	return fmt.Sprintf("%s-%s", prefix, variables)
}
