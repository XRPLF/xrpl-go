package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXChainBridge_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		bridge   XChainBridge
		expected FlatXChainBridge
	}{
		{
			name: "pass - classic addresses",
			bridge: XChainBridge{
				IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				IssuingChainIssue: "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
				LockingChainDoor:  "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn",
				LockingChainIssue: "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
			},
			expected: FlatXChainBridge{
				"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"IssuingChainIssue": "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
				"LockingChainDoor":  "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn",
				"LockingChainIssue": "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
			},
		},
		{
			name: "pass - X-addresses flatten to classic addresses",
			bridge: XChainBridge{
				IssuingChainDoor:  "XVPcpSm47b1CZkf5AkKM9a84dQHe3m4sBhsrA4XtnBECTAc",
				IssuingChainIssue: "XVjHFoA9fDUh732JG1RsjLV9cLJosjfMoknkGELWhQ76Eyn",
				LockingChainDoor:  "X7tR8jTwG7vJXotpo2oV43qtSxQXTSqrh7MjigzRZyqVvVr",
				LockingChainIssue: "XVjHFoA9fDUh732JG1RsjLV9cLJosjfMoknkGELWhQ76Eyn",
			},
			expected: FlatXChainBridge{
				"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"IssuingChainIssue": "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
				"LockingChainDoor":  "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn",
				"LockingChainIssue": "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.bridge.Flatten())
		})
	}
}
