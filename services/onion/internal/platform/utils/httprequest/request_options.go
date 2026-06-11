package httprequest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"time"

	"github.com/beevik/etree"

	"github.com/voxire/onion/internal/platform/logger"
)

type (
	RequestOptions    func(c *requestConfiguration)
	unmarshalFunction func(ctx context.Context, bodyBytes []byte, target any) bool
)

type basicAuth struct {
	username string
	password string
}

type requestConfiguration struct {
	headers       Headers
	contentType   string
	timeout       time.Duration
	maxBodyBytes  int64
	unmarshalFunc unmarshalFunction
	targetStruct  any
	Certificate   *tls.Certificate
	CertPool      *x509.CertPool
	basicAuth     *basicAuth
}

func WithHeaders(headers Headers) RequestOptions {
	return func(c *requestConfiguration) {
		c.headers = headers
	}
}

func WithContentType(contentType string) RequestOptions {
	return func(c *requestConfiguration) {
		c.contentType = contentType
	}
}

func WithTarget(someTarget any) RequestOptions {
	return func(c *requestConfiguration) {
		c.targetStruct = someTarget
	}
}

func WithCertificate(cert tls.Certificate) RequestOptions {
	return func(c *requestConfiguration) {
		c.Certificate = &cert
	}
}

func WithBasicAuth(username, password string) RequestOptions {
	return func(c *requestConfiguration) {
		c.basicAuth = &basicAuth{username, password}
	}
}

func UnmarshalWith(unmarshalFunc unmarshalFunction) RequestOptions {
	return func(c *requestConfiguration) {
		c.unmarshalFunc = unmarshalFunc
	}
}

func UnmarshalJSON() RequestOptions {
	return UnmarshalWith(jsonUnmarshal)
}

func UnmarshalXML() RequestOptions {
	return UnmarshalWith(xmlUnmarshal)
}

func UnmarshalETreeXML() RequestOptions {
	return UnmarshalWith(eTreeXMLUnmarshal)
}

func WithTimeout(duration time.Duration) RequestOptions {
	return func(c *requestConfiguration) {
		c.timeout = duration
	}
}

func WithMaxResponseBytes(maxBytes int64) RequestOptions {
	return func(c *requestConfiguration) {
		if maxBytes > 0 {
			c.maxBodyBytes = maxBytes
		}
	}
}

func getDefaultConfig() requestConfiguration {
	const (
		defaultTimeout          = 5 * time.Second
		defaultMaxResponseBytes = 8 << 20
	)

	return requestConfiguration{
		unmarshalFunc: jsonUnmarshal,
		timeout:       defaultTimeout,
		maxBodyBytes:  defaultMaxResponseBytes,
	}
}

func getConfigs(options ...RequestOptions) requestConfiguration {
	defaultConfig := getDefaultConfig()
	for _, option := range options {
		option(&defaultConfig)
	}

	return defaultConfig
}

func jsonUnmarshal(ctx context.Context, resp []byte, theObj any) bool {
	err := json.Unmarshal(resp, theObj)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to json unmarshal response",
			logger.Error(err),
			logger.String("response", string(resp)),
		)

		return false
	}

	return true
}

func xmlUnmarshal(ctx context.Context, resp []byte, theObj any) bool {
	err := xml.Unmarshal(resp, theObj)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to xml unmarshal response",
			logger.Error(err),
			logger.String("response", string(resp)),
		)

		return false
	}

	return true
}

func eTreeXMLUnmarshal(ctx context.Context, resp []byte, theObj any) bool {
	switch obj := theObj.(type) {
	case *etree.Document:
		if len(resp) == 0 {
			return false
		}

		err := obj.ReadFromBytes(resp)
		if err != nil {
			logger.Ctx(ctx).Warn("Failed to unmarshal etree xml response",
				logger.Error(err),
			)

			return false
		}
	default:
		logger.Ctx(ctx).Warn("Only etree.Document is supported for eTreeXMLUnmarshal")

		return false
	}

	return true
}
