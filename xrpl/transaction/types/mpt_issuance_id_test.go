package types

import "testing"

func TestMPTIssuanceIDString(t *testing.T) {
	tests := []struct {
		name          string
		mptIssuanceID MPTIssuanceID
		want          string
	}{
		{
			name:          "Empty MPTIssuanceID",
			mptIssuanceID: MPTIssuanceID(""),
			want:          "",
		},
		{
			name:          "Non-empty MPTIssuanceID",
			mptIssuanceID: MPTIssuanceID("1234567890abcdef"),
			want:          "1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mptIssuanceID.String(); got != tt.want {
				t.Errorf("MPTIssuanceID.String(), got: %v but we want %v", got, tt.want)
			}
		})
	}
}
