package order

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComputeTotals(t *testing.T) {
	cases := []struct {
		name           string
		subtotal       int64
		taxPercent     float64
		servicePercent float64
		wantTax        int64
		wantService    int64
		wantTotal      int64
	}{
		{
			name:           "10% tax, 0% service",
			subtotal:       100000,
			taxPercent:     10,
			servicePercent: 0,
			wantTax:        10000,
			wantService:    0,
			wantTotal:      110000,
		},
		{
			name:           "10% tax, 5% service",
			subtotal:       100000,
			taxPercent:     10,
			servicePercent: 5,
			wantTax:        10000,
			wantService:    5000,
			wantTotal:      115000,
		},
		{
			name:           "zero everything",
			subtotal:       0,
			taxPercent:     10,
			servicePercent: 5,
			wantTax:        0,
			wantService:    0,
			wantTotal:      0,
		},
		{
			name:           "fractional percentages truncate toward zero",
			subtotal:       33333,
			taxPercent:     10,
			servicePercent: 5,
			wantTax:        3333, // 3333.3 -> 3333
			wantService:    1666, // 1666.65 -> 1666
			wantTotal:      38332,
		},
		{
			name:           "no tax no service",
			subtotal:       50000,
			taxPercent:     0,
			servicePercent: 0,
			wantTax:        0,
			wantService:    0,
			wantTotal:      50000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tax, service, total := computeTotals(tc.subtotal, tc.taxPercent, tc.servicePercent)
			assert.Equal(t, tc.wantTax, tax, "tax")
			assert.Equal(t, tc.wantService, service, "service")
			assert.Equal(t, tc.wantTotal, total, "total")
			// Invariant: total is always subtotal + tax + service.
			assert.Equal(t, tc.subtotal+tax+service, total)
		})
	}
}

func TestGenerateOrderNumber(t *testing.T) {
	when := time.Date(2026, time.February, 9, 12, 0, 0, 0, time.UTC)

	got := generateOrderNumber(when)
	assert.True(t, strings.HasPrefix(got, "ORD-260209-"), "unexpected prefix: %s", got)

	// The random suffix is 3 bytes hex-encoded => 6 uppercase hex chars.
	suffix := strings.TrimPrefix(got, "ORD-260209-")
	assert.Len(t, suffix, 6)
	assert.Equal(t, strings.ToUpper(suffix), suffix, "suffix must be uppercase hex")

	// Two calls must (with overwhelming probability) differ.
	assert.NotEqual(t, generateOrderNumber(when), generateOrderNumber(when))
}

func TestIsAllowedProofURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want bool
	}{
		{"googleusercontent allowed", "https://lh3.googleusercontent.com/abc.png", true},
		{"drive allowed", "https://drive.google.com/file/d/xyz/view", true},
		{"docs allowed", "https://docs.google.com/uc?id=1", true},
		{"script googleusercontent allowed", "https://script.googleusercontent.com/x", true},
		{"uppercase host normalized", "https://DRIVE.Google.COM/x", true},
		{"http scheme allowed", "http://drive.google.com/x", true},
		{"evil host rejected", "https://evil.example.com/x.png", false},
		{"subdomain spoof rejected", "https://drive.google.com.evil.com/x", false},
		{"scheme-relative rejected", "//drive.google.com/x", false},
		{"javascript scheme rejected", "javascript:alert(1)", false},
		{"ftp scheme rejected", "ftp://drive.google.com/x", false},
		{"empty rejected", "", false},
		{"garbage rejected", "not a url", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isAllowedProofURL(tc.url))
		})
	}
}
