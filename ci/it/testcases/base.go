// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package testcases

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/testcontainers/testcontainers-go"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Base structs/interfaces for integration tests

type TestCase interface {
	SetupContainers(ctx context.Context) error
	RunTests(ctx context.Context, t *testing.T) error
	Cleanup(ctx context.Context, t *testing.T)
}

type IntegrationTestcaseBase struct {
	ConfigTemplate string
	Containers     *Containers
	alreadyPrinted bool
}

func (tc *IntegrationTestcaseBase) SetupContainers(ctx context.Context) error {
	return nil
}

func (tc *IntegrationTestcaseBase) RunTests(ctx context.Context, t *testing.T) error {
	return nil
}

func (tc *IntegrationTestcaseBase) Cleanup(ctx context.Context, t *testing.T) {
	if tc.Containers != nil {
		tc.Containers.Cleanup(ctx, t)
	}
}

func (tc *IntegrationTestcaseBase) maybePrint() {
	if tc.alreadyPrinted {
		return
	}
	tc.alreadyPrinted = true

	ctx := context.Background()

	// Print Docker container info for debugging
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running docker ps: %v\n", err)
	} else {
		fmt.Printf("Docker containers:\n%s\n", output)
	}

	// Print all Docker networks info
	cmd = exec.Command("docker", "network", "ls")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error listing docker networks: %v\n", err)
	} else {
		fmt.Printf("Docker networks:\n%s\n", output)

		// Inspect each network
		networks := strings.Split(strings.TrimSpace(string(output)), "\n")[1:] // Skip header row
		for _, network := range networks {
			networkID := strings.Fields(network)[0]
			cmd = exec.Command("docker", "network", "inspect", networkID)
			inspectOutput, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error inspecting network %s: %v\n", networkID, err)
			} else {
				fmt.Printf("Network %s config:\n%s\n", networkID, inspectOutput)
			}
		}
	}

	// Print networking info for each container
	containers := []struct {
		name      string
		container *testcontainers.Container
	}{
		{"Kibana", tc.Containers.Kibana},
		{"Quesma", tc.Containers.Quesma},
		{"Elasticsearch", tc.Containers.Elasticsearch},
		{"ClickHouse", tc.Containers.ClickHouse},
	}

	for _, c := range containers {
		if c.container != nil {
			containerID := (*c.container).GetContainerID()
			cmd = exec.Command("docker", "inspect", containerID)
			output, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error inspecting network settings for %s: %v\n", c.name, err)
			} else {
				fmt.Printf("%s container network settings:\n%s\n", c.name, output)
			}

			ips, _ := (*c.container).ContainerIPs(ctx)
			ports, _ := (*c.container).Ports(ctx)
			inspect, _ := (*c.container).Inspect(ctx)

			fmt.Printf("%s container IPs: %v\n", c.name, ips)
			fmt.Printf("%s container ports: %v\n", c.name, ports)
			fmt.Printf("%s container inspect: %v\n", c.name, inspect)
		}
	}
}

func (tc *IntegrationTestcaseBase) getKibanaEndpoint() string {
	tc.maybePrint()
	ctx := context.Background()
	q := *tc.Containers.Kibana
	p, err1 := q.MappedPort(ctx, "5601/tcp")
	h, err2 := q.Host(ctx)
	fmt.Printf("Kibana host: %s, port: %s, err1: %v, err2: %v\n", h, p.Port(), err1, err2)

	return fmt.Sprintf("http://%s:%s", h, p.Port())
}

func (tc *IntegrationTestcaseBase) getQuesmaEndpoint() string {
	tc.maybePrint()
	ctx := context.Background()
	q := *tc.Containers.Quesma
	p, err1 := q.MappedPort(ctx, "8080/tcp")
	h, err2 := q.Host(ctx)
	fmt.Printf("Quesma host: %s, port: %s, err1: %v, err2: %v\n", h, p.Port(), err1, err2)
	tc.getKibanaEndpoint()
	return fmt.Sprintf("http://%s:%s", h, p.Port())
}

