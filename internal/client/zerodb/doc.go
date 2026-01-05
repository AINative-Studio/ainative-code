// Package zerodb provides a client for ZeroDB NoSQL table operations.
//
// The zerodb package implements:
//   - Table creation with schema definition
//   - Document insertion with validation
//   - MongoDB-style query filtering
//   - Document updates and deletions
//   - Table listing and management
//
// Basic Usage:
//
//	client := zerodb.New(
//	    zerodb.WithAPIClient(apiClient),
//	    zerodb.WithProjectID("my-project"),
//	)
//
//	// Create a table
//	err := client.CreateTable(ctx, "users", schema)
//
//	// Insert documents
//	id, err := client.Insert(ctx, "users", document)
//
//	// Query documents
//	results, err := client.Query(ctx, "users", filter)
//
//	// Update document
//	err = client.Update(ctx, "users", id, updates)
//
//	// Delete document
//	err = client.Delete(ctx, "users", id)
package zerodb
