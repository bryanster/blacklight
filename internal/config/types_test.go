package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"strings"
	"testing"
)

// TestSecretIsRedactedInEveryRendering covers the ways a Config actually
// reaches a log or a support bundle. Each one goes through a different method,
// so each one is a separate way to leak the session key.
func TestSecretIsRedactedInEveryRendering(t *testing.T) {
	const value = "a-secret-that-must-never-be-printed-anywhere"

	cfg, errs := parse(envWith(map[string]string{envSessionSecret: value}))
	if len(errs) > 0 {
		t.Fatalf("parse() = %v, want no errors", errs)
	}

	renderings := map[string]func() string{
		`fmt "%v" of the config`:  func() string { return fmt.Sprintf("%v", cfg) },
		`fmt "%+v" of the config`: func() string { return fmt.Sprintf("%+v", cfg) },
		`fmt "%#v" of the config`: func() string { return fmt.Sprintf("%#v", cfg) },
		// A Config has held a field with no String method since M1-004, so "%s"
		// of one is a verb the vet and staticcheck are both right to object to.
		// It stays: reaching for the wrong verb is exactly the accident this test
		// is about, and a redaction that only holds for the verbs a linter
		// approves of is not a redaction.
		//nolint:staticcheck // SA5009: the wrong verb is the case under test.
		`fmt "%s" of the config`:  func() string { return fmt.Sprintf("%s", any(cfg)) },
		`fmt "%v" of the secret`:  func() string { return fmt.Sprintf("%v", cfg.Session.Secret) },
		`fmt "%#v" of the secret`: func() string { return fmt.Sprintf("%#v", cfg.Session.Secret) },
		"json of the config": func() string {
			b, err := json.Marshal(cfg)
			if err != nil {
				t.Fatalf("json.Marshal(cfg): %v", err)
			}
			return string(b)
		},
		"slog json handler": func() string {
			var buf bytes.Buffer
			slog.New(slog.NewJSONHandler(&buf, nil)).Info("startup", "config", cfg)
			return buf.String()
		},
		"slog text handler": func() string {
			var buf bytes.Buffer
			slog.New(slog.NewTextHandler(&buf, nil)).Info("startup", "config", cfg)
			return buf.String()
		},
		"slog attribute of the secret": func() string {
			var buf bytes.Buffer
			slog.New(slog.NewJSONHandler(&buf, nil)).Info("startup", "secret", cfg.Session.Secret)
			return buf.String()
		},
	}

	for name, render := range renderings {
		t.Run(name, func(t *testing.T) {
			got := render()
			if strings.Contains(got, value) {
				t.Errorf("the secret appears in %s:\n\t%s", name, got)
			}
			if !strings.Contains(got, redactedPlaceholder) {
				t.Errorf("%s does not show %s, so the redaction may not have run:\n\t%s",
					name, redactedPlaceholder, got)
			}
		})
	}
}

func TestSecretReveal(t *testing.T) {
	secret := NewSecret([]byte("the-value"))

	if got, want := string(secret.Reveal()), "the-value"; got != want {
		t.Errorf("Reveal() = %q, want %q", got, want)
	}

	// Reveal hands out a copy: a caller that scribbles on the slice must not
	// be able to change the key every other caller sees.
	revealed := secret.Reveal()
	revealed[0] = 'X'
	if got, want := string(secret.Reveal()), "the-value"; got != want {
		t.Errorf("after mutating a revealed copy, Reveal() = %q, want %q", got, want)
	}
}

func TestZeroSecretRendersAsUnset(t *testing.T) {
	var secret Secret

	if !secret.IsZero() {
		t.Error("IsZero() = false for the zero Secret")
	}
	if got := fmt.Sprintf("%v", secret); got != unsetPlaceholder {
		t.Errorf("%%v of the zero Secret = %q, want %q", got, unsetPlaceholder)
	}
}

