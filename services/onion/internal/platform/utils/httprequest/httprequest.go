package httprequest

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/voxire/onion/internal/platform/logger"
)

const (
	ContentTypeKey   = "Content-Type"
	UserAgent        = "User-Agent"
	Authorization    = "Authorization"
	DefaultUserAgent = "Onion-HTTP-Client/1.0"
)

var (
	ErrResponseTooLarge = errors.New("response body too large")
	ErrUnmarshal        = errors.New("failed to unmarshal results")
)

type (
	Headers      map[string]string
	HTTPResponse struct {
		StatusCode       int
		Headers          http.Header
		finalRedirectURL string
		Body             []byte
	}
)

func Post(ctx context.Context, urlStr string, payload []byte, options ...RequestOptions) (HTTPResponse, error) {
	return execRequest(ctx, urlStr, "POST", payload, options...)
}

func Get(ctx context.Context, urlStr string, options ...RequestOptions) (HTTPResponse, error) {
	response, err := execRequest(ctx, urlStr, "GET", nil, options...)
	if response.StatusCode >= http.StatusBadRequest {
		logger.Ctx(ctx).Warn("Http GET request invalid status code",
			logger.Error(err),
			logger.String("url", urlStr),
			logger.String("response", string(response.Body)),
			logger.Any("code", response.StatusCode),
		)

		return response, err
	}

	return response, err
}

func Head(ctx context.Context, urlStr string, options ...RequestOptions) (HTTPResponse, error) {
	return execRequest(ctx, urlStr, "HEAD", nil, options...)
}

func Delete(ctx context.Context, urlStr string, options ...RequestOptions) (HTTPResponse, error) {
	return execRequest(ctx, urlStr, "DELETE", nil, options...)
}

func Put(ctx context.Context, urlStr string, payload []byte, options ...RequestOptions) (HTTPResponse, error) {
	return execRequest(ctx, urlStr, "PUT", payload, options...)
}

func execRequest(ctx context.Context, urlPath, method string, payload []byte, options ...RequestOptions) (HTTPResponse, error) {
	_, err := url.Parse(urlPath)
	if err != nil {
		logger.Ctx(ctx).Warn("Error parsing url", logger.Error(err), logger.String("url", urlPath))

		return HTTPResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, method, urlPath, bytes.NewBuffer(payload))
	if err != nil {
		logger.Ctx(ctx).Warn("Error create request",
			logger.Error(err),
			logger.String("method", method),
			logger.String("url", urlPath),
		)

		return HTTPResponse{}, err
	}

	configs := getConfigs(options...)
	if configs.contentType != "" {
		req.Header.Set(ContentTypeKey, configs.contentType)
	}

	if configs.basicAuth != nil {
		req.SetBasicAuth(configs.basicAuth.username, configs.basicAuth.password)
	}

	req.Header.Set(UserAgent, DefaultUserAgent)

	for key, val := range configs.headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{Timeout: configs.timeout}

	if configs.Certificate != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{*configs.Certificate},
				RootCAs:      configs.CertPool,
				MinVersion:   tls.VersionTLS12,
			},
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Ctx(ctx).Warn("Http request error",
			logger.Error(err),
			logger.String("method", method),
			logger.String("url", urlPath),
		)

		return HTTPResponse{}, err
	}

	defer func(Body io.ReadCloser) {
		if resp.Body != nil {
			closeErr := Body.Close()
			if closeErr != nil {
				logger.Ctx(ctx).Warn("Http request error while closing response body reader",
					logger.Error(closeErr),
					logger.String("url", urlPath),
					logger.String("method", method),
				)
			}
		}
	}(resp.Body)

	response := HTTPResponse{}
	response.StatusCode = resp.StatusCode
	response.Headers = resp.Header
	bodyBytes, err := readResponseBody(resp.Body, configs.maxBodyBytes)
	response.Body = bodyBytes

	if err != nil {
		logger.Ctx(ctx).Warn("Http request error when reading response",
			logger.Error(err),
			logger.String("url", urlPath),
		)

		return response, err
	}

	response.finalRedirectURL = resp.Request.URL.String()

	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices && configs.targetStruct != nil {
		success := configs.unmarshalFunc(ctx, response.Body, configs.targetStruct)
		if !success {
			logger.Ctx(ctx).Warn("Http request error when unmarshalling results",
				logger.Error(ErrUnmarshal),
				logger.String("url", urlPath),
			)

			return response, ErrUnmarshal
		}
	}

	return response, nil
}

func readResponseBody(body io.Reader, maxBodyBytes int64) ([]byte, error) {
	if maxBodyBytes <= 0 {
		return io.ReadAll(body)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(body, maxBodyBytes+1))
	if err != nil {
		return bodyBytes, err
	}
	if int64(len(bodyBytes)) > maxBodyBytes {
		return bodyBytes[:maxBodyBytes], ErrResponseTooLarge
	}

	return bodyBytes, nil
}

func (resp HTTPResponse) WasRedirected() (wasRedirected bool, finalURL string) {
	return resp.finalRedirectURL != "", resp.finalRedirectURL
}

func (resp HTTPResponse) IsOk() bool {
	return resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
}

func (resp HTTPResponse) Is4XX() bool {
	return resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError
}

func (resp HTTPResponse) Is5XX() bool {
	return resp.StatusCode >= http.StatusInternalServerError
}
