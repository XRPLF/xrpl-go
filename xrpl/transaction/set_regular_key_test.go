package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSetRegularKey_TxType(t *testing.T) {
	entry := &SetRegularKey{}
	assert.Equal(t, SetRegularKeyTx, entry.TxType())
}

func TestSetRegularKey_Flatten(t *testing.T) {
	tests := []struct {
		name       string
		regularKey *SetRegularKey
		want       string
	}{
		{
			name: "pass - valid SetRegularKey",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SetRegularKeyTx,
				},
				RegularKey: "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
			},
			want: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "SetRegularKey",
				"RegularKey":      "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v"
			}`,
		},
		{
			name: "pass - without RegularKey",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SetRegularKeyTx,
				},
			},
			want: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "SetRegularKey"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.regularKey.Flatten(), []byte(tt.want))
			if err != nil {
				t.Error(err)
			}
		})
	}
}
func TestSetRegularKey_Validate(t *testing.T) {
	tests := []struct {
		name       string
		regularKey *SetRegularKey
		wantValid  bool
		wantErr    bool
	}{
		{
			name: "pass - valid SetRegularKey",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SetRegularKeyTx,
				},
				RegularKey: "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid SetRegularKey with X-address",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SetRegularKeyTx,
				},
				RegularKey: "XVYRdEocC28DRx94ZFGP3qNJ1D5Ln7ecXFMd3vREB5Pesju",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid SetRegularKey BaseTx",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				RegularKey: "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - RegularKey same as Account",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SetRegularKeyTx,
				},
				RegularKey: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid RegularKey address",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SetRegularKeyTx,
				},
				RegularKey: "invalidAddress",
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "pass - without RegularKey",
			regularKey: &SetRegularKey{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SetRegularKeyTx,
				},
			},
			wantValid: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.regularKey.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("Validate() = %v, want %v", valid, !tt.wantErr)
			}
		})
	}
}
