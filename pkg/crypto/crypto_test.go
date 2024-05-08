package crypto

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

const (
	privateKeyStr = "ce9c2fd75623e82a83ed743518ec7749f6f355f7301dd432400b087717fed2f2"
	mainnetKey    = "L49LKamtrPZxty5TG7jaFPHMRZbrvAr4Dvn5BHGdvmvbcTDNAbZj"
	pubKeyStr     = "0251e2dfcdeea17cc9726e4be0855cd0bae19e64f3e247b10760cd76851e7df47e"
)

func TestEncryptDecrypt(t *testing.T) {
	plaintext := "hello world"

	privClient, err := New(privateKeyStr)
	assert.NoError(t, err)

	pubClient, err := New("")
	assert.NoError(t, err)

	ciphertext, err := pubClient.Encrypt(plaintext, pubKeyStr)
	assert.NoError(t, err)

	decrypted, err := privClient.Decrypt(ciphertext)
	assert.NoError(t, err)

	assert.Equal(t, plaintext, decrypted)
}

func TestSignVerify(t *testing.T) {
	plaintext := "hello world"
	invalidSignature := "3044022066504a82e2bc23167214e05497a1ca957add9cacc078aa69f5417079a4d56f0002206b215920b046c779d4a58d4029c26dbadcaf6d3c884b3463f44e70ef9146c1cd"

	privClient, err := New(privateKeyStr)
	assert.NoError(t, err)

	pubClient, err := New("")
	assert.NoError(t, err)

	signature := privClient.Sign(plaintext)
	println(signature)
	verified, err := pubClient.Verify(plaintext, signature, pubKeyStr)
	assert.NoError(t, err)

	assert.True(t, verified)

	verified, err = pubClient.Verify(plaintext, invalidSignature, pubKeyStr)
	assert.NoError(t, err)

	assert.False(t, verified)
}

func TestWIF(t *testing.T) {
	privClient, err := New(privateKeyStr)
	assert.NoError(t, err)

	wifPrivKey, err := privClient.WIF(&chaincfg.MainNetParams)
	assert.NoError(t, err)
	assert.Equal(t, wifPrivKey, mainnetKey)
}
