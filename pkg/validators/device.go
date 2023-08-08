package validators

import "github.com/pkg/errors"

var (
	ErrInvalidDeviceType = errors.New("invalid device type")
)

// ValidateDeviceType checks if the device type is one of the allowed types
func ValidateDeviceType(deviceType string) error {
	switch deviceType {
	case "web", "android", "ios":
		return nil
	default:
		return ErrInvalidDeviceType
	}
}
