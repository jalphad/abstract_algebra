package gf

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jalphad/abstract_algebra/testcases/gf"

	v1 "github.com/jalphad/testforge/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestFieldImplementation(t *testing.T) {
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
		Id: "field-operations",
	})
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}

	// Parse the input
	var input gf.FieldTestInput
	err = json.Unmarshal(getResp.Data.Input, &input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	t.Logf("Testing with GF(%d) and %d operations", input.Prime, len(input.Operations))

	// Create a field with the given prime
	gfp := NewField(input.Prime)

	// Execute the operations
	results := make([]int16, len(input.Operations))
	for i, op := range input.Operations {
		// Create elements for the operation
		elem1 := gfp.Element(op.Arg1)
		elem2 := gfp.Element(op.Arg2)

		// Perform the operation
		var result Element
		switch op.Op {
		case "add":
			result = gfp.Add(elem1, elem2)
		case "sub":
			result = gfp.Sub(elem1, elem2)
		case "mul":
			result = gfp.Mul(elem1, elem2)
		case "div":
			result = gfp.Div(elem1, elem2)
		default:
			t.Fatalf("Unknown operation: %s", op.Op)
		}

		results[i] = result.Value()
	}

	// Create response
	response := gf.FieldTestResponse{
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
		t.Errorf("Field implementation validation failed: %s", submitResp.Message)
	}
}
