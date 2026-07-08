package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// GenerateIdentityCert generates a self-signed ECDSA private key and X.509 certificate.
// Returns the private key PEM, certificate PEM, and a short 4-byte fingerprint string.
func GenerateIdentityCert(deviceID string) (string, string, string, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", "", err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", "", err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: deviceID,
		},
		NotBefore:             time.Now().Add(-10 * time.Minute), // Buffer for clock drift
		NotAfter:              time.Now().AddDate(10, 0, 0),      // 10 years
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", "", err
	}

	// PEM encode private key
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", "", err
	}
	privBlock := &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}
	privPEM := pem.EncodeToMemory(privBlock)

	// PEM encode certificate
	certBlock := &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}
	certPEM := pem.EncodeToMemory(certBlock)

	// Compute SHA-256 fingerprint (we use the first 4 bytes formatted as hex for a friendly verifier)
	hash := sha256.Sum256(derBytes)
	fingerprintParts := make([]string, 4)
	for i := 0; i < 4; i++ {
		fingerprintParts[i] = fmt.Sprintf("%02X", hash[i])
	}
	fingerprint := strings.Join(fingerprintParts, ":")

	return string(privPEM), string(certPEM), fingerprint, nil
}
