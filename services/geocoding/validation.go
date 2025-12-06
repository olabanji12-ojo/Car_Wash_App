package geocoding

import (
    "errors"
    "strings"
)

var (
    ErrInvalidAddress = errors.New("invalid address format")
)

// ValidateAddress performs basic validation on an address
func ValidateAddress(address string) error {
    if strings.TrimSpace(address) == "" {
        return errors.New("address cannot be empty")
    }
    
    // Check minimum length (e.g., at least 5 characters)
    if len(strings.TrimSpace(address)) < 5 {
        return errors.New("address is too short")
    }
    
    // Check for minimum components (at least street number and name)
    parts := strings.Fields(address)
    if len(parts) < 2 {
        return errors.New("address should include street number and name")
    }
    
    return nil
}

// FormatAddress standardizes the address format
func FormatAddress(address string) string {
    // Trim whitespace and clean up multiple spaces
    return strings.Join(strings.Fields(strings.TrimSpace(address)), " ")
}