package forney

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
	"github.com/jalphad/abstract_algebra/exercises/4-gfpoly"
	"github.com/jalphad/abstract_algebra/testcases/forney"

	v1 "github.com/jalphad/testforge/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestForneyImplementation(t *testing.T) {
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
		Id: "forney-algorithm",
	})
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}

	// Parse the input
	var input forney.ForneyTestInput
	err = json.Unmarshal(getResp.Data.Input, &input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	t.Logf("Testing with GF(%d^%d), %d syndromes, Λ(x) degree %d, %d error positions",
		input.Prime, input.Degree, len(input.Syndromes), input.LambdaDegree, len(input.ErrorPositions))

	// Create field with the given parameters
	field, err := gfpn.NewField(input.Prime, input.Degree, input.IrreducibleCoeffs)
	if err != nil {
		t.Fatalf("Failed to create field: %v", err)
	}

	// Helper to parse element strings
	order := 1
	for i := 0; i < input.Degree; i++ {
		order *= int(input.Prime)
	}

	parseElement := func(s string) (gfpn.Element, error) {
		for j := 0; j < order; j++ {
			elem := field.Element(j)
			if elem.String() == s {
				return elem, nil
			}
		}
		return field.Zero(), err
	}

	// Parse syndromes
	syndromes := make([]gfpn.Element, len(input.Syndromes))
	for i, s := range input.Syndromes {
		elem, err := parseElement(s)
		if err != nil {
			t.Fatalf("Failed to parse syndrome %d: %s", i, s)
		}
		syndromes[i] = elem
	}

	// Parse lambda coefficients
	lambdaCoeffs := make([]gfpn.Element, len(input.LambdaCoeffs))
	for i, s := range input.LambdaCoeffs {
		elem, err := parseElement(s)
		if err != nil {
			t.Fatalf("Failed to parse lambda coefficient %d: %s", i, s)
		}
		lambdaCoeffs[i] = elem
	}

	// Create lambda polynomial
	lambda := gfpoly.NewPolynomial(field, lambdaCoeffs)

	// Compute omega using student implementation
	omega := ComputeOmega(field, syndromes, lambda)

	// Compute error magnitudes using student implementation
	magnitudes := ComputeErrorMagnitudes(field, lambda, omega, input.ErrorPositions)

	// Convert magnitudes to strings for response
	magnitudeStrings := make([]string, len(magnitudes))
	for i, mag := range magnitudes {
		magnitudeStrings[i] = mag.String()
	}

	// Create response
	response := forney.ForneyTestResponse{
		OmegaCoeffs:     omega.Coefficients(),
		ErrorMagnitudes: magnitudeStrings,
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
		t.Errorf("Forney algorithm implementation validation failed: %s", submitResp.Message)
	}
}
