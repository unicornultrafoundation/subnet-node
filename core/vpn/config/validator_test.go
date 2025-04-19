package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *VPNConfig
		wantErr error
	}{
		{
			name: "valid config",
			config: &VPNConfig{
				Enable:    true,
				VirtualIP: "10.0.0.1",
				Subnet:    24,
				Routes:    []string{"10.0.0.0/24"},
				MTU:       1500,
			},
			wantErr: nil,
		},
		{
			name: "disabled config",
			config: &VPNConfig{
				Enable: false,
			},
			wantErr: nil,
		},
		{
			name: "missing virtual IP",
			config: &VPNConfig{
				Enable: true,
				Subnet: 24,
				Routes: []string{"10.0.0.0/24"},
				MTU:    1500,
			},
			wantErr: ErrVirtualIPNotSet,
		},
		{
			name: "invalid virtual IP",
			config: &VPNConfig{
				Enable:    true,
				VirtualIP: "invalid",
				Subnet:    24,
				Routes:    []string{"10.0.0.0/24"},
				MTU:       1500,
			},
			wantErr: ErrInvalidVirtualIP,
		},
		{
			name: "missing routes",
			config: &VPNConfig{
				Enable:    true,
				VirtualIP: "10.0.0.1",
				Subnet:    24,
				Routes:    []string{},
				MTU:       1500,
			},
			wantErr: ErrRoutesNotSet,
		},
		{
			name: "invalid routes",
			config: &VPNConfig{
				Enable:    true,
				VirtualIP: "10.0.0.1",
				Subnet:    24,
				Routes:    []string{"invalid"},
				MTU:       1500,
			},
			wantErr: ErrInvalidRoutes,
		},
		{
			name: "MTU too small",
			config: &VPNConfig{
				Enable:    true,
				VirtualIP: "10.0.0.1",
				Subnet:    24,
				Routes:    []string{"10.0.0.0/24"},
				MTU:       500,
			},
			wantErr: ErrInvalidMTU,
		},
		{
			name: "MTU too large",
			config: &VPNConfig{
				Enable:    true,
				VirtualIP: "10.0.0.1",
				Subnet:    24,
				Routes:    []string{"10.0.0.0/24"},
				MTU:       10000,
			},
			wantErr: ErrInvalidMTU,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
