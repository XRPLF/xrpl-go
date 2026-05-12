package wallet

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestNewWalletFromSeed(t *testing.T) {
	testCases := []struct {
		Seed           string
		PublicKey      string
		PrivateKey     string
		ClassicAddress types.Address
		MasterAddress  types.Address
	}{
		{
			Seed:           "sEd7io6yt5dFJrcePgRiFVHvmkJhJD1",
			PublicKey:      "EDC9DA1AA7513D891B58B3C9BEBAE3EB12620AFF4ABBA806B23BB3FA62109CE87F",
			PrivateKey:     "EDE01A1644C9FDE0367A7A285CA69798066C131C1133E1128B170CA65AEA5C6D19",
			ClassicAddress: "rn5M6BQCmQAzBxms9A84qEpx1Fdn9y7jdD",
		},
		{
			Seed:           "sEdTLE1G6QVc8znymeRZD3s5oajQcY5",
			PublicKey:      "ED676AD70576E126B46F6AF52D908FAB8F352F3A0BA05F48613DF017F6B83205B6",
			PrivateKey:     "ED31053AC2D97A74EA3A401CF23EEDA20400500CCFB82F442A8E0E6096A4150A9F",
			ClassicAddress: "r9D79PwpT5Z5xztgiQQgmxcYbF249PefnW",
		},
		{
			Seed:           "sEd71D6u2LkA36TMfJ5rApVsgZXQE9F",
			PublicKey:      "ED0A8A14F3226B2109047662605898F96F61764A9269B7823453A04A7B4F524C0E",
			PrivateKey:     "ED2C0EAB27E1411DBB8FACC88D531A69967DA0E45AC7821A4041A5AEE24BB8FF29",
			ClassicAddress: "rs7cvHcsEF54DEs2y24Tpph3Xf71xUUrFu",
		},
		{
			Seed:          "sh8i92YRnEjJy3fpFkL8txQSCVo79",
			PublicKey:     "03AEEFE1E8ED4BBC009DE996AC03A8C6B5713B1554794056C66E5B8D1753C7DD0E",
			PrivateKey:    "004265A28F3E18340A490421D47B2EB8DBC2C0BF2C24CEFEA971B61CED2CABD233",
			MasterAddress: "rUAi7pipxGpYfPNg3LtPcf2ApiS8aw9A93",
		},
	}

	for _, tc := range testCases {
		wallet, err := FromSeed(tc.Seed, tc.MasterAddress.String())
		if err != nil {
			t.Errorf("Error generating wallet from seed: %s", err)
		}

		if wallet.PublicKey != tc.PublicKey {
			t.Errorf("Public key does not match expected value. Expected: %s, got: %s", tc.PublicKey, wallet.PublicKey)
		}

		if wallet.PrivateKey != tc.PrivateKey {
			t.Errorf("Private key does not match expected value. Expected: %s, got: %s", tc.PrivateKey, wallet.PrivateKey)
		}

		if tc.MasterAddress != "" {
			if wallet.ClassicAddress != tc.MasterAddress {
				t.Errorf("Classic address does not match expected value. Expected: %s, got: %s", tc.MasterAddress, wallet.ClassicAddress)
			}
		} else {
			if wallet.ClassicAddress != tc.ClassicAddress {
				t.Errorf("Classic address does not match expected value. Expected: %s, got: %s", tc.ClassicAddress, wallet.ClassicAddress)
			}
		}
	}
}

