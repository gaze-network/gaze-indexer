package httpclient

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/valyala/fasthttp"
)

type Config struct {
	// Enable debug mode
	Debug bool

	// Default headers
	Headers map[string]string
}

type Client struct {
	baseURL string
	Config
}

func New(baseURL string, config ...Config) (*Client, error) {
	if _, err := url.Parse(baseURL); err != nil {
		return nil, errors.Join(errs.InvalidArgument, errors.Wrap(err, "can't parse base url"))
	}
	var cf Config
	if len(config) > 0 {
		cf = config[0]
	}
	if len(cf.Headers) == 0 {
		cf.Headers = make(map[string]string)
	}
	return &Client{
		baseURL: baseURL,
		Config:  cf,
	}, nil
}

type RequestOptions struct {
	path     string
	method   string
	Body     []byte
	Query    url.Values
	Header   map[string]string
	FormData url.Values
}

type HttpResponse struct {
	URL string
	fasthttp.Response
}

func (r *HttpResponse) UnmarshalBody(out any) error {
	err := json.Unmarshal(r.Body(), out)
	if err != nil {
		return errors.Wrapf(err, "can't unmarshal json body from %v, %v", r.URL, string(r.Body()))
	}
	return nil
}

func (h *Client) request(ctx context.Context, reqOptions RequestOptions) (*HttpResponse, error) {
	start := time.Now()
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(reqOptions.method)
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range reqOptions.Header {
		req.Header.Set(k, v)
	}
	parsedUrl := utils.Must(url.Parse(h.baseURL)) // checked in httpclient.New
	parsedUrl.Path = reqOptions.path
	parsedUrl.RawQuery = reqOptions.Query.Encode()

	// remove %20 from url (empty space)
	url := strings.TrimSuffix(parsedUrl.String(), "%20")
	url = strings.Replace(url, "%20?", "?", 1)
	req.SetRequestURI(url)
	if reqOptions.Body != nil {
		req.Header.SetContentType("application/json")
		req.SetBody(reqOptions.Body)
	} else if reqOptions.FormData != nil {
		req.Header.SetContentType("application/x-www-form-urlencoded")
		req.SetBodyString(reqOptions.FormData.Encode())
	}

	resp := fasthttp.AcquireResponse()
	startDo := time.Now()

	defer func() {
		if h.Debug {
			logger := logger.With(
				slog.String("method", reqOptions.method),
				slog.String("url", url),
				slog.Duration("duration", time.Since(start)),
				slog.Duration("latency", time.Since(startDo)),
				slog.Int("req_header_size", len(req.Header.Header())),
				slog.Int("req_content_length", req.Header.ContentLength()),
			)

			if resp.StatusCode() >= 0 {
				logger = logger.With(
					slog.Int("status_code", resp.StatusCode()),
					slog.String("resp_content_type", string(resp.Header.ContentType())),
					slog.Int("resp_content_length", len(resp.Body())),
				)
			}

			logger.InfoContext(ctx, "Finished make request", slog.String("package", "httpclient"))
		}

		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	if err := fasthttp.Do(req, resp); err != nil {
		return nil, errors.Wrapf(err, "url: %s", url)
	}

	httpResponse := HttpResponse{
		URL: url,
	}
	resp.CopyTo(&httpResponse.Response)

	return &httpResponse, nil
}

func (h *Client) Do(ctx context.Context, method, path string, reqOptions RequestOptions) (*HttpResponse, error) {
	reqOptions.path = path
	reqOptions.method = method
	return h.request(ctx, reqOptions)
}

func (h *Client) Get(ctx context.Context, path string, reqOptions RequestOptions) (*HttpResponse, error) {
	reqOptions.path = path
	reqOptions.method = fasthttp.MethodGet
	return h.request(ctx, reqOptions)
}

func (h *Client) Post(ctx context.Context, path string, reqOptions RequestOptions) (*HttpResponse, error) {
	reqOptions.path = path
	reqOptions.method = fasthttp.MethodPost
	return h.request(ctx, reqOptions)
}

func (h *Client) Put(ctx context.Context, path string, reqOptions RequestOptions) (*HttpResponse, error) {
	reqOptions.path = path
	reqOptions.method = fasthttp.MethodPut
	return h.request(ctx, reqOptions)
}

func (h *Client) Patch(ctx context.Context, path string, reqOptions RequestOptions) (*HttpResponse, error) {
	reqOptions.path = path
	reqOptions.method = fasthttp.MethodPatch
	return h.request(ctx, reqOptions)
}

func (h *Client) Delete(ctx context.Context, path string, reqOptions RequestOptions) (*HttpResponse, error) {
	reqOptions.path = path
	reqOptions.method = fasthttp.MethodDelete
	return h.request(ctx, reqOptions)
}
