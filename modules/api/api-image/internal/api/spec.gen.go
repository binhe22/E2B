// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
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

	"H4sIAAAAAAAC/8xW32vbOhT+V8S599HUuS33xW/h3rKFrqwsUBglFM0+jtXZkisdp4Tg/30c2YmVOGkD",
	"K9teUlc6Oj++79NnbyA1VW00anKQbMClBVbSP071+l6WDUKyaSO41itlja5QE2/W1tRoSaEPTU2Gj06r",
	"ukZ6VBkvZehSq2pSRkMCsww1qVyhFSYXUvAB0R8QL4VKi/5XOUEFChyqCemcSZUkzCACWtcICTiySi+h",
	"jSDD2regCCv/MIroF6S1cs3/20aTqngsQN1UkDyANhk+OViM8nM4PjfKYsZxh3MOyfpGFoyUtcYex2gM",
	"jA/2cEAEubGVJEhAabq6HKZVmnCJlpuv0Dm5PJUIzhkAhizc7hyd8ykOG3bdxhl8MmN99JsdBFkXvOcw",
	"bayi9ZyF1xWe1uoG19OGCk8s13tu0K4hAi2ZOJjezR5vrr8OxaQ/Ai1nVDo3447/x9U3Y76L6d2Mjykq",
	"cViFCFZoOxjgn4vJxYSxNjVqWStI4MovRVBLKnyLMf8skcZ1PqIsqRBpgSmnZUAlb80ySOADEmsGXW20",
	"64a9nEzGSb7gc4OOxIt0wjVpis7lTcnjtRHEqFeeLeN8/f0Sd8bRtV6dKJMaTf0NlnVdqtQfjJ9cJ4Du",
	"9vPT3xZzSOCveLCHuPeGOLQC39F+8/Ndw+VapBb55gqpwzvtVUFy6VgRPI6XQtxrw50E95NyJGRZil3k",
	"EYTnw95PQbBzFFmWn3NIHl5HZWeX7WJkOm+AZJEaqxmlcLQ2gn/fkzVvEEda6SyoVI6UXgblA4p2izza",
	"Vnf7Wf7zRAu5TSCM7n3BrtCOeGKZvhdRr029dbezdbozsl8Lv68f4H8K/vCaxJvBTNuOkRIJj3kfr5/N",
	"TRe+Zaf/O8u8/1lZIaF1/jp4a2ZPHJw5sPfQ98k2GAVIHb4jFn+ABDr0fp8EfP1zJBC8NT0N4fvyYcFY",
	"9qcOdXArtVxyhfvb0EH3qWPjGX1cBJ9jubHi/nY4xvbdLtofAQAA///mIlRXSQoAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
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
	var res = make(map[string]func() ([]byte, error))
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
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
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
