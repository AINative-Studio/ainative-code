package cmd

import (
	"fmt"
	"reflect"
	"strings"
)

// sensitiveFieldPatterns contains patterns for identifying sensitive fields
var sensitiveFieldPatterns = []string{
	"api_key",
	"apikey",
	"api-key",
	"token",
	"secret",
	"password",
	"passwd",
	"pwd",
	"credential",
	"access_key",
	"secret_key",
	"session_token",
	"refresh_token",
	"client_secret",
	"encryption_key",
	"private_key",
	"auth_token",  // More specific than just "auth"
	"bearer",
}

// isSensitiveField checks if a field name indicates it contains sensitive data
func isSensitiveField(fieldName string) bool {
	lowerField := strings.ToLower(fieldName)

	// False positive exceptions - these are NOT sensitive even though they contain sensitive patterns
	exceptions := []string{
		"max_tokens",
		"num_tokens",
		"total_tokens",
		"completion_tokens",
		"prompt_tokens",
	}

	for _, exception := range exceptions {
		if lowerField == exception {
			return false
		}
	}

	// Special handling for token pattern to avoid false positives like "max_tokens"
	// We check for token only when it's part of specific sensitive patterns
	tokenSensitivePatterns := []string{
		"_token", "token_", "-token", "token-",
		"access_token", "refresh_token", "session_token", "bearer_token",
		"auth_token", "api_token", "id_token",
	}

	// Exact match for "token" is also sensitive
	if lowerField == "token" {
		return true
	}

	for _, pattern := range tokenSensitivePatterns {
		if strings.Contains(lowerField, pattern) {
			return true
		}
	}

	// Check other patterns (excluding the generic "token")
	for _, pattern := range sensitiveFieldPatterns {
		// Skip "token" as we handle it separately above
		if pattern == "token" {
			continue
		}
		if strings.Contains(lowerField, pattern) {
			return true
		}
	}
	return false
}

// maskValue masks a sensitive value, showing only first 3 and last 6 characters
// For values shorter than 15 characters, shows only the first 3 characters
func maskValue(value string) string {
	if value == "" {
		return ""
	}

	// Minimum length to show any characters
	const minLength = 4
	const prefixLen = 3
	const suffixLen = 6

	valueLen := len(value)

	// If value is too short, just show first few characters
	if valueLen < minLength {
		return strings.Repeat("*", valueLen)
	}

	// If value is short, show prefix only
	if valueLen < (prefixLen + suffixLen + 3) {
		prefix := value[:prefixLen]
		return fmt.Sprintf("%s%s", prefix, strings.Repeat("*", valueLen-prefixLen))
	}

	// For longer values, show prefix...suffix
	prefix := value[:prefixLen]
	suffix := value[valueLen-suffixLen:]
	return fmt.Sprintf("%s...%s", prefix, suffix)
}

// maskSensitiveData recursively masks sensitive data in a map
func maskSensitiveData(data interface{}) interface{} {
	return maskSensitiveDataRecursive(data, "")
}

// maskSensitiveDataRecursive recursively processes data structures and masks sensitive values
func maskSensitiveDataRecursive(data interface{}, parentKey string) interface{} {
	if data == nil {
		return nil
	}

	val := reflect.ValueOf(data)

	switch val.Kind() {
	case reflect.Map:
		result := make(map[string]interface{})
		iter := val.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			value := iter.Value().Interface()

			// Check if this key represents sensitive data
			if isSensitiveField(key) || isSensitiveField(parentKey+"."+key) {
				// Mask the value if it's a string
				if strVal, ok := value.(string); ok && strVal != "" {
					result[key] = maskValue(strVal)
				} else {
					result[key] = value
				}
			} else {
				// Recursively process nested structures
				// Build the full path for the next level
				nextParentKey := key
				if parentKey != "" {
					nextParentKey = parentKey + "." + key
				}
				result[key] = maskSensitiveDataRecursive(value, nextParentKey)
			}
		}
		return result

	case reflect.Slice, reflect.Array:
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = maskSensitiveDataRecursive(val.Index(i).Interface(), parentKey)
		}
		return result

	case reflect.Struct:
		// Convert struct to map for processing
		result := make(map[string]interface{})
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			fieldName := field.Name
			if isSensitiveField(fieldName) || isSensitiveField(parentKey+"."+fieldName) {
				// Mask the value if it's a string
				if fieldValue.Kind() == reflect.String && fieldValue.String() != "" {
					result[fieldName] = maskValue(fieldValue.String())
				} else {
					result[fieldName] = fieldValue.Interface()
				}
			} else {
				result[fieldName] = maskSensitiveDataRecursive(fieldValue.Interface(), fieldName)
			}
		}
		return result

	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return maskSensitiveDataRecursive(val.Elem().Interface(), parentKey)

	default:
		return data
	}
}

// formatConfigOutput formats configuration data for display
func formatConfigOutput(data interface{}, indent int) string {
	var sb strings.Builder
	indentStr := strings.Repeat("  ", indent)

	val := reflect.ValueOf(data)
	if !val.IsValid() {
		return ""
	}

	switch val.Kind() {
	case reflect.Map:
		iter := val.MapRange()
		keys := make([]string, 0)
		values := make(map[string]interface{})

		// Collect and sort keys
		for iter.Next() {
			key := iter.Key().String()
			keys = append(keys, key)
			values[key] = iter.Value().Interface()
		}

		// Sort keys for consistent output
		// Note: Using a simple approach here; could use sort.Strings for alphabetical
		for _, key := range keys {
			value := values[key]
			if value == nil {
				continue
			}

			// Check if value is a nested structure
			valReflect := reflect.ValueOf(value)
			if valReflect.Kind() == reflect.Map && valReflect.Len() > 0 {
				sb.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
				sb.WriteString(formatConfigOutput(value, indent+1))
			} else if valReflect.Kind() == reflect.Slice && valReflect.Len() > 0 {
				sb.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
				sb.WriteString(formatConfigOutput(value, indent+1))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %v\n", indentStr, key, value))
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			sb.WriteString(fmt.Sprintf("%s- %v\n", indentStr, val.Index(i).Interface()))
		}

	default:
		sb.WriteString(fmt.Sprintf("%s%v\n", indentStr, data))
	}

	return sb.String()
}
