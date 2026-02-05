package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNetAddress(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    NetAddress
		wantErr bool
	}{
		{
			name: "correct address",
			s:    "localhost:8080",
			want: NetAddress{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ParseNetAddress(tt.s)
			if tt.wantErr {
				assert.Error(t, gotErr, gotErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
