package types

import "testing"

func TestBatchIDString(t *testing.T) {
	tests := []struct {
		name    string
		batchId BatchID
		want    string
	}{
		{
			name:    "Empty BatchID",
			batchId: BatchID(""),
			want:    "",
		},
		{
			name:    "Non-empty BatchID",
			batchId: BatchID("1234567890abcdef"),
			want:    "1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.batchId.String(); got != tt.want {
				t.Errorf("BatchID.String(), got: %v but we want %v", got, tt.want)
			}
		})
	}
}
