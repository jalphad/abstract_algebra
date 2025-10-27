package arithpoly

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jalphad/abstract_algebra/exercises/1-gf"
	"github.com/jalphad/abstract_algebra/testcases/arithpoly"

	v1 "github.com/jalphad/testforge/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPolyDivImplementation(t *testing.T) {
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
		Id: "poly-div",
	})
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}

	// Parse the input
	var input arithpoly.PolyDivTestInput
	err = json.Unmarshal(getResp.Data.Input, &input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	t.Logf("Testing polynomial division in GF(%d)", input.Prime)
	t.Logf("  Dividend: %v", input.Dividend)
	t.Logf("  Divisor:  %v", input.Divisor)

	// Create a field with the given prime
	field := gf.NewField(input.Prime)

	// Convert int16 slices to Polynomial (Element slices)
	dividend := valuesToPoly(field, input.Dividend)
	divisor := valuesToPoly(field, input.Divisor)

	// Perform polynomial division
	quotient, remainder := PolyDiv(field, dividend, divisor)

	// Convert results back to int16 slices
	quotientValues := polyToValues(quotient)
	remainderValues := polyToValues(remainder)

	t.Logf("  Quotient:  %v", quotientValues)
	t.Logf("  Remainder: %v", remainderValues)

	// Create response
	response := arithpoly.PolyDivTestResponse{
		Quotient:  quotientValues,
		Remainder: remainderValues,
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
		t.Errorf("Polynomial division validation failed: %s", submitResp.Message)
	}
}

// valuesToPoly converts int16 slice to Polynomial
func valuesToPoly(field gf.Field, values []int16) Polynomial {
	poly := make(Polynomial, len(values))
	for i, v := range values {
		poly[i] = field.Element(int(v))
	}
	return poly
}

// polyToValues converts Polynomial to int16 slice
func polyToValues(poly Polynomial) []int16 {
	values := make([]int16, len(poly))
	for i, elem := range poly {
		values[i] = elem.Value()
	}
	return values
}
