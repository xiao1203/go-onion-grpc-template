package auth

import (
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "math/big"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/newmo-oss/ergo"
)

type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksDoc struct {
	Keys []jwksKey `json:"keys"`
}

type JWKSCache struct {
	url     string
	ttl     time.Duration
	mu      sync.RWMutex
	expires time.Time
	keys    map[string]*rsa.PublicKey
	client  *http.Client
}

func NewJWKSCache(url string, ttl time.Duration) *JWKSCache {
	return &JWKSCache{
		url:    url,
		ttl:    ttl,
		keys:   map[string]*rsa.PublicKey{},
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *JWKSCache) KeyFor(kid string) (*rsa.PublicKey, error) {
    c.mu.RLock()
    if k, ok := c.keys[kid]; ok && time.Now().Before(c.expires) {
        c.mu.RUnlock()
        return k, nil
    }
	c.mu.RUnlock()
	// refresh
	if err := c.refresh(); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
    if k, ok := c.keys[kid]; ok {
        return k, nil
    }
    return nil, ergo.NewSentinel("jwks: key not found")
}

func (c *JWKSCache) refresh() error {
    resp, err := c.client.Get(c.url)
    if err != nil {
        return err
    }
    defer func() { _ = resp.Body.Close() }()
	var doc jwksDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return err
	}
	m := map[string]*rsa.PublicKey{}
	for _, k := range doc.Keys {
		if !strings.EqualFold(k.Kty, "RSA") {
			continue
		}
		pub, err := jwkToRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		m[k.Kid] = pub
	}
	c.mu.Lock()
	c.keys = m
	c.expires = time.Now().Add(c.ttl)
	c.mu.Unlock()
	return nil
}

func jwkToRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	n := new(big.Int).SetBytes(nBytes)
	return &rsa.PublicKey{N: n, E: e}, nil
}