func TestSecretStrength(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool // accepted?
	}{
		{name: "openssl rand -base64 32", value: testSecret, want: true},
		{name: "openssl rand -hex 32", value: "3f7a1c9e5b2d8046af13c5e7920bd4681fa3c05d9e7b264180ac539fe62d7b41", want: true},
		{name: "32 raw bytes", value: "Xq7!vP2#mL9$wR4%tK6^zN8&cB3*hJ5(", want: true},
		{name: "31 raw bytes", value: "Xq7!vP2#mL9$wR4%tK6^zN8&cB3*hJ5", want: false},
		{name: "base64 of 31 bytes", value: "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MA==", want: false},
		{name: "empty", value: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var secret Secret
			err := secret.UnmarshalText([]byte(tt.value))

			if got := err == nil; got != tt.want {
				t.Fatalf("UnmarshalText(%q) error = %v, want accepted = %v", tt.value, err, tt.want)
			}
			if err != nil && strings.Contains(err.Error(), tt.value) && tt.value != "" {
				t.Errorf("the error repeats the secret back: %v", err)
			}
		})
	}
}

func TestSecretBytesMeasuresTheSmallestReading(t *testing.T) {
	tests := []struct {
		value string
		want  int
	}{
		{value: "", want: 0},
		{value: testSecret, want: 32}, // 44 base64 characters
		{value: "cmVhbGx5IG5vdCB0aGlydHktdHdvIGJ5dGVzIQ==", want: 28},    // long, thin
		{value: "correct horse battery staple, and then some", want: 43}, // decodes as nothing
		{value: "00112233445566778899aabbccddeeff", want: 16},            // 32 hex characters
	}

	for _, tt := range tests {
		if got := secretBytes(tt.value); got != tt.want {
			t.Errorf("secretBytes(%q) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestURLNormalisation(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "https://example.internal", want: "https://example.internal"},
		{raw: "https://example.internal/", want: "https://example.internal"},
		{raw: "https://example.internal///", want: "https://example.internal"},
		{raw: "http://example.internal:8080/base/", want: "http://example.internal:8080/base"},
		{raw: "  https://example.internal  ", want: "https://example.internal"},
	}

	for _, tt := range tests {
		var u URL
		if err := u.UnmarshalText([]byte(tt.raw)); err != nil {
			t.Errorf("UnmarshalText(%q) = %v", tt.raw, err)
			continue
		}
		if got := u.String(); got != tt.want {
			t.Errorf("UnmarshalText(%q) produced %q, want %q", tt.raw, got, tt.want)
		}
	}
}

// TestURLJoinsPathsCleanly is the reason a trailing slash is normalised away:
// every consumer builds an absolute URL from this value.
func TestURLJoinsPathsCleanly(t *testing.T) {
	for _, raw := range []string{"https://example.internal/base", "https://example.internal/base/"} {
		var u URL
		if err := u.UnmarshalText([]byte(raw)); err != nil {
			t.Fatalf("UnmarshalText(%q) = %v", raw, err)
		}
		if got, want := u.JoinPath("auth", "callback").String(), "https://example.internal/base/auth/callback"; got != want {
			t.Errorf("from %q, JoinPath gave %q, want %q", raw, got, want)
		}
	}
}

// TestURLRoundTripsThroughText keeps a dumped config readable: the URL is one
// string, not the innards of net/url.URL, and it parses back to itself.
func TestURLRoundTripsThroughText(t *testing.T) {
	var u URL
	if err := u.UnmarshalText([]byte("https://example.internal/base")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}

	encoded, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal(URL): %v", err)
	}
	if got, want := string(encoded), `"https://example.internal/base"`; got != want {
		t.Fatalf("json.Marshal(URL) = %s, want %s", got, want)
	}

	var back URL
	if err := json.Unmarshal(encoded, &back); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", encoded, err)
	}
	if back.String() != u.String() {
		t.Errorf("round trip gave %q, want %q", back.String(), u.String())
	}
}

func TestZeroURLIsPrintable(t *testing.T) {
	var u URL

	if !u.IsZero() {
		t.Error("IsZero() = false for the zero URL")
	}
	// A failed load is still printed by whoever reports the failure.
	if got := u.String(); got != "" {
		t.Errorf("String() = %q, want the empty string", got)
	}
	if got := fmt.Sprintf("%v", Config{}); !strings.Contains(got, unsetPlaceholder) {
		t.Errorf("printing the zero Config produced %q", got)
	}
}

