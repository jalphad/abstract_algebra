package gfpoly

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jalphad/abstract_algebra/exercises/3-gfpn"
	"github.com/jalphad/abstract_algebra/testcases/gfpoly"

	v1 "github.com/jalphad/testforge/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGFPolyImplementation(t *testing.T) {
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
		Id: "gfpoly-operations",
	})
	if err != nil {
		t.Fatalf("Failed to get test case: %v", err)
	}

	// Parse the input
	var input gfpoly.GFPolyTestInput
	err = json.Unmarshal(getResp.Data.Input, &input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	t.Logf("Testing with GF(%d^%d) and %d polynomial operations", input.Prime, input.Degree, len(input.Operations))

	// Create a field with the given parameters
	field, err := gfpn.NewField(input.Prime, input.Degree, input.IrreducibleCoeffs)
	if err != nil {
		t.Fatalf("Failed to create field: %v", err)
	}

	// Execute the operations
	results := make([]gfpoly.PolyResult, len(input.Operations))
	for i, op := range input.Operations {
		// Convert coefficient indices to elements
		poly1 := indicesToPoly(field, op.Poly1)
		poly2 := indicesToPoly(field, op.Poly2)

		var result gfpoly.PolyResult

		switch op.Op {
		case "add":
			p := Add(poly1, poly2)
			result.Polynomial = polyToStrings(p)

		case "sub":
			p := Subtract(poly1, poly2)
			result.Polynomial = polyToStrings(p)

		case "mul":
			p := Multiply(poly1, poly2)
			result.Polynomial = polyToStrings(p)

		case "scalar_mul":
			scalar := field.Element(op.Scalar)
			p := ScalarMultiply(scalar, poly1)
			result.Polynomial = polyToStrings(p)

		case "derivative":
			p := FormalDerivative(poly1)
			result.Polynomial = polyToStrings(p)

		case "divide":
			q, r := Divide(poly1, poly2)
			result.Quotient = polyToStrings(q)
			result.Remainder = polyToStrings(r)

		case "eval":
			point := field.Element(op.Point)
			value := poly1.Evaluate(point)
			result.Value = value.String()

		default:
			t.Fatalf("Unknown operation: %s", op.Op)
		}

		results[i] = result
	}

	// Create response
	response := gfpoly.GFPolyTestResponse{
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
		t.Errorf("GF(p^n) polynomial implementation validation failed: %s", submitResp.Message)
	}
}

// Helper functions

// indicesToPoly converts a slice of element indices to a polynomial
func indicesToPoly(field gfpn.Field, indices []int) Polynomial {
	if len(indices) == 0 {
		return NewPolynomial(field, []gfpn.Element{})
	}

	elements := make([]gfpn.Element, len(indices))
	for i, idx := range indices {
		elements[i] = field.Element(idx)
	}

	return NewPolynomial(field, elements)
}

// polyToStrings converts a polynomial to a slice of coefficient strings
func polyToStrings(p Polynomial) []string {
	coeffs := p.Coefficients()
	if len(coeffs) == 0 {
		return []string{}
	}

	result := make([]string, len(coeffs))
	for i, c := range coeffs {
		result[i] = c.String()
	}
	return result
}
