package provider

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	docker "github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
)

func TestDockerGetFrontendName(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  []string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: []string{"Host-foo-docker-localhost"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
					},
				},
			},
			expected: []string{"Headers-User-Agent-bat-0-1-0"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
			},
			expected: []string{"Host-foo-bar"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/test",
					},
				},
			},
			expected: []string{"Path-test"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "PathPrefix:/test2",
					},
				},
			},
			expected: []string{"PathPrefix-test2"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "PathPrefix:/test2&&Host:foo.bar",
					},
				},
			},
			expected: []string{"Host-foo-bar", "PathPrefix-test2"},
		},
	}

	for _, e := range containers {
		actual := provider.getFrontendName(e.container)
		keys := []string{}
		for key := range actual {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		sort.Strings(e.expected)
		for index, key := range keys {
			if key != e.expected[index] {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		}
	}
}

func TestDockerGetFrontendRule(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  []string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: []string{"Host:foo.docker.localhost"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
			},
			expected: []string{"Host:bar.docker.localhost"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
			},
			expected: []string{"Host:foo.bar"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/test",
					},
				},
			},
			expected: []string{"Path:/test"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/test&&Host:foo.bar",
					},
				},
			},
			expected: []string{"Path:/test", "Host:foo.bar"},
		},
	}

	for _, e := range containers {
		actual := provider.getFrontendRule(e.container)
		sort.Strings(actual)
		sort.Strings(e.expected)
		for index, rule := range actual {
			if rule != e.expected[index] {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		}
	}
}

func TestDockerGetBackend(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "foo",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
			},
			expected: "bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.backend": "foobar",
					},
				},
			},
			expected: "foobar",
		},
	}

	for _, e := range containers {
		actual := provider.getBackend(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetIPAddress(t *testing.T) { // TODO
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet": {
							IPAddress: "10.11.12.13",
						},
					},
				},
			},
			expected: "10.11.12.13",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.docker.network": "testnet",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"nottestnet": {
							IPAddress: "10.11.12.13",
						},
					},
				},
			},
			expected: "10.11.12.13",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.docker.network": "testnet2",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet1": {
							IPAddress: "10.11.12.13",
						},
						"testnet2": {
							IPAddress: "10.11.12.14",
						},
					},
				},
			},
			expected: "10.11.12.14",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
					HostConfig: &container.HostConfig{
						NetworkMode: "host",
					},
				},
				Config: &container.Config{
					Labels: map[string]string{},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet1": {
							IPAddress: "10.11.12.13",
						},
						"testnet2": {
							IPAddress: "10.11.12.14",
						},
					},
				},
			},
			expected: "127.0.0.1",
		},
	}

	for _, e := range containers {
		actual := provider.getIPAddress(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetPort(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: "",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: "80",
		},
		// FIXME handle this better..
		// {
		// 	container: docker.ContainerJSON{
		// 		Name:   "bar",
		// 		Config: &container.Config{},
		// 		NetworkSettings: &docker.NetworkSettings{
		// 			Ports: map[docker.Port][]docker.PortBinding{
		// 				"80/tcp":  []docker.PortBinding{},
		// 				"443/tcp": []docker.PortBinding{},
		// 			},
		// 		},
		// 	},
		// 	expected: "80",
		// },
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: "8080",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: "8080",
		},
	}

	for _, e := range containers {
		actual := provider.getPort(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetWeight(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "1",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.weight": "10",
					},
				},
			},
			expected: "10",
		},
	}

	for _, e := range containers {
		actual := provider.getWeight(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetDomain(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "docker.localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.domain": "foo.bar",
					},
				},
			},
			expected: "foo.bar",
		},
	}

	for _, e := range containers {
		actual := provider.getDomain(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetProtocol(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "http",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.protocol": "https",
					},
				},
			},
			expected: "https",
		},
	}

	for _, e := range containers {
		actual := provider.getProtocol(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetPassHostHeader(t *testing.T) {
	provider := &Docker{}
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "true",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.passHostHeader": "false",
					},
				},
			},
			expected: "false",
		},
	}

	for _, e := range containers {
		actual := provider.getPassHostHeader(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetLabel(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				Config: &container.Config{},
			},
			expected: "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			expected: "",
		},
	}

	for _, e := range containers {
		label, err := getLabel(e.container, "foo")
		if e.expected != "" {
			if err == nil || !strings.Contains(err.Error(), e.expected) {
				t.Fatalf("expected an error with %q, got %v", e.expected, err)
			}
		} else {
			if label != "bar" {
				t.Fatalf("expected label 'bar', got %s", label)
			}
		}
	}
}

func TestDockerGetLabels(t *testing.T) {
	containers := []struct {
		container      docker.ContainerJSON
		expectedLabels map[string]string
		expectedError  string
	}{
		{
			container: docker.ContainerJSON{
				Config: &container.Config{},
			},
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "fooz",
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "Label not found: bar",
		},
		{
			container: docker.ContainerJSON{
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "fooz",
						"bar": "barz",
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
				"bar": "barz",
			},
			expectedError: "",
		},
	}

	for _, e := range containers {
		labels, err := getLabels(e.container, []string{"foo", "bar"})
		if !reflect.DeepEqual(labels, e.expectedLabels) {
			t.Fatalf("expect %v, got %v", e.expectedLabels, labels)
		}
		if e.expectedError != "" {
			if err == nil || !strings.Contains(err.Error(), e.expectedError) {
				t.Fatalf("expected an error with %q, got %v", e.expectedError, err)
			}
		}
	}
}

func TestDockerTraefikFilter(t *testing.T) {
	provider := Docker{}
	containers := []struct {
		container        docker.ContainerJSON
		exposedByDefault bool
		expected         bool
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			exposedByDefault: true,
			expected:         false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "false",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "80",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "anything",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: false,
			expected:         false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			exposedByDefault: false,
			expected:         true,
		},
	}

	for _, e := range containers {
		actual := provider.containerFilter(e.container, e.exposedByDefault)
		if actual != e.expected {
			t.Fatalf("expected %v for %+v, got %+v", e.expected, e, actual)
		}
	}
}

func TestDockerLoadDockerConfig(t *testing.T) {
	cases := []struct {
		containers        []docker.ContainerJSON
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			containers:        []docker.ContainerJSON{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			containers: []docker.ContainerJSON{
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test",
					},
					Config: &container.Config{},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost": {
							Rule: "Host:test.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test": {
							URL:    "http://127.0.0.1:80",
							Weight: 1,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test1",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.backend":              "foobar",
							"traefik.frontend.entryPoints": "http,https",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test2",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.backend": "foobar",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost": {
							Rule: "Host:test1.docker.localhost",
						},
					},
				},
				"frontend-Host-test2-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test2-docker-localhost": {
							Rule: "Host:test2.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL:    "http://127.0.0.1:80",
							Weight: 1,
						},
						"server-test2": {
							URL:    "http://127.0.0.1:80",
							Weight: 1,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
		},
	}

	provider := &Docker{
		Domain:           "docker.localhost",
		ExposedByDefault: true,
	}

	for _, c := range cases {
		actualConfig := provider.loadDockerConfig(c.containers)
		// Compare backends
		if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
			t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
		}
		if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
			t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
		}
	}
}
