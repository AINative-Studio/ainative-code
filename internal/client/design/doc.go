// Package design provides a client for interacting with the AINative Design API.
//
// The design client enables uploading, retrieving, and managing design tokens
// that represent design decisions as data (colors, typography, spacing, etc.).
//
// Example usage:
//
//	// Create a design client
//	designClient := design.New(
//		design.WithAPIClient(apiClient),
//		design.WithProjectID("my-project"),
//	)
//
//	// Upload tokens
//	tokens := []*design.Token{
//		{
//			Name:  "primary-color",
//			Value: "#007bff",
//			Type:  "color",
//			Category: "colors",
//		},
//	}
//
//	result, err := designClient.UploadTokens(ctx, tokens, design.ConflictOverwrite, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Uploaded %d tokens\n", result.UploadedCount)
package design
