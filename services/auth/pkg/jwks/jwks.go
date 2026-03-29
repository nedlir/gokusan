package jwks

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
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

// JWKSClient fetches and caches JWKS keys in Redis under the key pattern jwks:<realm>.
type JWKSClient struct {
	url      string
	realm    string
	cacheTTL time.Duration
	redis    *redis.Client
}

func NewJWKSClient(url, realm string, cacheTTL time.Duration, redisClient *redis.Client) *JWKSClient {
	return &JWKSClient{
		url:      url,
		realm:    realm,
		cacheTTL: cacheTTL,
		redis:    redisClient,
	}
}

func (j *JWKSClient) GetKey(kid string) (*rsa.PublicKey, error) {
	keys, err := j.loadKeys()
	if err != nil {
		return nil, err
	}
	key, ok := keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid '%s' not found", kid)
	}
	return key, nil
}

// loadKeys returns the parsed RSA keys, reading from Redis cache first and
// falling back to a live JWKS fetch on a cache miss.
func (j *JWKSClient) loadKeys() (map[string]*rsa.PublicKey, error) {
	ctx := context.Background()
	cacheKey := "jwks:" + j.realm

	cached, err := j.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		return j.parseJWKSBytes(cached)
	}
	if err != redis.Nil {
		log.Printf("Redis JWKS cache read error: %v — falling back to live fetch", err)
	}

	// Cache miss: fetch from Keycloak
	body, err := j.fetchRaw()
	if err != nil {
		return nil, err
	}

	// Store raw JSON in Redis with TTL
	if setErr := j.redis.Set(ctx, cacheKey, body, j.cacheTTL).Err(); setErr != nil {
		log.Printf("Failed to cache JWKS in Redis: %v", setErr)
	}

	return j.parseJWKSBytes(body)
}

func (j *JWKSClient) fetchRaw() ([]byte, error) {
	resp, err := http.Get(j.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %v", err)
	}

	log.Printf("Fetched fresh JWKS from %s", j.url)
	return body, nil
}

func (j *JWKSClient) parseJWKSBytes(data []byte) (map[string]*rsa.PublicKey, error) {
	var jwks JWKS
	if err := json.Unmarshal(data, &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %v", err)
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}
		pub, err := parseRSAKey(key)
		if err != nil {
			log.Printf("Failed to parse RSA key %s: %v", key.Kid, err)
			continue
		}
		keys[key.Kid] = pub
	}
	return keys, nil
}

func parseRSAKey(key JWK) (*rsa.PublicKey, error) {
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