func TestNewWalletFromSecret(t *testing.T) {
	testCases := []struct {
		Seed           string
		PublicKey      string
		PrivateKey     string
		ClassicAddress types.Address
	}{
		{
			Seed:           "sEd7io6yt5dFJrcePgRiFVHvmkJhJD1",
			PublicKey:      "EDC9DA1AA7513D891B58B3C9BEBAE3EB12620AFF4ABBA806B23BB3FA62109CE87F",
			PrivateKey:     "EDE01A1644C9FDE0367A7A285CA69798066C131C1133E1128B170CA65AEA5C6D19",
			ClassicAddress: "rn5M6BQCmQAzBxms9A84qEpx1Fdn9y7jdD",
		},
		{
			Seed:           "sEdTLE1G6QVc8znymeRZD3s5oajQcY5",
			PublicKey:      "ED676AD70576E126B46F6AF52D908FAB8F352F3A0BA05F48613DF017F6B83205B6",
			PrivateKey:     "ED31053AC2D97A74EA3A401CF23EEDA20400500CCFB82F442A8E0E6096A4150A9F",
			ClassicAddress: "r9D79PwpT5Z5xztgiQQgmxcYbF249PefnW",
		},
		{
			Seed:           "sEd71D6u2LkA36TMfJ5rApVsgZXQE9F",
			PublicKey:      "ED0A8A14F3226B2109047662605898F96F61764A9269B7823453A04A7B4F524C0E",
			PrivateKey:     "ED2C0EAB27E1411DBB8FACC88D531A69967DA0E45AC7821A4041A5AEE24BB8FF29",
			ClassicAddress: "rs7cvHcsEF54DEs2y24Tpph3Xf71xUUrFu",
		},
	}

	for _, tc := range testCases {
		wallet, err := FromSecret(tc.Seed)
		if err != nil {
			t.Errorf("Error generating wallet from seed: %s", err)
		}

		if wallet.PublicKey != tc.PublicKey {
			t.Errorf("Public key does not match expected value. Expected: %s, got: %s", tc.PublicKey, wallet.PublicKey)
		}

		if wallet.PrivateKey != tc.PrivateKey {
			t.Errorf("Private key does not match expected value. Expected: %s, got: %s", tc.PrivateKey, wallet.PrivateKey)
		}

		if wallet.ClassicAddress != tc.ClassicAddress {
			t.Errorf("Classic address does not match expected value. Expected: %s, got: %s", tc.ClassicAddress, wallet.ClassicAddress)
		}
	}
}

func TestNewWalletFromMnemonic(t *testing.T) {
	testCases := []struct {
		Mnemonic       string
		PublicKey      string
		PrivateKey     string
		ClassicAddress types.Address
	}{
		{
			Mnemonic:       "midnight help already frost arena force omit physical please dwarf envelope royal dice surge eight often muscle tired blast begin waste fat rescue debate",
			PublicKey:      "028E831F16FD85ABEDA7577B6F4F26500FAB80AEA54B8A89EEC6FA44BCC7AF5678",
			PrivateKey:     "00C503FC86436D384F37F946E8DE3B8D9B4D09961424B7ABEF47DEE229A499D557",
			ClassicAddress: "rpa9S5fRbS2ZAf2cdFGtezhGYgom1iD4yh",
		},
		{
			Mnemonic:       "honey tip lunch empower omit invite nuclear tent brother sadness still exercise odor harbor alcohol huge wait swamp vessel tired swallow supreme silk spawn",
			PublicKey:      "038D4EA460687B0FF43E95D9CB56E439AC2C65D890C9FD8B553FF24997C5F91D9A",
			PrivateKey:     "00C2C6F996F5FAD5168CAAAD6E58E518641A41E8E164A6FADC540F52D728E2A4AB",
			ClassicAddress: "rPdkbYi6ok7HFpbRo6CSxeDUg951tVdA1m",
		},
		{
			Mnemonic:       "hen toe quarter robust elevator badge coconut all place desk pen school topic life seminar run salute paddle hurdle impact push amount oblige citizen",
			PublicKey:      "0388AE366DF0D8819760B82319C7A04CA06CC1D49EB5D41ED8DA0C8903DE8FF812",
			PrivateKey:     "001CBDDCA87FB57FB5A9D96ECE29D5710DA8421891F40CFA2262120DBB61E5050D",
			ClassicAddress: "rsKbuMTkzR5HU96j8pdGsSuzmZEiZ6mKh5",
		},
	}

	for _, tc := range testCases {
		wallet, err := FromMnemonic(tc.Mnemonic)
		if err != nil {
			t.Errorf("Error generating wallet from mnemonic: %s", err)
		}

		if wallet.PublicKey != tc.PublicKey {
			t.Errorf("Public key does not match expected value. Expected: %s, got: %s", tc.PublicKey, wallet.PublicKey)
		}

		if wallet.PrivateKey != tc.PrivateKey {
			t.Errorf("Private key does not match expected value. Expected: %s, got: %s", tc.PrivateKey, wallet.PrivateKey)
		}

		if wallet.ClassicAddress != tc.ClassicAddress {
			t.Errorf("Classic address does not match expected value. Expected: %s, got: %s", tc.ClassicAddress, wallet.ClassicAddress)
		}
	}
}

