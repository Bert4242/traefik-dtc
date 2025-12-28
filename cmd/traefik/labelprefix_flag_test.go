package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/traefik/v3/cmd"
)

func TestLabelPrefixFlagDecode(t *testing.T) {
	cfg := cmd.NewTraefikConfiguration()

	err := flag.Decode([]string{"--providers.docker.labelPrefix=custom."}, &cfg)
	require.NoError(t, err)

	if cfg.Configuration.Providers.Docker == nil {
		t.Fatalf("expected docker provider to be initialized")
	}

	require.Equal(t, "custom.", cfg.Configuration.Providers.Docker.LabelPrefix)
}
