package helpers

import "regexp"

// IsNumericRequestID checks if a string is a numeric request ID
func IsNumericRequestID(requestID string) bool {
	match, err := regexp.MatchString(`^[0-9]+$`, requestID)
	if err != nil {
		return false
	}
	return match
}
