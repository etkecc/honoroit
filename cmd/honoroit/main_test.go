package main

import "testing"

func TestListenAddress(t *testing.T) {
	tests := []struct {
		port string
		want string
	}{
		{"8080", ":8080"},
		{":8080", ":8080"},
		{" 8080 ", ":8080"},
		{"0.0.0.0:8080", "0.0.0.0:8080"},
		{"[::1]:8080", "[::1]:8080"},
	}
	for _, tt := range tests {
		if got := listenAddress(tt.port); got != tt.want {
			t.Errorf("listenAddress(%q) = %q, want %q", tt.port, got, tt.want)
		}
	}
}
