// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0

package it

import (
	"context"
	"quesma.com/its/testcases"
	"testing"
)

func runIntegrationTest(t *testing.T, testCase testcases.TestCase) {
	ctx := context.Background()
	t.Cleanup(func() {
		testCase.Cleanup(ctx, t)
	})
	if err := testCase.SetupContainers(ctx); err != nil {
		t.Fatalf("Failed to setup containers: %s", err)
	}
	if err := testCase.RunTests(ctx, t); err != nil {
		t.Fatalf("Failed to run tests: %s", err)
	}
}

func TestReadingClickHouseTablesIntegrationTestcase(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase2(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase3(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase4(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase5(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase6(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase7(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase8(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase9(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase10(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase11(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase12(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase13(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase14(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase15(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase16(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase17(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase18(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase19(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase20(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase21(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase22(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase23(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase24(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase25(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase26(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase27(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase28(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase29(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}

func TestReadingClickHouseTablesIntegrationTestcase30(t *testing.T) {
	testCase := testcases.NewReadingClickHouseTablesIntegrationTestcase()
	runIntegrationTest(t, testCase)
}
