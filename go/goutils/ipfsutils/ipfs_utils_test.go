package ipfsutils

import "testing"

func TestParseMultiAddrURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid url",
			url:     "/dns/ipfs/tcp/5001",
			want:    "http://ipfs:5001",
			wantErr: false,
		},
		{
			name:    "valid url with https scheme",
			url:     "/dns/ipfs/tcp/5001/https",
			want:    "https://ipfs:5001",
			wantErr: false,
		},
		{
			name:    "invalid url",
			url:     "//dns/ipfs/tcp/5001",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "/dns//ipfs/tcp/5001/https",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "/dns/ipfs//tcp/5001/https",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "/dns/ipfs/tcp//5001/https",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "/dns/ipfs/tcp/5001//https",
			want:    "",
			wantErr: true,
		},
		{
			name:    "valid url",
			url:     "https://ipfs:5001",
			want:    "https://ipfs:5001",
			wantErr: false,
		},
		{
			name:    "valid url",
			url:     "http://ipfs",
			want:    "http://ipfs",
			wantErr: false,
		},
		{
			name:    "invalid url",
			url:     "ssh://ipfs:5001",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "ipfs:5001",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
