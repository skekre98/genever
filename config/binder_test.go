package config_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/skekre98/genever/config"
)

func TestBinder_Bind_SimpleTypes(t *testing.T) {
	type SimpleConfig struct {
		Name    string `config:"name" validate:"required"`
		Port    int    `config:"port" validate:"min=1,max=65535"`
		Enabled bool   `config:"enabled"`
	}

	tests := []struct {
		name    string
		source  map[string]any
		want    SimpleConfig
		wantErr bool
	}{
		{
			name: "valid config",
			source: map[string]any{
				"name":    "test-app",
				"port":    8080,
				"enabled": true,
			},
			want: SimpleConfig{
				Name:    "test-app",
				Port:    8080,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "type conversion - string to int",
			source: map[string]any{
				"name":    "test-app",
				"port":    "8080",
				"enabled": true,
			},
			want: SimpleConfig{
				Name:    "test-app",
				Port:    8080,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "validation error - missing required field",
			source: map[string]any{
				"port":    8080,
				"enabled": true,
			},
			wantErr: true,
		},
		{
			name: "validation error - port too high",
			source: map[string]any{
				"name":    "test-app",
				"port":    99999,
				"enabled": true,
			},
			wantErr: true,
		},
		{
			name: "validation error - port too low",
			source: map[string]any{
				"name":    "test-app",
				"port":    0,
				"enabled": true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binder := config.NewBinder()
			var got SimpleConfig

			err := binder.Bind(tt.source, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bind() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBinder_Bind_NestedStructs(t *testing.T) {
	type Database struct {
		Host string `config:"host" validate:"required"`
		Port int    `config:"port" validate:"required,min=1,max=65535"`
	}

	type Server struct {
		Port    int           `config:"port" validate:"required,min=1,max=65535"`
		Timeout time.Duration `config:"timeout"`
	}

	type AppConfig struct {
		Name     string   `config:"name" validate:"required"`
		Database Database `config:"database"`
		Server   Server   `config:"server"`
	}

	tests := []struct {
		name    string
		source  map[string]any
		want    AppConfig
		wantErr bool
	}{
		{
			name: "valid nested config",
			source: map[string]any{
				"name": "test-app",
				"database": map[string]any{
					"host": "localhost",
					"port": 5432,
				},
				"server": map[string]any{
					"port":    8080,
					"timeout": "30s",
				},
			},
			want: AppConfig{
				Name: "test-app",
				Database: Database{
					Host: "localhost",
					Port: 5432,
				},
				Server: Server{
					Port:    8080,
					Timeout: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "validation error - missing nested required field",
			source: map[string]any{
				"name": "test-app",
				"database": map[string]any{
					"port": 5432,
				},
				"server": map[string]any{
					"port": 8080,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binder := config.NewBinder()
			var got AppConfig

			err := binder.Bind(tt.source, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bind() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBinder_Bind_Slices(t *testing.T) {
	type Config struct {
		Tags    []string `config:"tags"`
		Ports   []int    `config:"ports" validate:"required,min=1"`
		Enabled []bool   `config:"enabled"`
	}

	tests := []struct {
		name    string
		source  map[string]any
		want    Config
		wantErr bool
	}{
		{
			name: "valid slices",
			source: map[string]any{
				"tags":    []string{"web", "api", "production"},
				"ports":   []int{8080, 8081, 8082},
				"enabled": []bool{true, false, true},
			},
			want: Config{
				Tags:    []string{"web", "api", "production"},
				Ports:   []int{8080, 8081, 8082},
				Enabled: []bool{true, false, true},
			},
			wantErr: false,
		},
		{
			name: "comma-separated string to slice",
			source: map[string]any{
				"tags":    "web,api,production",
				"ports":   []int{8080},
				"enabled": []bool{true},
			},
			want: Config{
				Tags:    []string{"web", "api", "production"},
				Ports:   []int{8080},
				Enabled: []bool{true},
			},
			wantErr: false,
		},
		{
			name: "validation error - empty required slice",
			source: map[string]any{
				"tags":    []string{"web"},
				"ports":   []int{},
				"enabled": []bool{true},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binder := config.NewBinder()
			var got Config

			err := binder.Bind(tt.source, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bind() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBinder_Bind_DurationConversion(t *testing.T) {
	type Config struct {
		Timeout       time.Duration `config:"timeout" validate:"required"`
		RetryInterval time.Duration `config:"retry_interval"`
	}

	tests := []struct {
		name    string
		source  map[string]any
		want    Config
		wantErr bool
	}{
		{
			name: "string to duration",
			source: map[string]any{
				"timeout":        "30s",
				"retry_interval": "5m",
			},
			want: Config{
				Timeout:       30 * time.Second,
				RetryInterval: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "various duration formats",
			source: map[string]any{
				"timeout":        "1h30m",
				"retry_interval": "500ms",
			},
			want: Config{
				Timeout:       90 * time.Minute,
				RetryInterval: 500 * time.Millisecond,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binder := config.NewBinder()
			var got Config

			err := binder.Bind(tt.source, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bind() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBinder_Bind_ComplexValidation(t *testing.T) {
	type Config struct {
		Email    string   `config:"email" validate:"required,email"`
		URL      string   `config:"url" validate:"required,url"`
		MinValue int      `config:"min_value" validate:"gte=10,lte=100"`
		OneOf    string   `config:"one_of" validate:"oneof=dev staging prod"`
		Tags     []string `config:"tags" validate:"dive,min=2"`
	}

	tests := []struct {
		name    string
		source  map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid complex config",
			source: map[string]any{
				"email":     "test@example.com",
				"url":       "https://example.com",
				"min_value": 50,
				"one_of":    "prod",
				"tags":      []string{"web", "api"},
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			source: map[string]any{
				"email":     "invalid-email",
				"url":       "https://example.com",
				"min_value": 50,
				"one_of":    "prod",
				"tags":      []string{"web", "api"},
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			source: map[string]any{
				"email":     "test@example.com",
				"url":       "not-a-url",
				"min_value": 50,
				"one_of":    "prod",
				"tags":      []string{"web", "api"},
			},
			wantErr: true,
		},
		{
			name: "value out of range",
			source: map[string]any{
				"email":     "test@example.com",
				"url":       "https://example.com",
				"min_value": 5,
				"one_of":    "prod",
				"tags":      []string{"web", "api"},
			},
			wantErr: true,
		},
		{
			name: "invalid oneof",
			source: map[string]any{
				"email":     "test@example.com",
				"url":       "https://example.com",
				"min_value": 50,
				"one_of":    "invalid",
				"tags":      []string{"web", "api"},
			},
			wantErr: true,
		},
		{
			name: "tag too short (dive validation)",
			source: map[string]any{
				"email":     "test@example.com",
				"url":       "https://example.com",
				"min_value": 50,
				"one_of":    "prod",
				"tags":      []string{"a", "api"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binder := config.NewBinder()
			var got Config

			err := binder.Bind(tt.source, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBinder_BindError(t *testing.T) {
	binder := config.NewBinder()

	t.Run("decode error", func(t *testing.T) {
		type Config struct {
			Value int `config:"value"`
		}
		source := map[string]any{
			"value": "not-a-number",
		}
		var cfg Config

		err := binder.Bind(source, &cfg)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var bindErr *config.BindError
		if !errors.As(err, &bindErr) {
			t.Errorf("expected BindError, got %T", err)
		}
		if bindErr.Stage != "decode" {
			t.Errorf("expected stage 'decode', got %s", bindErr.Stage)
		}
	})

	t.Run("validate error", func(t *testing.T) {
		type Config struct {
			Value int `config:"value" validate:"min=10"`
		}
		source := map[string]any{
			"value": 5,
		}
		var cfg Config

		err := binder.Bind(source, &cfg)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var bindErr *config.BindError
		if !errors.As(err, &bindErr) {
			t.Errorf("expected BindError, got %T", err)
		}
		if bindErr.Stage != "validate" {
			t.Errorf("expected stage 'validate', got %s", bindErr.Stage)
		}
	})
}

func TestBinder_Bind_EmptySource(t *testing.T) {
	type Config struct {
		Name string `config:"name"`
		Port int    `config:"port"`
	}

	binder := config.NewBinder()
	var got Config
	source := map[string]any{}

	// Should succeed with zero values
	err := binder.Bind(source, &got)
	if err != nil {
		t.Errorf("Bind() with empty source error = %v", err)
	}

	expected := Config{}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Bind() got = %+v, want %+v", got, expected)
	}
}

func TestBinder_Bind_ExtraFields(t *testing.T) {
	type Config struct {
		Name string `config:"name"`
		Port int    `config:"port"`
	}

	binder := config.NewBinder()
	var got Config
	source := map[string]any{
		"name":       "test",
		"port":       8080,
		"extra":      "ignored",
		"another":    123,
		"unexpected": true,
	}

	err := binder.Bind(source, &got)
	if err != nil {
		t.Errorf("Bind() error = %v", err)
	}

	expected := Config{
		Name: "test",
		Port: 8080,
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Bind() got = %+v, want %+v", got, expected)
	}
}

func TestBindError_Error(t *testing.T) {
	err := &config.BindError{
		Stage: "decode",
		Err:   errors.New("test error"),
	}

	expected := "config decode error: test error"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestBindError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &config.BindError{
		Stage: "validate",
		Err:   innerErr,
	}

	if err.Unwrap() != innerErr {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), innerErr)
	}
}
