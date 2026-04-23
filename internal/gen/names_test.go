package gen

import (
	"testing"
)

func TestToPascal(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"user_id", "UserID"},
		{"video_card", "VideoCard"},
		{"http_url", "HTTPURL"},
		{"created_at", "CreatedAt"},
		{"is_active", "IsActive"},
		{"api_key", "APIKey"},
		{"json_data", "JSONData"},
		{"uuid", "UUID"},
		{"id", "ID"},
		{"email", "Email"},
		{"music_theme_id", "MusicThemeID"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := ToPascal(tt.in)
			if got != tt.want {
				t.Errorf("ToPascal(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestToCamel(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"user_id", "userID"},
		{"video_card", "videoCard"},
		{"created_at", "createdAt"},
		{"is_active", "isActive"},
		{"api_key", "apiKey"},
		{"id", "id"},
		{"email", "email"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := ToCamel(tt.in)
			if got != tt.want {
				t.Errorf("ToCamel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"user_id", "user-id"},
		{"video_card", "video-card"},
		{"VideoCard", "video-card"},
		{"musicTheme", "music-theme"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := ToKebab(tt.in)
			if got != tt.want {
				t.Errorf("ToKebab(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestToPlural(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"user", "users"},
		{"card", "cards"},
		{"category", "categories"},
		{"class", "classes"},
		{"box", "boxes"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := ToPlural(tt.in)
			if got != tt.want {
				t.Errorf("ToPlural(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBaseType(t *testing.T) {
	if BaseType("*time.Time") != "time.Time" {
		t.Error("expected *time.Time → time.Time")
	}
	if BaseType("string") != "string" {
		t.Error("expected string → string")
	}
}
