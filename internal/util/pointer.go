// Package util provides utility functions for the dehydrated-api-go application.
package util

// String returns the string value of a string pointer.
// If the pointer is nil, it returns an empty string.
func String(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}

// StringSlice returns the string slice value of a string slice pointer.
// If the pointer is nil, it returns nil.
func StringSlice(s *[]string) []string {
	if s == nil {
		return nil
	}
	return *s
}

// StringSlicePtr returns a pointer to the given string slice.
func StringSlicePtr(s []string) *[]string {
	return &s
}

// Bool returns the boolean value of a boolean pointer.
// If the pointer is nil, it returns false.
func Bool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// BoolPtr returns a pointer to the given boolean.
func BoolPtr(b bool) *bool {
	return &b
}
