// Copyright 2023 Adevinta

// Package checktypes provides utilities for working with checktypes
// and chektype catalogs.
package checktypes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	checkcatalog "github.com/adevinta/vulcan-check-catalog/pkg/model"
	types "github.com/adevinta/vulcan-types"

	"github.com/adevinta/lava/internal/checktype/build"
	"github.com/adevinta/lava/internal/urlutil"
)

var (
	// ErrMalformedCatalog is returned by [NewCatalog] when the format of
	// the retrieved catalog is not valid.
	ErrMalformedCatalog = errors.New("malformed catalog")

	// ErrMissingCatalog is returned by [NewCatalog] when no
	// catalog URLs are provided.
	ErrMissingCatalog = errors.New("missing catalog URLs")

	// ErrInvalidURL is returned by the [NewCatalog] one of the provided
	// catalog URL's is not valid.
	ErrInvalidURL = errors.New("invalid URL")
)

// Checktype represents a Vulcan checktype.
type Checktype checkcatalog.Checktype

// Accepts reports whether the specified checktype accepts an asset
// type.
func Accepts(ct checkcatalog.Checktype, at types.AssetType) bool {
	for _, accepted := range ct.Assets {
		if accepted == string(at) {
			return true
		}
	}
	return false
}

// Catalog represents a collection of Vulcan checktypes.
type Catalog map[string]checkcatalog.Checktype

// NewCatalog retrieves the specified checktype catalogs and
// consolidates them in a single catalog with all the checktypes
// indexed by name. If a checktype is duplicated it is overridden with
// the last one.
func NewCatalog(urls []string) (Catalog, error) {
	if len(urls) == 0 {
		return nil, ErrMissingCatalog
	}
	checktypes := make(Catalog)
	for _, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidURL, err)
		}
		// var decData struct {
		// 	Checktypes []Checktype `json:"checktypes"`
		// }
		isDir, err := isDir(parsedURL)
		if err != nil {
			return nil, err
		}
		// If the url points to a directory, Lava considers the it points to
		// the code of a checktype defined in that directory.
		if isDir {
			code := build.Code(parsedURL.Path)
			checktype, err := code.Build(context.Background())
			if err != nil {
				return nil, err
			}
			checktypes[checktype.Name] = checktype
			continue
		}
		data, err := urlutil.Get(parsedURL)
		if err != nil {
			return nil, err
		}

		var decData struct {
			Checktypes []checkcatalog.Checktype `json:"checktypes"`
		}
		err = json.Unmarshal(data, &decData)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMalformedCatalog, err)
		}

		for _, checktype := range decData.Checktypes {
			checktypes[checktype.Name] = checktype
		}
	}
	return checktypes, nil
}

// isDir returns true if a URL points to a local existing directory.
func isDir(u *url.URL) (bool, error) {
	if u.Scheme != "" {
		return false, nil
	}
	info, err := os.Stat(u.Path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
