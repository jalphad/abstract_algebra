package syndrome

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
	"github.com/jalphad/abstract_algebra/testcases/syndrome"

	v1 "github.com/jalphad/testforge/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSyndromeImplementation(t *testing.T) {
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
		Id: "syndrome-calculation",
	})
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}

	// Parse the input
	var input syndrome.SyndromeTestInput
	err = json.Unmarshal(getResp.Data.Input, &input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	t.Logf("Testing with GF(%d^%d), codeword length: %d, EC symbols: %d",
		input.Prime, input.Degree, len(input.Received), input.NumECSymbols)

	// Create field with the given parameters
	field, err := gfpn.NewField(input.Prime, input.Degree, input.IrreducibleCoeffs)
	if err != nil {
		t.Fatalf("Failed to create field: %v", err)
	}

	// Get generator root
	generatorRoot := field.Element(input.GeneratorRootIdx)

	// Calculate syndromes using student implementation
	syndromes := CalculateSyndromes(field, input.Received, input.NumECSymbols, generatorRoot)

	// Convert syndromes to strings
	syndromeStrings := make([]string, len(syndromes))
	for i, s := range syndromes {
		syndromeStrings[i] = s.String()
	}

	// Check if errors exist
	hasErrors := HasErrors(syndromes)

	// Create response
	response := syndrome.SyndromeTestResponse{
		Syndromes: syndromeStrings,
		HasErrors: hasErrors,
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
		t.Errorf("Syndrome calculation implementation validation failed: %s", submitResp.Message)
	}
}
