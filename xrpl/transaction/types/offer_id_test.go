package types

import "testing"

func TestOfferIDString(t *testing.T) {
	tests := []struct {
		name    string
		offerID OfferID
		want    string
	}{
		{
			name:    "Empty OfferID",
			offerID: OfferID(""),
			want:    "",
		},
		{
			name:    "Non-empty OfferID",
			offerID: OfferID("1234567890abcdef"),
			want:    "1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.offerID.String(); got != tt.want {
				t.Errorf("OfferID.String(), got: %v but we want %v", got, tt.want)
			}
		})
	}
}
