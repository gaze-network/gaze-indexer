package reportingclientv2

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/pkg/crypto"
	"github.com/gaze-network/indexer-network/pkg/httpclient"
	"github.com/gaze-network/indexer-network/pkg/logger"
)

type Config struct {
	Disabled     bool         `mapstructure:"disabled"`
	ReportCenter ReportCenter `mapstructure:"report_center"`
	NodeInfo     NodeInfo     `mapstructure:"node_info"`
}

type NodeInfo struct {
	Name       string `mapstructure:"name"`
	WebsiteURL string `mapstructure:"website_url"`
	APIURL     string `mapstructure:"api_url"`
}

type ReportCenter struct {
	BaseURL   string `mapstructure:"base_url"`
	PublicKey string `mapstructure:"public_key"`
}

type ReportingClient struct {
	httpClient   *httpclient.Client
	cryptoClient *crypto.Client
	config       Config
}

const (
	defaultBaseURL   = "https://indexer.api.gaze.network"
	defaultPublicKey = "0251e2dfcdeea17cc9726e4be0855cd0bae19e64f3e247b10760cd76851e7df47e"
)

func New(config Config, indexerPrivateKey string) (*ReportingClient, error) {
	config.ReportCenter.BaseURL = utils.Default(config.ReportCenter.BaseURL, defaultBaseURL)
	config.ReportCenter.PublicKey = utils.Default(config.ReportCenter.PublicKey, defaultPublicKey)
	httpClient, err := httpclient.New(config.ReportCenter.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "can't create http client")
	}

	cryptoClient, err := crypto.New(indexerPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "can't create crypto client")
	}
	return &ReportingClient{
		httpClient:   httpClient,
		config:       config,
		cryptoClient: cryptoClient,
	}, nil
}

type SubmitBlockReportPayload struct {
	EncryptedData    string `json:"encryptedData"`
	IndexerPublicKey string `json:"indexerPublicKey"`
}

type SubmitBlockReportPayloadData struct {
	Type                string         `json:"type"`
	ClientVersion       string         `json:"clientVersion"`
	DBVersion           int            `json:"dbVersion"`
	EventHashVersion    int            `json:"eventHashVersion"`
	Network             common.Network `json:"network"`
	BlockHeight         uint64         `json:"blockHeight"`
	BlockHash           chainhash.Hash `json:"blockHash"`
	EventHash           chainhash.Hash `json:"eventHash"`
	CumulativeEventHash chainhash.Hash `json:"cumulativeEventHash"`
	IndexerPublicKey    string         `json:"indexerPublicKey"`
}

func (r *ReportingClient) SubmitBlockReport(ctx context.Context, payload SubmitBlockReportPayloadData) error {
	payload.IndexerPublicKey = r.cryptoClient.PublicKey()
	data, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "can't marshal payload")
	}

	encryptedData, err := r.cryptoClient.Encrypt(string(data), r.config.ReportCenter.PublicKey)
	if err != nil {
		return errors.Wrap(err, "can't encrypt data")
	}
	bodyStruct := SubmitBlockReportPayload{
		EncryptedData:    encryptedData,
		IndexerPublicKey: r.cryptoClient.PublicKey(),
	}
	body, err := json.Marshal(bodyStruct)
	if err != nil {
		return errors.Wrap(err, "can't marshal payload")
	}
	resp, err := r.httpClient.Post(ctx, "/v2/report/block", httpclient.RequestOptions{
		Body: body,
	})
	if err != nil {
		return errors.Wrap(err, "can't send request")
	}
	if resp.StatusCode() >= 400 {
		logger.WarnContext(ctx, "failed to submit block report", slog.Any("payload", payload), slog.Any("responseBody", resp.Body()))
	}
	logger.DebugContext(ctx, "block report submitted", slog.Any("payload", payload))
	return nil
}

type SubmitNodeReportPayload struct {
	Data             SubmitNodeReportPayloadData `json:"data"`
	IndexerPublicKey string                      `json:"indexerPublicKey"`
	Signature        string                      `json:"string"`
}

type SubmitNodeReportPayloadData struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Network    common.Network `json:"network"`
	WebsiteURL string         `json:"websiteUrl,omitempty"`
	APIURL     string         `json:"apiUrl,omitempty"`
}

func (r *ReportingClient) SubmitNodeReport(ctx context.Context, module string, network common.Network) error {
	payload := SubmitNodeReportPayload{
		Data: SubmitNodeReportPayloadData{
			Name:       r.config.NodeInfo.Name,
			Type:       module,
			Network:    network,
			WebsiteURL: r.config.NodeInfo.WebsiteURL,
			APIURL:     r.config.NodeInfo.APIURL,
		},
		IndexerPublicKey: r.cryptoClient.PublicKey(),
	}

	dataPayload, err := json.Marshal(payload.Data)
	if err != nil {
		return errors.Wrap(err, "can't marshal payload data")
	}

	payload.Signature = r.cryptoClient.Sign(string(dataPayload))

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "can't marshal payload")
	}
	resp, err := r.httpClient.Post(ctx, "/v2/report/node", httpclient.RequestOptions{
		Body: body,
	})
	if err != nil {
		return errors.Wrap(err, "can't send request")
	}
	if resp.StatusCode() >= 400 {
		logger.WarnContext(ctx, "failed to submit node report", slog.Any("payload", payload), slog.Any("responseBody", resp.Body()))
	}
	logger.InfoContext(ctx, "node report submitted", slog.Any("payload", payload))
	return nil
}
