package cmd

import (
	"encoding/json"
	"fmt"
)

// zerodbOutputJSON outputs data as formatted JSON
func zerodbOutputJSON(data interface{}) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
