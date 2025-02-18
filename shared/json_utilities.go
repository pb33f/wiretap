package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
)

// Function to check if json is a subset of superSet
func IsSubset(json, superSet interface{}) bool {
	// Check if json is an object
	if obj, ok := json.(map[string]interface{}); ok {
		// Check if superSet is an object
		if ssObj, ok := superSet.(map[string]interface{}); ok {
			// Call helper function to check if map1 is a subset of ssObj
			return isMapSubset(obj, ssObj)
		}
	}

	// Check if json is a array
	if array, ok := json.([]interface{}); ok {
		// Check if superSet is a array
		if ssArray, ok := superSet.([]interface{}); ok {
			// Call helper function to check if array is a subset of ssArray
			return isSliceSubset(array, ssArray)
		}
	}

	// Check if json is a string
	if str, ok := json.(string); ok {
		// Check if superSet is a string
		if ssStr, ok := superSet.(string); ok {
			// Call helper function to compare strings
			return StringCompare(str, ssStr)
		}

		// if json is a string and superSet is not, then it can't be a subset
		return false
	}

	// If json is neither object nor array, then it must be a primitive type and can be compared directly
	return json == superSet
}

// Helper function to check if objOne is a subset of objTwo
func isMapSubset(objOne, objTwo map[string]interface{}) bool {
	for key, valueObjOne := range objOne {
		// Check if the key exists in objTwo
		if valueObjTwo, exists := objTwo[key]; exists {
			// Recursively check if the values are deeply equal
			return IsSubset(valueObjOne, valueObjTwo)
		} else {
			// Key doesn't exist in objTwo
			return false
		}
	}
	return true
}

// Helper function to check if arrOne is a subset of arrTwo
func isSliceSubset(arrOne, arrTwo []interface{}) bool {
	// Check if all elements of arrOne exist in arrTwo
	for _, elementOne := range arrOne {
		found := false
		for _, elementTwo := range arrTwo {
			if IsSubset(elementOne, elementTwo) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Function to check if a string contains regex-specific characters
func isPotentialRegex(s string) bool {
	// Regex characters to check for
	regexChars := []string{".", "*", "+", "?", "^", "$", "[", "]", "(", ")", "|", "{", "}"}

	// Check if the string contains any of the characters used in regular expressions
	for _, char := range regexChars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

// Helper function to compare strings. Since mock definitions support
// regex, we need to check if the string is a regex and if so, compare it
func StringCompare(patternOrStr, str string) bool {
	// Try to compile the pattern to check if it's a valid regex
	if isPotentialRegex(patternOrStr) {
		if _, err := regexp.Compile(patternOrStr); err != nil {
			// If it's not a valid regex, do a normal string comparison
			return patternOrStr == str
		}

		// If it's a valid regex, check if it matches the string
		matched, err := regexp.MatchString(patternOrStr, str)
		if err != nil {
			return false
		}

		return matched
	}

	// If it's not a regex, do a normal string comparison
	return patternOrStr == str
}

// Helper function to get the value based on a JSON path (dot notation or array index).
func getValueByPath(data interface{}, path string) (interface{}, error) {
	// Split the path by dots and array indices (i.e., "[7]")
	re := regexp.MustCompile(`\.`)
	parts := re.Split(path, -1)

	// Traverse the path
	for _, part := range parts {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Handle array access, e.g., "[7]"
			// Get the array index
			index := part[1 : len(part)-1]
			// Traverse the array using the index
			if idx, err := strconv.Atoi(index); err == nil {
				if arr, ok := data.([]interface{}); ok && idx < len(arr) {
					data = arr[idx]
				} else {
					return nil, fmt.Errorf("invalid array index or not an array")
				}
			} else {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
		} else {
			// Handle map key access
			if m, ok := data.(map[string]interface{}); ok {
				if val, found := m[part]; found {
					data = val
				} else {
					return nil, fmt.Errorf("key not found: %s", part)
				}
			} else {
				return nil, fmt.Errorf("not a map: %v", data)
			}
		}
	}

	return data, nil
}

// Function to replace template variables in JSON path format
func ReplaceTemplateVars(jsonStr string, vars interface{}) (string, error) {
	// Regular expression to match the ${var} format (full path)
	re := regexp.MustCompile(`\$\{([a-zA-Z0-9._\[\]]+)\}`)

	matches := re.Match([]byte(jsonStr))
	pterm.Info.Println(matches)

	var handleMatchedString = func(match string) string {
		// Extract the path within the ${}
		path := match[2 : len(match)-1]

		// Get the value from the vars object using the JSON path
		value, err := getValueByPath(vars, path)
		if err != nil {
			// If the path is not valid, return the original placeholder
			return match
		}

		// Convert the value to a string (you can handle other types here if needed)
		switch v := value.(type) {
		case string:
			return v
		case float64: // Handle numbers
			return fmt.Sprintf("%v", v)
		case bool: // Handle booleans
			return fmt.Sprintf("%v", v)
		default:
			// Handle other types, like arrays or maps, as needed
			return fmt.Sprintf("%v", v)
		}
	}

	// Replace the placeholders with corresponding values
	result := re.ReplaceAllStringFunc(jsonStr, handleMatchedString)

	return result, nil
}