func TestSign(t *testing.T) {
	testCases := []struct {
		name           string
		wallet         *Wallet
		tx             map[string]any
		expectedTxBlob string
		expectedHash   string
	}{
		{
			name: "Test Sign with Wallet 1",
			wallet: &Wallet{
				PublicKey:      "EDE5638D8055CCD45EBF7F5FFD59FC1703D6BC00800BBA19F158119DAA1A52A8D5",
				PrivateKey:     "ED0A961B472E78B89F1AE6A7CC4FB55FD083B36661D3D124E1BA29998346AE1AA1",
				ClassicAddress: "raJB6EHNSJa3jV7FqWNrAhcL6FEDE3PGc5",
			},
			tx: map[string]any{
				"Account":         "raJB6EHNSJa3jV7FqWNrAhcL6FEDE3PGc5",
				"TransactionType": "Payment",
				"Amount":          "15",
				"Destination":     "rDwvihpE48E48F8rvNrqTb2UGWv62xqYTg",
				"Flags":           uint32(0),
				"Fee":             "12",
				"Sequence":        uint32(1798962),
				"SigningPubKey":   "028E831F16FD85ABEDA7577B6F4F26500FAB80AEA54B8A89EEC6FA44BCC7AF5678",
			},
			expectedTxBlob: "120000220000000024001B733261400000000000000F68400000000000000C7321EDE5638D8055CCD45EBF7F5FFD59FC1703D6BC00800BBA19F158119DAA1A52A8D57440A973391D589C1D81E55516420A8D095DD98D2FC1F85E53C427EEEC22C6D3DEBADFA184005F5539E6A672CC4FA468125981584DDCE9365A6C7076F2E9CAF86B0E81143A18A088CF12B2D3E51F47A75D2A9859EF61ECA78314858233827B488ECB8D0EB940E7AC85CE41E343CF",
			expectedHash:   "37186C50D0A3FAB1218B5F7DAA235E4592F5080AF1C81F9B2678D4751C103CDF",
		},
		{
			name: "Test Sign with Wallet 2",
			wallet: &Wallet{
				PublicKey:      "ED839AE597DD34FDD0A806CE7690A7F0A753CAEEAC7A0B1DE1EF6EC647DD3CBC6D",
				PrivateKey:     "ED3841062F784C9F890249D736A0A36424497B0C53679A346B88A147F09B30CB9F",
				ClassicAddress: "rLY96NyP8Wq5yX5NQ3XdeZdUyUFBRbWNgd",
			},
			tx: map[string]any{
				"Account":         "rLY96NyP8Wq5yX5NQ3XdeZdUyUFBRbWNgd",
				"TransactionType": "Payment",
				"Amount":          "15",
				"Destination":     "rDwvihpE48E48F8rvNrqTb2UGWv62xqYTg",
				"Flags":           uint32(0),
				"Fee":             "12",
				"Sequence":        uint32(1798962),
				"SigningPubKey":   "028E831F16FD85ABEDA7577B6F4F26500FAB80AEA54B8A89EEC6FA44BCC7AF5678",
			},
			expectedTxBlob: "120000220000000024001B733261400000000000000F68400000000000000C7321ED839AE597DD34FDD0A806CE7690A7F0A753CAEEAC7A0B1DE1EF6EC647DD3CBC6D7440BAD74EEF422A1F8DDB28E005ECC3A19E5678EA36349722E1B8F7B76528A81D22E1CB0F0C15593613A3970B1D95CBB901C7C0F701381AD70C50FA8411778254078114D64E8ABC22DA5D143D60AEA083E17D3508DEE19C8314858233827B488ECB8D0EB940E7AC85CE41E343CF",
			expectedHash:   "94DB52A478D965C73A17427593004FD14FF74853DD667D8F3E8BF2A1FE66A11F",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txBlob, hash, err := tc.wallet.Sign(tc.tx)
			if err != nil {
				t.Fatalf("Error signing transaction: %v", err)
			}

			if txBlob == "" {
				t.Error("Expected non-empty txBlob, got empty string")
			}

			if hash == "" {
				t.Error("Expected non-empty hash, got empty string")
			}
		})
	}
}

func TestEnsureClassicAddress(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    types.Address
		expectedErr error
	}{
		{
			name:     "pass - classic address is returned unchanged",
			input:    "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			expected: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		},
		{
			name:     "pass - x-address without tag is converted to classic",
			input:    "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
			expected: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		},
		{
			name:        "fail - x-address with explicit zero tag is rejected",
			input:       "XV5sbjUmgPpvXv4ixFWZ5ptAYZ6PD2m4Er6SnvjVLpMWPjR",
			expectedErr: ErrAddressTagNotZero,
		},
		{
			name:        "fail - x-address with non-zero tag is rejected",
			input:       "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGZMhc9YTE92ehJ2Fu",
			expectedErr: ErrAddressTagNotZero,
		},
		{
			name:     "pass - non-x-address string is returned unchanged",
			input:    "not-a-valid-address",
			expected: "not-a-valid-address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ensureClassicAddress(tc.input)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}