func TestURLWrapsTheStandardType(t *testing.T) {
	var u URL
	if err := u.UnmarshalText([]byte("https://example.internal/base")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}

	// Consumers need the stdlib type for anything that takes a *url.URL; this
	// only compiles while URL keeps exposing it.
	hostOf := func(std *url.URL) string { return std.Host }
	if got, want := hostOf(u.URL), "example.internal"; got != want {
		t.Errorf("Host = %q, want %q", got, want)
	}
}

func TestLogLevelMapsToSlog(t *testing.T) {
	tests := map[LogLevel]slog.Level{
		LevelDebug: slog.LevelDebug,
		LevelInfo:  slog.LevelInfo,
		LevelWarn:  slog.LevelWarn,
		LevelError: slog.LevelError,
	}

	for level, want := range tests {
		if got := level.Slog(); got != want {
			t.Errorf("%q.Slog() = %v, want %v", level, got, want)
		}
	}
	if len(tests) != len(logLevels) {
		t.Errorf("%d levels are mapped but %d are accepted; every level needs a slog equivalent",
			len(tests), len(logLevels))
	}
}

func TestCIDRsAcceptsRangesAndBareAddresses(t *testing.T) {
	var trusted CIDRs
	if err := trusted.UnmarshalText([]byte("10.0.0.0/8, 192.168.1.7,2001:db8::/32")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}

	tests := map[string]bool{
		"10.4.5.6":    true,
		"10.0.0.0":    true,
		"192.168.1.7": true,
		"192.168.1.8": false,
		"2001:db8::1": true,
		"2001:db9::1": false,
		"172.16.0.1":  false,
		"127.0.0.1":   false,
		// A dual-stack listener reports an IPv4 peer in this form. An operator
		// who wrote 10.0.0.0/8 means this address too.
		"::ffff:10.4.5.6": true,
	}
	for addr, want := range tests {
		parsed, err := netip.ParseAddr(addr)
		if err != nil {
			t.Fatalf("netip.ParseAddr(%q): %v", addr, err)
		}
		if got := trusted.Contains(parsed); got != want {
			t.Errorf("Contains(%s) = %t, want %t", addr, got, want)
		}
	}
}

// TestCIDRsMasksAHostAddressWrittenAsARange covers the typo that would
// otherwise trust nothing at all: netip.Prefix.Contains reports false for every
// address when the bits outside the mask are set.
func TestCIDRsMasksAHostAddressWrittenAsARange(t *testing.T) {
	var trusted CIDRs
	if err := trusted.UnmarshalText([]byte("10.1.2.3/8")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}
	if !trusted.Contains(netip.MustParseAddr("10.9.9.9")) {
		t.Error(`Contains(10.9.9.9) = false for "10.1.2.3/8"; the range was not masked`)
	}
}

func TestEmptyCIDRsTrustsNobody(t *testing.T) {
	var trusted CIDRs

	if !trusted.IsZero() {
		t.Error("IsZero() = false for an unset list")
	}
	if trusted.Contains(netip.MustParseAddr("10.0.0.1")) {
		t.Error("an unset list contains an address; an unproxied deployment would trust its clients")
	}
	if trusted.Contains(netip.Addr{}) {
		t.Error("an invalid address is contained; a peer whose address did not parse must not be trusted")
	}
}

func TestCIDRsRejectsWhatIsNotAnAddress(t *testing.T) {
	for _, raw := range []string{"not-an-address", "10.0.0.0/33", "10.0.0.0/8,nonsense", ""} {
		var trusted CIDRs
		if err := trusted.UnmarshalText([]byte(raw)); err == nil {
			t.Errorf("UnmarshalText(%q) = nil, want an error", raw)
		}
	}
}

func TestCIDRsRoundTripsThroughText(t *testing.T) {
	var trusted CIDRs
	if err := trusted.UnmarshalText([]byte("10.0.0.0/8,192.168.1.7")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}

	text, err := trusted.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	var again CIDRs
	if err := again.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText(%q): %v", text, err)
	}
	if got, want := again.String(), trusted.String(); got != want {
		t.Errorf("round trip = %q, want %q", got, want)
	}
	if got, want := trusted.String(), "10.0.0.0/8,192.168.1.7/32"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
