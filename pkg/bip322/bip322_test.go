package bip322

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyMessage(t *testing.T) {
	type testcase struct {
		Address   string
		Message   string
		Signature string // base64
		Expected  bool
	}
	testcases := []testcase{
		{
			Address:   "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:   "",
			Signature: "AkcwRAIgM2gBAQqvZX15ZiysmKmQpDrG83avLIT492QBzLnQIxYCIBaTpOaD20qRlEylyxFSeEA2ba9YOixpX8z46TSDtS40ASECx/EgAxlkQpQ9hYjgGu6EBCPMVPwVIVJqO4XCsMvViHI=",
			Expected:  true,
		},
		{
			Address:   "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:   "",
			Signature: "AkgwRQIhAPkJ1Q4oYS0htvyuSFHLxRQpFAY56b70UvE7Dxazen0ZAiAtZfFz1S6T6I23MWI2lK/pcNTWncuyL8UL+oMdydVgzAEhAsfxIAMZZEKUPYWI4BruhAQjzFT8FSFSajuFwrDL1Yhy",
			Expected:  true,
		},
		{
			Address:   "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:   "Hello World",
			Signature: "AkcwRAIgZRfIY3p7/DoVTty6YZbWS71bc5Vct9p9Fia83eRmw2QCICK/ENGfwLtptFluMGs2KsqoNSk89pO7F29zJLUx9a/sASECx/EgAxlkQpQ9hYjgGu6EBCPMVPwVIVJqO4XCsMvViHI=",
			Expected:  true,
		},
		{
			Address:   "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:   "Hello World",
			Signature: "AkgwRQIhAOzyynlqt93lOKJr+wmmxIens//zPzl9tqIOua93wO6MAiBi5n5EyAcPScOjf1lAqIUIQtr3zKNeavYabHyR8eGhowEhAsfxIAMZZEKUPYWI4BruhAQjzFT8FSFSajuFwrDL1Yhy",
			Expected:  true,
		},
		{
			Address:   "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:   "",
			Signature: "INVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVALIDINVA",
			Expected:  false,
		},
		{
			Address:   "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:   "",
			Signature: "AkgwRQIhAPkJ1Q4oYS0htvyuSFHLxRQpFAY56b70UvE7Dxazen0ZAiAtZfFz1S6T6I23MWI2lK/pcNTWncuyL8UL+oMdydVgzAEhAsfxIAMZZEKUPYWI4BruhAQjzFT8FSFSajuFwrDLXXXX",
			Expected:  false,
		},
		{
			Address:   "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:   "Hello World",
			Signature: "BkgwRQIhAOzyynlqt93lOKJr+wmmxIens//zPzl9tqIOua93wO6MAiBi5n5EyAcPScOjf1lAqIUIQtr3zKNeavYabHyR8eGhowEhAsfxIAMZZEKUPYWI4BruhAQjzFT8FSFSajuFwrDLXXXX",
			Expected:  false,
		},
		{
			Address:   "bc1ppv609nr0vr25u07u95waq5lucwfm6tde4nydujnu8npg4q75mr5sxq8lt3",
			Message:   "",
			Signature: "AUDVvVp7mCtPZtoORKYcMM+idx9yy5+z4TGeoI/PWEUscd5x0QYJ6IPQ/anBSMWPWSRPqHVrEjOIWhP9FsZSMFdG",
			Expected:  true,
		},
		{
			Address:   "bc1ppv609nr0vr25u07u95waq5lucwfm6tde4nydujnu8npg4q75mr5sxq8lt3",
			Message:   "",
			Signature: "AUDYeG/k6AL9pNuhgK8aJqxIqBIObX867yc3QgdfS70sWEdUg0Msv0Ps24Pt5aQmcI2wZdwI3Egp5tA5PW+wTOw6",
			Expected:  true,
		},
		{
			Address:   "bc1ppv609nr0vr25u07u95waq5lucwfm6tde4nydujnu8npg4q75mr5sxq8lt3",
			Message:   "Hello World",
			Signature: "AUCkOlzIYSN6T+QzENjlp61Pa2l4EyDDH8c4pFANOwoh3oGi/iZHscAExUSePhbS94KIMgcg+yNp+LsckO+AfLQQ",
			Expected:  true,
		},
		{
			Address:   "bc1ppv609nr0vr25u07u95waq5lucwfm6tde4nydujnu8npg4q75mr5sxq8lt3",
			Message:   "Hello World",
			Signature: "AUD5MwxtURP3tAip3fS5vVRwa4L15wEyTIG0BQ3DPktJpXvQe7Sh8kf+mVaO4ldEP+vhiVZ/sXvOHEbQQnsiYpCq",
			Expected:  true,
		},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s_%s", tc.Address, tc.Message), func(t *testing.T) {
			address, err := btcutils.SafeNewAddress(tc.Address)
			require.NoError(t, err)
			signature, err := base64.StdEncoding.DecodeString(tc.Signature)
			require.NoError(t, err)

			verified := VerifyMessage(&address, signature, tc.Message)
			assert.Equal(t, tc.Expected, verified)
		})
	}
}

func TestSignMessage(t *testing.T) {
	type testcase struct {
		PrivateKey *btcec.PrivateKey
		Address    string
		Message    string
	}

	testcases := []testcase{
		{
			PrivateKey: lo.Must(btcutil.DecodeWIF("L3VFeEujGtevx9w18HD1fhRbCH67Az2dpCymeRE1SoPK6XQtaN2k")).PrivKey,
			Address:    "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:    "",
		},
		{
			PrivateKey: lo.Must(btcutil.DecodeWIF("L3VFeEujGtevx9w18HD1fhRbCH67Az2dpCymeRE1SoPK6XQtaN2k")).PrivKey,
			Address:    "bc1q9vza2e8x573nczrlzms0wvx3gsqjx7vavgkx0l",
			Message:    "Hello World",
		},
		{
			PrivateKey: lo.Must(btcutil.DecodeWIF("L3VFeEujGtevx9w18HD1fhRbCH67Az2dpCymeRE1SoPK6XQtaN2k")).PrivKey,
			Address:    "bc1ppv609nr0vr25u07u95waq5lucwfm6tde4nydujnu8npg4q75mr5sxq8lt3",
			Message:    "",
		},
		{
			PrivateKey: lo.Must(btcutil.DecodeWIF("L3VFeEujGtevx9w18HD1fhRbCH67Az2dpCymeRE1SoPK6XQtaN2k")).PrivKey,
			Address:    "bc1ppv609nr0vr25u07u95waq5lucwfm6tde4nydujnu8npg4q75mr5sxq8lt3",
			Message:    "Hello World",
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s_%s", tc.Address, tc.Message), func(t *testing.T) {
			address, err := btcutils.SafeNewAddress(tc.Address)
			require.NoError(t, err)
			signature, err := SignMessage(tc.PrivateKey, &address, tc.Message)
			require.NoError(t, err)

			verified := VerifyMessage(&address, signature, tc.Message)
			assert.True(t, verified)
		})
	}
}
