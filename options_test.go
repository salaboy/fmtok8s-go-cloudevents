package server

import (
	"testing"

	ce "github.com/cloudevents/sdk-go/v2"
	"gotest.tools/v3/assert"
)

func TestOptions(t *testing.T) {
	proto, err := ce.NewHTTP()
	assert.NilError(t, err)

	ceClient, err := ce.NewClient(proto, ce.WithTimeNow(), ce.WithUUIDs())
	assert.NilError(t, err)

	testCases := []struct {
		name    string
		opt     Option
		wantErr string
	}{
		{
			name:    "empty sink",
			opt:     WithSink(""),
			wantErr: "sink empty",
		},
		{
			name:    "empty address",
			opt:     WithListenAddress(""),
			wantErr: "address empty",
		},
		{
			name:    "ce protocol nil",
			opt:     WithCEClient(ceClient, nil),
			wantErr: "one parameter is nil",
		},
		{
			name:    "ce client nil",
			opt:     WithCEClient(nil, proto),
			wantErr: "one parameter is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var srv Server
			err := tc.opt(&srv)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}
