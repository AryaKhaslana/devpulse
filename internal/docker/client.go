package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type ContainerInfo struct {
	ID     string            `json:"Id"`
	Names  []string          `json:"Names"`
	Image  string            `json:"Image"`
	State  string            `json:"State"`  // "running", "exited", etc.
	Status string            `json:"Status"` // "Up 2 hours", "Exited (1)"
	Labels map[string]string `json:"Labels"`
}

func (c *ContainerInfo) CleanName() string {
	if len(c.Names) == 0 {
		if len(c.ID) >= 12 {
			return c.ID[:12]
		}
		return c.ID
	}
	return strings.TrimPrefix(c.Names[0], "/")
}

func (c *ContainerInfo) IsRunning() bool {
	return strings.EqualFold(c.State, "running")
}

type Client interface {
	ListContainers(ctx context.Context, all bool) ([]ContainerInfo, error)
	RestartContainer(ctx context.Context, nameOrID string) error
	GetContainerLogs(ctx context.Context, nameOrID string, tail int) (string, error)
}

type SocketClient struct {
	client *http.Client
}

func NewClient(socketPath string) *SocketClient {
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}

	return &SocketClient{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", socketPath)
				},
			},
			Timeout: 15 * time.Second,
		},
	}
}

func (c *SocketClient) ListContainers(ctx context.Context, all bool) ([]ContainerInfo, error) {
	url := "http://localhost/containers/json"
	if all {
		url += "?all=1"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("docker API returned %d: %s", resp.StatusCode, string(body))
	}

	var containers []ContainerInfo
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("failed to decode containers list: %w", err)
	}

	return containers, nil
}

func (c *SocketClient) RestartContainer(ctx context.Context, nameOrID string) error {
	nameOrID = strings.TrimPrefix(nameOrID, "/")
	url := fmt.Sprintf("http://localhost/containers/%s/restart?t=10", nameOrID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to restart container %s: %w", nameOrID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("docker restart failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *SocketClient) GetContainerLogs(ctx context.Context, nameOrID string, tail int) (string, error) {
	if tail <= 0 {
		tail = 10
	}
	nameOrID = strings.TrimPrefix(nameOrID, "/")
	url := fmt.Sprintf("http://localhost/containers/%s/logs?stdout=1&stderr=1&tail=%d", nameOrID, tail)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for container %s: %w", nameOrID, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return cleanDockerLogStream(body), nil
}

func cleanDockerLogStream(raw []byte) string {
	if len(raw) == 0 {
		return "(Log kosong)"
	}

	var sb strings.Builder
	for len(raw) > 8 {
		if raw[0] == 1 || raw[0] == 2 {
			size := int(raw[4])<<24 | int(raw[5])<<16 | int(raw[6])<<8 | int(raw[7])
			raw = raw[8:]
			if size > len(raw) {
				size = len(raw)
			}
			sb.Write(raw[:size])
			raw = raw[size:]
		} else {
			sb.Write(raw)
			break
		}
	}
	res := strings.TrimSpace(sb.String())
	if res == "" {
		return strings.TrimSpace(string(raw))
	}
	return res
}
