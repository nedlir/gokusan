package jwks

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKSClient struct {
	url       string
	keys      map[string]*rsa.PublicKey
	mutex     sync.RWMutex
	lastFetch time.Time
	cacheTTL  time.Duration
}

func NewJWKSClient(url string, cacheTTL time.Duration) *JWKSClient {
	return &JWKSClient{
		url:      url,
		keys:     make(map[string]*rsa.PublicKey),
		cacheTTL: cacheTTL,
	}
}

func (j *JWKSClient) GetKey(kid string) (*rsa.PublicKey, error) {
	j.mutex.RLock()
	if key, exists := j.keys[kid]; exists && time.Since(j.lastFetch) < j.cacheTTL {
		j.mutex.RUnlock()
		return key, nil
	}
	j.mutex.RUnlock()

	if err := j.fetchKeys(); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}

	j.mutex.RLock()
	defer j.mutex.RUnlock()
	if key, exists := j.keys[kid]; exists {
		return key, nil
	}

	return nil, fmt.Errorf("key with kid '%s' not found", kid)
}

func (j *JWKSClient) fetchKeys() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if time.Since(j.lastFetch) < j.cacheTTL {
		return nil
	}

	resp, err := http.Get(j.url)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %v", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %v", err)
	}

	j.keys = make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}
		publicKey, err := j.parseRSAKey(key)
		if err != nil {
			log.Printf("Failed to parse RSA key %s: %v", key.Kid, err)
			continue
		}
		j.keys[key.Kid] = publicKey
	}

	j.lastFetch = time.Now()
	log.Printf("Successfully fetched %d keys from JWKS", len(j.keys))
	return nil
}

func (j *JWKSClient) parseRSAKey(key JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %v", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %v", err)
	}
	exponent := int(big.NewInt(0).SetBytes(eBytes).Int64())
	return &rsa.PublicKey{N: big.NewInt(0).SetBytes(nBytes), E: exponent}, nil
}