func (tc *IntegrationTestcaseBase) getElasticsearchEndpoint() string {
	tc.maybePrint()
	ctx := context.Background()
	q := *tc.Containers.Elasticsearch
	p, err1 := q.MappedPort(ctx, "9200/tcp")
	h, err2 := q.Host(ctx)
	fmt.Printf("Elasticsearch host: %s, port: %s, err1: %v, err2: %v\n", h, p.Port(), err1, err2)
	tc.getKibanaEndpoint()
	return fmt.Sprintf("http://%s:%s", h, p.Port())
}

func (tc *IntegrationTestcaseBase) getClickHouseClient() (*sql.DB, error) {
	tc.maybePrint()
	ctx := context.Background()
	q := *tc.Containers.ClickHouse
	p, err1 := q.MappedPort(ctx, "9000/tcp")
	h, err2 := q.Host(ctx)
	fmt.Printf("ClickHouse host: %s, port: %s, err1: %v, err2: %v\n", h, p.Port(), err1, err2)
	tc.getElasticsearchEndpoint()
	tc.getQuesmaEndpoint()
	options := clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", h, p.Port())},
		TLS:  nil,
		Auth: clickhouse.Auth{
			Username: "default", // Replace with your ClickHouse username
			Password: "",        // Replace with your ClickHouse password, if any
		},
	}
	db := clickhouse.OpenDB(&options)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}
	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(30)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return db, nil
}

func (tc *IntegrationTestcaseBase) ExecuteClickHouseQuery(ctx context.Context, query string) (*sql.Rows, error) {
	db, err := tc.getClickHouseClient()
	if err != nil {
		return nil, err
	}
	if errP := db.Ping(); errP != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", errP)
	}
	rows, errQ := db.QueryContext(ctx, query)

	if errQ != nil {
		return nil, errQ
	}
	defer db.Close()
	return rows, nil
}

func (tc *IntegrationTestcaseBase) ExecuteClickHouseStatement(ctx context.Context, stmt string) (sql.Result, error) {
	db, err := tc.getClickHouseClient()
	if err != nil {
		return nil, err
	}
	if errP := db.Ping(); errP != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", errP)
	}
	res, errQ := db.ExecContext(ctx, stmt)

	if errQ != nil {
		return nil, errQ
	}
	defer db.Close()
	return res, nil
}

func (tc *IntegrationTestcaseBase) FetchClickHouseColumns(ctx context.Context, tableName string) (map[string]string, error) {
	rows, err := tc.ExecuteClickHouseQuery(ctx, fmt.Sprintf("SELECT name, type FROM system.columns WHERE table = '%s'", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var name, colType string
		if err := rows.Scan(&name, &colType); err != nil {
			return nil, err
		}
		result[name] = colType
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (tc *IntegrationTestcaseBase) FetchClickHouseComments(ctx context.Context, tableName string) (map[string]string, error) {
	rows, err := tc.ExecuteClickHouseQuery(ctx, fmt.Sprintf("SELECT name, comment FROM system.columns WHERE table = '%s'", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var name, comment string
		if err := rows.Scan(&name, &comment); err != nil {
			return nil, err
		}
		result[name] = comment
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (tc *IntegrationTestcaseBase) RequestToQuesma(ctx context.Context, t *testing.T, method, uri string, requestBody []byte) (*http.Response, []byte) {
	endpoint := tc.getQuesmaEndpoint()
	resp, err := tc.doRequest(ctx, method, endpoint+uri, requestBody, nil)
	if err != nil {
		t.Fatalf("Error sending %s request to the endpoint '%s': %s", method, uri, err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body of %s request to the endpoint '%s': %s", method, uri, err)
	}
	return resp, responseBody
}

func (tc *IntegrationTestcaseBase) RequestToElasticsearch(ctx context.Context, method, uri string, body []byte) (*http.Response, error) {
	endpoint := tc.getElasticsearchEndpoint()
	return tc.doRequest(ctx, method, endpoint+uri, body, nil)
}

func (tc *IntegrationTestcaseBase) doRequest(ctx context.Context, method, endpoint string, body []byte, headers http.Header) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("elastic", "quesmaquesma")
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}
