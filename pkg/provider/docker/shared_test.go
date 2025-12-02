package docker

import (
	"strconv"
	"testing"

	containertypes "github.com/moby/moby/api/types/container"
	networktypes "github.com/moby/moby/api/types/network"
	swarmtypes "github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedNormalizeLabels(t *testing.T) {
	t.Parallel()

	shared := Shared{LabelPrefix: "custom."}

	labels := map[string]string{
		DefaultLabelPrefix + "enable": "false",
		"custom.enable":               "true",
		"custom.docker.network":       "mynet",
		"unrelated":                   "value",
	}

	normalized := shared.normalizeLabels(labels)

	assert.Equal(t, map[string]string{
		"traefik.enable":         "true",
		"traefik.docker.network": "mynet",
		"unrelated":              "value",
	}, normalized)
}

func TestSharedNormalizeLabelsKeepDefault(t *testing.T) {
	t.Parallel()

	shared := Shared{LabelPrefix: "custom.", KeepDefaultLabelPrefixLabelAsDefault: true}

	labels := map[string]string{
		DefaultLabelPrefix + "enable": "false",
		"custom.enable":               "true",
		"custom.docker.network":       "mynet",
	}

	normalized := shared.normalizeLabels(labels)

	assert.Equal(t, map[string]string{
		"traefik.enable":         "true",
		"traefik.docker.network": "mynet",
	}, normalized)
}

func TestExtractDockerLabelsWithCustomPrefix(t *testing.T) {
	t.Parallel()

	shared := Shared{ExposedByDefault: true, Network: "default", LabelPrefix: "custom."}
	container := dockerData{Labels: map[string]string{"custom.enable": "false", "custom.docker.network": "customNet"}}

	conf, err := shared.extractDockerLabels(container)
	require.NoError(t, err)

	assert.False(t, conf.Enable)
	assert.Equal(t, "customNet", conf.Network)
}

func Test_getPort_docker(t *testing.T) {
	testCases := []struct {
		desc       string
		container  containertypes.InspectResponse
		serverPort string
		expected   string
	}{
		{
			desc:      "no binding, no server port label",
			container: containerJSON(name("foo")),
			expected:  "",
		},
		{
			desc: "binding, no server port label",
			container: containerJSON(ports(networktypes.PortMap{
				networktypes.MustParsePort("80/tcp"): {},
			})),
			expected: "80",
		},
		{
			desc: "binding, multiple ports, no server port label",
			container: containerJSON(ports(networktypes.PortMap{
				networktypes.MustParsePort("80/tcp"):  {},
				networktypes.MustParsePort("443/tcp"): {},
			})),
			expected: "80",
		},
		{
			desc:       "no binding, server port label",
			container:  containerJSON(),
			serverPort: "8080",
			expected:   "8080",
		},
		{
			desc: "binding, server port label",
			container: containerJSON(
				ports(networktypes.PortMap{
					networktypes.MustParsePort("80/tcp"): {},
				})),
			serverPort: "8080",
			expected:   "8080",
		},
		{
			desc: "binding, multiple ports, server port label",
			container: containerJSON(ports(networktypes.PortMap{
				networktypes.MustParsePort("8080/tcp"): {},
				networktypes.MustParsePort("80/tcp"):   {},
			})),
			serverPort: "8080",
			expected:   "8080",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getPort(dData, test.serverPort)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_getPort_swarm(t *testing.T) {
	testCases := []struct {
		service    swarmtypes.Service
		serverPort string
		networks   map[string]*networktypes.Summary
		expected   string
	}{
		{
			service: swarmService(
				withEndpointSpec(modeDNSRR),
			),
			networks:   map[string]*networktypes.Summary{},
			serverPort: "8080",
			expected:   "8080",
		},
	}

	for serviceID, test := range testCases {
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			var p SwarmProvider
			require.NoError(t, p.Init())

			dData, err := p.parseService(t.Context(), test.service, test.networks)
			require.NoError(t, err)

			actual := getPort(dData, test.serverPort)
			assert.Equal(t, test.expected, actual)
		})
	}
}
