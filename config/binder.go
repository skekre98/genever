package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

// Binder decodes map[string]any data into Go structs and validates the result.
//
// Binder uses two-stage processing:
//  1. Decode: Converts untyped maps to typed structs using mapstructure
//  2. Validate: Checks struct field values against validation rules
//
// The Binder supports:
//   - Automatic type conversion (string to int, string to duration, etc.)
//   - Nested struct decoding
//   - Slice and map handling
//   - Rich validation rules via struct tags
//   - Custom decode hooks for complex types
//
// Struct fields should use `config` tags for field mapping and `validate`
// tags for validation rules.
//
// Example struct:
//
//	type ServerConfig struct {
//	    Port    int           `config:"port" validate:"required,min=1,max=65535"`
//	    Host    string        `config:"host" validate:"required,hostname"`
//	    Timeout time.Duration `config:"timeout" validate:"required"`
//	}
type Binder struct {
	validator *validator.Validate
}

// BindError represents an error that occurred during the bind or validate stage.
//
// BindError wraps the underlying error and indicates which stage failed.
// This allows callers to distinguish between decode errors (invalid data types)
// and validation errors (invalid data values).
type BindError struct {
	// Stage indicates which phase failed: "decode" or "validate"
	Stage string

	// Err is the underlying error from mapstructure or validator
	Err error
}

// Error implements the error interface.
func (e *BindError) Error() string {
	return fmt.Sprintf("config %s error: %v", e.Stage, e.Err)
}

// Unwrap returns the underlying error, enabling errors.Is and errors.As.
func (e *BindError) Unwrap() error {
	return e.Err
}

// NewBinder creates a new Binder with default decode hooks and validators.
//
// The default configuration includes:
//   - String to time.Duration conversion ("5s" -> 5*time.Second)
//   - Comma-separated string to slice conversion ("a,b,c" -> []string{"a","b","c"})
//   - Weak type conversion (string "123" -> int 123)
//   - Standard validation rules from go-playground/validator
func NewBinder() *Binder {
	return &Binder{
		validator: validator.New(),
	}
}

// Bind decodes the source map into the target struct and validates it.
//
// The target parameter must be a pointer to a struct. Bind will populate the
// struct fields based on the source data and then validate the result.
//
// The binding process:
//  1. Decode source map into target struct using field tags
//  2. Apply type conversions (strings to durations, etc.)
//  3. Validate all fields against their validation rules
//
// If either stage fails, a BindError is returned with the stage and underlying
// error. The target struct may be partially populated if decode succeeds but
// validation fails.
//
// Example:
//
//	var cfg ServerConfig
//	source := map[string]any{
//	    "port": "8080",
//	    "host": "localhost",
//	    "timeout": "30s",
//	}
//	err := binder.Bind(source, &cfg)
//
// Returns a BindError if:
//   - Decode fails: type mismatch, invalid format, unknown field
//   - Validate fails: value violates validation rules
func (b *Binder) Bind(source map[string]any, target any) error {
	if err := b.decode(source, target); err != nil {
		return &BindError{
			Stage: "decode",
			Err:   err,
		}
	}

	if err := b.validate(target); err != nil {
		return &BindError{
			Stage: "validate",
			Err:   err,
		}
	}

	return nil
}

func (b *Binder) decode(source map[string]any, target any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		TagName: "config",
	})
	if err != nil {
		return err
	}

	return decoder.Decode(source)
}

func (b *Binder) validate(target any) error {
	return b.validator.Struct(target)
}
