package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}
type Tokens struct {
	secret []byte
	now    func() time.Time
}

func New(secret string) *Tokens { return &Tokens{[]byte(secret), time.Now} }
func Random() string            { return rand.Text() }
func Hash(s string) string      { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }
func (t *Tokens) Issue(userID, sessionID string, deadline time.Time) (string, time.Time, error) {
	now := t.now().UTC()
	if !now.Before(deadline) {
		return "", time.Time{}, errors.New("session expired")
	}
	exp := now.Add(15 * time.Minute)
	if deadline.Before(exp) {
		exp = deadline
	}
	c := Claims{SessionID: sessionID, RegisteredClaims: jwt.RegisteredClaims{Subject: userID, Issuer: "zyra-api", Audience: jwt.ClaimStrings{"zyra-web"}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(exp)}}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(t.secret)
	return raw, exp, err
}
func (t *Tokens) Parse(raw string) (*Claims, error) {
	c := &Claims{}
	_, err := jwt.ParseWithClaims(raw, c, func(_ *jwt.Token) (any, error) { return t.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("zyra-api"), jwt.WithAudience("zyra-web"), jwt.WithExpirationRequired(), jwt.WithTimeFunc(t.now))
	if err != nil {
		return nil, err
	}
	if c.Subject == "" || c.SessionID == "" {
		return nil, errors.New("invalid claims")
	}
	return c, nil
}
