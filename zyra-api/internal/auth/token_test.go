package auth

import (
	"testing"
	"time"
)

func TestAccessExpiryAndSessionDeadline(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	tokens := New("test-secret-that-is-long-enough-for-signing")
	tokens.now = func() time.Time { return now }
	raw, exp, err := tokens.Issue("user", "session", now.Add(7*24*time.Hour))
	if err != nil || !exp.Equal(now.Add(15*time.Minute)) {
		t.Fatalf("expiry: %v %v", exp, err)
	}
	if _, err = tokens.Parse(raw); err != nil {
		t.Fatal(err)
	}
	now = now.Add(15 * time.Minute)
	if _, err = tokens.Parse(raw); err == nil {
		t.Fatal("expired access token accepted")
	}
	deadline := now.Add(2 * time.Minute)
	_, exp, err = tokens.Issue("user", "session", deadline)
	if err != nil || !exp.Equal(deadline) {
		t.Fatal("access escaped session deadline")
	}
	now = deadline
	if _, _, err = tokens.Issue("user", "session", deadline); err == nil {
		t.Fatal("expired session accepted")
	}
}
