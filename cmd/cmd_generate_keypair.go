package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/crypto"
	"github.com/spf13/cobra"
)

type generateKeypairCmdOptions struct {
	Path string
}

func NewGenerateKeypairCommand() *cobra.Command {
	opts := &generateKeypairCmdOptions{}

	cmd := &cobra.Command{
		Use:   "generate-keypair",
		Short: "Generate new public/private keypair for encryption and signature generation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateKeypairHandler(opts, cmd, args)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.Path, "path", "/data/keys", `Path to save to key pair file`)

	return cmd
}

func generateKeypairHandler(opts *generateKeypairCmdOptions, _ *cobra.Command, _ []string) error {
	fmt.Printf("Generating key pair\n")
	privKeyBytes := make([]byte, 32)

	_, err := rand.Read(privKeyBytes)
	if err != nil {
		return errors.Wrap(errs.SomethingWentWrong, "random bytes")
	}
	_, pubKey := btcec.PrivKeyFromBytes(privKeyBytes)
	serializedPubKey := pubKey.SerializeCompressed()

	// fmt.Println(hex.EncodeToString(privKeyBytes))

	fmt.Printf("Public key: %s\n", hex.EncodeToString(serializedPubKey))
	err = os.MkdirAll(opts.Path, 0o755)
	if err != nil {
		return errors.Wrap(errs.SomethingWentWrong, "create directory")
	}

	privateKeyPath := path.Join(opts.Path, "priv.key")

	_, err = os.Stat(privateKeyPath)
	if err == nil {
		fmt.Printf("Existing private key found at %s\n[WARNING] THE EXISTING PRIVATE KEY WILL BE LOST\nType [replace] to replace existing private key: ", privateKeyPath)
		var ans string
		fmt.Scanln(&ans)
		if ans != "replace" {
			fmt.Printf("Keypair generation aborted\n")
			return nil
		}
	}

	err = os.WriteFile(privateKeyPath, []byte(hex.EncodeToString(privKeyBytes)), 0o644)
	if err != nil {
		return errors.Wrap(err, "write private key file")
	}
	fmt.Printf("Private key saved at %s\n", privateKeyPath)

	wifKeyPath := path.Join(opts.Path, "priv_wif_mainnet.key")
	client, err := crypto.New(hex.EncodeToString(privKeyBytes))
	if err != nil {
		return errors.Wrap(err, "new crypto client")
	}
	wifKey, err := client.WIF(&chaincfg.MainNetParams)
	if err != nil {
		return errors.Wrap(err, "get WIF key")
	}

	err = os.WriteFile(wifKeyPath, []byte(wifKey), 0o644)
	if err != nil {
		return errors.Wrap(err, "write WIF private key file")
	}
	fmt.Printf("WIF private key saved at %s\n", wifKeyPath)

	publicKeyPath := path.Join(opts.Path, "pub.key")
	err = os.WriteFile(publicKeyPath, []byte(hex.EncodeToString(serializedPubKey)), 0o644)
	if err != nil {
		return errors.Wrap(errs.SomethingWentWrong, "write public key file")
	}
	fmt.Printf("Public key saved at %s\n", publicKeyPath)
	return nil
}
