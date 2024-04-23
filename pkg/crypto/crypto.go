package crypto

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	ecies "github.com/ecies/go/v2"
)

type Client struct {
	privateKey      *btcec.PrivateKey
	eciesPrivateKey *ecies.PrivateKey
}

func New(privateKeyStr string) (*Client, error) {
	if privateKeyStr != "" {
		privateKeyBytes, err := hex.DecodeString(privateKeyStr)
		println(len(privateKeyBytes))
		if err != nil {
			return nil, errors.Wrap(err, "decode private key")
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		eciesPrivateKey := ecies.NewPrivateKeyFromBytes(privateKeyBytes)
		return &Client{
			privateKey:      privateKey,
			eciesPrivateKey: eciesPrivateKey,
		}, nil
	}
	return &Client{}, nil
}

func (c *Client) Sign(message string) string {
	messageHash := chainhash.DoubleHashB([]byte(message))
	signature := ecdsa.Sign(c.privateKey, messageHash)
	return hex.EncodeToString(signature.Serialize())
}

func (c *Client) Verify(message, sigStr, pubKeyStr string) (bool, error) {
	sigBytes, err := hex.DecodeString(sigStr)
	if err != nil {
		return false, errors.Wrap(err, "signature decode")
	}

	pubBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		return false, errors.Wrap(err, "pubkey decode")
	}
	pubKey, err := btcec.ParsePubKey(pubBytes)
	if err != nil {
		return false, errors.Wrap(err, "pubkey parse")
	}

	messageHash := chainhash.DoubleHashB([]byte(message))

	signature, err := ecdsa.ParseSignature(sigBytes)
	if err != nil {
		return false, errors.Wrap(err, "signature parse")
	}
	return signature.Verify(messageHash, pubKey), nil
}

func (c *Client) Encrypt(message, pubKeyStr string) (string, error) {
	pubKey, err := ecies.NewPublicKeyFromHex(pubKeyStr)
	if err != nil {
		return "", errors.Wrap(err, "parse pubkey")
	}

	ciphertext, err := ecies.Encrypt(pubKey, []byte(message))
	if err != nil {
		return "", errors.Wrap(err, "encrypt message")
	}

	ciphertextStr := base64.StdEncoding.EncodeToString(ciphertext)
	return ciphertextStr, nil
}

func (c *Client) Decrypt(ciphertextStr string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextStr)
	if err != nil {
		return "", errors.Wrap(err, "decode ciphertext")
	}
	plaintext, err := ecies.Decrypt(c.eciesPrivateKey, ciphertext)
	if err != nil {
		return "", errors.Wrap(err, "decrypt")
	}
	return string(plaintext), nil
}
