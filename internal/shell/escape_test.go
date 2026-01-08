package shell

import (
	"testing"
)

func TestShellEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "safe string",
			input: "simple-name_123",
			want:  "simple-name_123",
		},
		{
			name:  "string with spaces",
			input: "hello world",
			want:  "'hello world'",
		},
		{
			name:  "string with single quote",
			input: "it's working",
			want:  "'it'\\''s working'",
		},
		{
			name:  "string with semicolon (command injection attempt)",
			input: "test; rm -rf /",
			want:  "'test; rm -rf /'",
		},
		{
			name:  "string with backticks",
			input: "`whoami`",
			want:  "'`whoami`'",
		},
		{
			name:  "string with dollar sign",
			input: "$HOME/test",
			want:  "'$HOME/test'",
		},
		{
			name:  "path traversal attempt",
			input: "../../../etc/passwd",
			want:  "'../../../etc/passwd'",
		},
		{
			name:  "empty string",
			input: "",
			want:  "''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellEscape(tt.input)
			if got != tt.want {
				t.Errorf("ShellEscape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateVolumeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid simple name",
			input: "myvolume",
			want:  true,
		},
		{
			name:  "valid with dash",
			input: "my-volume",
			want:  true,
		},
		{
			name:  "valid with underscore",
			input: "my_volume",
			want:  true,
		},
		{
			name:  "valid with dot",
			input: "my.volume",
			want:  true,
		},
		{
			name:  "valid with numbers",
			input: "volume123",
			want:  true,
		},
		{
			name:  "valid complex",
			input: "app_data-v1.0",
			want:  true,
		},
		{
			name:  "invalid: empty",
			input: "",
			want:  false,
		},
		{
			name:  "invalid: starts with dash",
			input: "-volume",
			want:  false,
		},
		{
			name:  "invalid: starts with dot",
			input: ".volume",
			want:  false,
		},
		{
			name:  "invalid: path traversal",
			input: "../volume",
			want:  false,
		},
		{
			name:  "invalid: contains slash",
			input: "my/volume",
			want:  false,
		},
		{
			name:  "invalid: contains backslash",
			input: "my\\volume",
			want:  false,
		},
		{
			name:  "invalid: contains space",
			input: "my volume",
			want:  false,
		},
		{
			name:  "invalid: contains semicolon",
			input: "volume;rm",
			want:  false,
		},
		{
			name:  "invalid: contains dollar",
			input: "$volume",
			want:  false,
		},
		{
			name:  "invalid: command injection attempt",
			input: "test`whoami`",
			want:  false,
		},
		{
			name:  "invalid: too long",
			input: string(make([]byte, 256)),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateVolumeName(tt.input)
			if got != tt.want {
				t.Errorf("ValidateVolumeName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizePathForRemote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple absolute path",
			input: "/tmp/test",
			want:  "/tmp/test",
		},
		{
			name:  "path traversal removed",
			input: "/tmp/../../../etc/passwd",
			want:  "/tmp/etc/passwd",
		},
		{
			name:  "relative path made absolute",
			input: "tmp/test",
			want:  "/tmp/test",
		},
		{
			name:  "double slashes removed",
			input: "/tmp//test",
			want:  "/tmp/test",
		},
		{
			name:  "multiple issues combined",
			input: "tmp/..//test/./path",
			want:  "/tmp/test/./path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePathForRemote(tt.input)
			if got != tt.want {
				t.Errorf("SanitizePathForRemote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
