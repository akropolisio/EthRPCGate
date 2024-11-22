package utils

import (
	"errors"
	"testing"
)

func TestConvertKaonAddress(t *testing.T) {
	// TODO: fix
	bech32addressMainnet := "qc1q3422djj7p4mjsgn7m3k3kymd2s36jnrpzcn7xx"
	bech32addressTestnet := "tq1qxagv83u8vgg656de4aa04xvxe7jfzguwmg020n"
	legacyaddressMainnet := "QYmyzKNjoox5LkaiUvibZdM252bftQotDx"
	legacyAddressTesnet := "ar2SzdHghSgeacypPn7zfDe3qfKAEwimus"

	var tests = []struct {
		address string
		want    string
		err     error
	}{
		{bech32addressMainnet, "", errors.New("invalid address")},
		{bech32addressTestnet, "", errors.New("invalid address")},
		{legacyaddressMainnet, "8585918c3ee7168ee9d79dd9b5883eb65d0e0db0", nil},
		{legacyAddressTesnet, "1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead", nil},
	}

	for _, tt := range tests {
		testname := tt.address
		t.Run(testname, func(t *testing.T) {
			got, _ := ConvertKaonAddress(tt.address)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}

}
