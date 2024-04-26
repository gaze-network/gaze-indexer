package reportingclient

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/pkg/httpclient"
	"github.com/gaze-network/indexer-network/pkg/logger"
)

type Config struct {
	Disabled      bool   `mapstructure:"disabled"`
	BaseURL       string `mapstructure:"base_url"`
	Name          string `mapstructure:"name"`
	WebsiteURL    string `mapstructure:"website_url"`
	IndexerAPIURL string `mapstructure:"indexer_api_url"`
}

type ReportingClient struct {
	httpClient *httpclient.Client
	config     Config
}

const defaultBaseURL = "https://indexer.api.gaze.network"

func New(config Config) (*ReportingClient, error) {
	baseURL := utils.Default(config.BaseURL, defaultBaseURL)
	httpClient, err := httpclient.New(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "can't create http client")
	}
	if config.Name == "" {
		return nil, errors.New("reporting.name config is required if reporting is enabled")
	}
	return &ReportingClient{
		httpClient: httpClient,
		config:     config,
	}, nil
}

type SubmitBlockReportPayload struct {
	Type                string         `json:"type"`
	ClientVersion       string         `json:"clientVersion"`
	DBVersion           int            `json:"dbVersion"`
	EventHashVersion    int            `json:"eventHashVersion"`
	Network             common.Network `json:"network"`
	BlockHeight         uint64         `json:"blockHeight"`
	BlockHash           chainhash.Hash `json:"blockHash"`
	EventHash           chainhash.Hash `json:"eventHash"`
	CumulativeEventHash chainhash.Hash `json:"cumulativeEventHash"`
}

func (r *ReportingClient) SubmitBlockReport(ctx context.Context, payload SubmitBlockReportPayload) error {
	ctx = logger.WithContext(ctx, slog.String("package", "reporting_client"), slog.Any("payload", payload))

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "can't marshal payload")
	}
	resp, err := r.httpClient.Post(ctx, "/v1/report/block", httpclient.RequestOptions{
		Body: body,
	})
	if err != nil {
		return errors.Wrap(err, "can't send request")
	}
	if resp.StatusCode() >= 400 {
		// TODO: unmashal response body and log it
		logger.WarnContext(ctx, "Reporting block event failed", slog.Any("resp_body", resp.Body()))
		return nil
	}
	logger.DebugContext(ctx, "Reported block event")
	return nil
}

type SubmitNodeReportPayload struct {
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	Network       common.Network `json:"network"`
	WebsiteURL    string         `json:"websiteURL,omitempty"`
	IndexerAPIURL string         `json:"indexerAPIURL,omitempty"`
}

func (r *ReportingClient) SubmitNodeReport(ctx context.Context, module string, network common.Network) error {
	payload := SubmitNodeReportPayload{
		Name:          r.config.Name,
		Type:          module,
		Network:       network,
		WebsiteURL:    r.config.WebsiteURL,
		IndexerAPIURL: r.config.IndexerAPIURL,
	}

	ctx = logger.WithContext(ctx, slog.String("package", "reporting_client"), slog.Any("payload", payload))

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "can't marshal payload")
	}
	resp, err := r.httpClient.Post(ctx, "/v1/report/node", httpclient.RequestOptions{
		Body: body,
	})
	if err != nil {
		return errors.Wrap(err, "can't send request")
	}
	if resp.StatusCode() >= 400 {
		logger.WarnContext(ctx, "Reporting node info failed", slog.Any("resp_body", resp.Body()))
	}
	logger.DebugContext(ctx, "Reported node info")
	return nil
}
