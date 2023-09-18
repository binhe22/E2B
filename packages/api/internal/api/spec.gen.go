// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9xYUW/jNgz+K4K2RzfO9boXv7XX7hbccCvaAhtQFINi07WutuRJdHpB4P8+ULJjO3HS",
	"pFuLYU9xLIoUyY8fKa94rItSK1BoebTipTCiAATj/oFazC7pQSoe8VJgxgOuRAE8atYCbuCvShpIeISm",
	"goDbOINC0CZcliRo0Uj1yOs64FJZFCqGnUp7AsdorknYllpZcOc+m07pJ9YKQSE9irLMZSxQahV+s1rR",
	"u07fjwZSHvEfwi4YoV+14ZUx2ngbCdjYyJKU8IhfiITREcEirwN+Nv3w9jbPK8xAYaOVgZcj42dvb/yr",
	"RpbqSiVk8af3CPEtmAWY1s26hYDL8ZVaSKNV0VgvjS7BoIQhcIf6ZgkFL5VgmE4ZZsCgpyXYxFXAy2qe",
	"y3hb0e8ZYAZmUwWTlvktTBumVb5kIo7BWjnPgc2XTh5BFJ2tudY5CEXGLAqs7LaxW/d+/MSgqoJH93xe",
	"yTyhQ1MliGRJSy5qD8FIHXaVdb+u48b62ueHOuA+MVvRjXUC28d0wsytBTzVphDoKho/nnbuSoXwCA6z",
	"BVgrHncp4i+duzHUaqHjzhr2GDlxLkHhYZDwsmNoeA2qWGp0wZ4zGWeEDlptSY7FBgRCMmZqyJQv2Wul",
	"+aHJHvDsOjgUw6/wvDuMBwegtbm/vkYP9+DqHOLKSFzeUr172+eukO70EyhiQXo1B2HA/NxCzZfan0gi",
	"vOEKV2JOrDOfIZYU4/NSfoFlq8x1owxE4kSbfvTHyfn17OQLLLvdwu3ybCVVql0/kpjT2tXpBTu/nvGA",
	"L8BYH5zp5MNkSuZ0CUqUkkf842Q6mVKlCcycbyGohXt4BNwO76/SIhN53g8mFSplxtHtLOER/wx4RVo2",
	"WuHpkTwtEQq3UeT5bymP7l+g7l5+a8JPGyVjxHKU0iuXpbTK8yUzgJVRkGx71zXUMetrD0MS6rrRflkS",
	"6oPLObcFq/uHmjhTPNoGk9bVRantSGo+uQpmgil43sD6MDvX2nbpcTPDhU6WG5kpqhxlKQyGRJ8niUDh",
	"qy7WjtsJ8sTzn2jLd+ztvmux2aVWxwh4YtFQu+l653ZRb6ocenip4ycwzAmxuJHq8ftcKmGWYxSWuJ2p",
	"zGGXVlpjrf9H8O3lCM8Gvt2WBqxrxNtEbDNd5QmbEzuROy8T0iAyA4e6rqrn3yDGdgbtD6z1P6zEg8vu",
	"/1JkdeCJMFy5xNc7CfEz4HDqIhrewYdXTcfrX2x2cFonEnrk0Rn/WyncHACPzGBzTXhJ9uy9sp2ByH33",
	"Hc3zL26ZxRnET2P59es7Ot5Q1Y0nXfYsLLPr0B4XP5elsB2d/HV5b1dQ3ajnpsDt8X27SczW6vd1itfD",
	"rj/fHURap/+a6aHdYcjuMmh6TEYZQmFoMHbpmR6Snuk7kllvbBwiu4PGwxAq4aobuOvQQGrAZvsQdONF",
	"htcF+I6gaBBgEi1DWQBDzXK5gP1Imq1t36wtH0uIvQvDCCuejVxah/zl7SbD28rr2OttPzn4W2xzYBfr",
	"NvwnTHUfQF4JCrfNLNqgVybnEV/5V3W4+EA3B2GkmOc+tH7FRzcVVY7N9cVGYShKOYHT+SSBBTnSmVxt",
	"fk2jTr/qfbezNKr/HQAA///XvLXN+hMAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
