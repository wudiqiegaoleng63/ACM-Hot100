package auth

import (
	"encoding/hex"
	"testing"
)

func TestHashToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "empty token",
			raw:  "",
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "verification token",
			raw:  "verify-token",
			want: "458ba985765983a9f2054fa2073b5e80e253c3e842266cbf6f10310945c374be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HashToken(tt.raw); got != tt.want {
				t.Fatalf("HashToken(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	generators := []struct {
		name     string
		generate func() (string, string, error)
	}{
		{name: "verification", generate: GenerateVerifyToken},
		{name: "password reset", generate: GenerateResetToken},
	}

	for _, generator := range generators {
		generator := generator
		t.Run(generator.name, func(t *testing.T) {
			t.Parallel()

			firstRaw, firstHash, err := generator.generate()
			if err != nil {
				t.Fatalf("generate first token: %v", err)
			}
			secondRaw, secondHash, err := generator.generate()
			if err != nil {
				t.Fatalf("generate second token: %v", err)
			}

			for _, token := range []struct {
				name string
				raw  string
				hash string
			}{
				{name: "first", raw: firstRaw, hash: firstHash},
				{name: "second", raw: secondRaw, hash: secondHash},
			} {
				if len(token.raw) != 64 {
					t.Errorf("%s raw token length = %d, want 64", token.name, len(token.raw))
				}
				if _, err := hex.DecodeString(token.raw); err != nil {
					t.Errorf("%s raw token is not hexadecimal: %v", token.name, err)
				}
				if token.hash != HashToken(token.raw) {
					t.Errorf("%s hash does not match raw token", token.name)
				}
			}

			if firstRaw == secondRaw || firstHash == secondHash {
				t.Fatal("two generated tokens must be distinct")
			}
		})
	}
}
