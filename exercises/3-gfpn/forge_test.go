package gfpn

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jalphad/abstract_algebra/testcases/gfpn"

	v1 "github.com/jalphad/testforge/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGFPNImplementation(t *testing.T) {
	// Connect to the separately running TestForge server
	// Start the server first with: go run src/forge/main.go
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := v1.NewTestingServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the test case from the server
	getResp, err := client.GetTestCase(ctx, &v1.GetTestCaseRequest{
		Id: "gfpn-operations",
	})
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}

	// Parse the input
	var input gfpn.GFPNTestInput
	err = json.Unmarshal(getResp.Data.Input, &input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	t.Logf("Testing with GF(%d^%d) and %d operations", input.Prime, input.Degree, len(input.Operations))

	// Create a field with the given parameters
	gf, err := NewField(input.Prime, input.Degree, input.IrreducibleCoeffs)
	if err != nil {
		t.Fatalf("Failed to create field: %v", err)
	}

	// Execute the operations
	results := make([]string, len(input.Operations))
	for i, op := range input.Operations {
		// Create elements for the operation
		elem1 := gf.Element(op.Arg1)
		elem2 := gf.Element(op.Arg2)

		// Perform the operation
		var result Element
		switch op.Op {
		case "add":
			result = gf.Add(elem1, elem2)
		case "sub":
			result = gf.Sub(elem1, elem2)
		case "mul":
			result = gf.Mul(elem1, elem2)
		case "div":
			result = gf.Div(elem1, elem2)
		case "neg":
			result = gf.Sub(gf.Zero(), elem1)
		case "inv":
			result = gf.Div(gf.One(), elem1)
		default:
			t.Fatalf("Unknown operation: %s", op.Op)
		}

		results[i] = result.String()
	}

	// Create response
	response := gfpn.GFPNTestResponse{
		Results: results,
	}

	// Encode response
	responseBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Submit the solution to the server
	submitResp, err := client.SubmitSolution(ctx, &v1.SubmitSolutionRequest{
		TestCaseId: getResp.Data.Id,
		Response:   responseBytes,
		ClientId:   "student-test",
	})
	if err != nil {
		t.Fatalf("Failed to submit solution: %v", err)
	}

	// Check validation result
	t.Logf("Validation Result:")
	t.Logf("  Valid: %t", submitResp.Valid)
	t.Logf("  Score: %.1f", submitResp.Score)
	t.Logf("  Message: %s", submitResp.Message)

	// Test fails if validation result is not valid
	if !submitResp.Valid {
		t.Errorf("GF(p^n) implementation validation failed: %s", submitResp.Message)
	}
}
