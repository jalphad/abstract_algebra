package berlekamp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
	"github.com/jalphad/abstract_algebra/testcases/berlekamp"

	v1 "github.com/jalphad/testforge/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestBerlekampMasseyImplementation(t *testing.T) {
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
		Id: "berlekamp-massey",
	})
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}

	// Parse the input
	var input berlekamp.BerlekampTestInput
	err = json.Unmarshal(getResp.Data.Input, &input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	t.Logf("Testing with GF(%d^%d), %d syndromes",
		input.Prime, input.Degree, len(input.Syndromes))

	// Create field with the given parameters
	field, err := gfpn.NewField(input.Prime, input.Degree, input.IrreducibleCoeffs)
	if err != nil {
		t.Fatalf("Failed to create field: %v", err)
	}

	// Parse syndromes from strings to elements
	syndromes := make([]gfpn.Element, len(input.Syndromes))
	order := 1
	for i := 0; i < input.Degree; i++ {
		order *= int(input.Prime)
	}

	for i, s := range input.Syndromes {
		found := false
		for j := 0; j < order; j++ {
			elem := field.Element(j)
			if elem.String() == s {
				syndromes[i] = elem
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Failed to parse syndrome: %s", s)
		}
	}

	// Run Berlekamp-Massey using student implementation
	lambda := BerlekampMassey(field, syndromes)

	// Convert to response format
	coeffs := lambda.Coefficients()
	errorLocator := make([]string, len(coeffs))
	for i, c := range coeffs {
		errorLocator[i] = c.String()
	}

	// Create response
	response := berlekamp.BerlekampTestResponse{
		ErrorLocator: errorLocator,
		Degree:       lambda.Degree(),
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
		t.Errorf("Berlekamp-Massey implementation validation failed: %s", submitResp.Message)
	}
}
