package attestation

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

// Simulated keypair for each GPU (in real code, this would be per-GPU)
var gpuPrivateKey *ecdsa.PrivateKey
var gpuPublicKey *ecdsa.PublicKey

// GenerateKeypair creates a new ECDSA keypair for a GPU

func GenerateKeypair() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	// In a real implementation, this would use a hardware-backed key
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	gpuPrivateKey = key
	gpuPublicKey = &key.PublicKey
	return key, &key.PublicKey
}

func GenerateMeasurement(gpuID uint64) []byte {
	// For simulation, hash the GPU ID (in real code, hash firmware/config)
	h := sha256.New()
	h.Write([]byte{byte(gpuID)})
	return h.Sum(nil)
}

func SignReport(report *AttestationReport) ([]byte, error) {
	// Hash the report fields
	h := sha256.New()
	h.Write(report.Measurement)
	h.Write(report.Nonce)
	// ... add other fields as needed
	digest := h.Sum(nil)
	r, s, err := ecdsa.Sign(rand.Reader, gpuPrivateKey, digest)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Digest: %v\n", digest)

	// Serialize r and s into a single byte slice
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

func GenerateAttestationReport(nonce []byte) AttestationReport {
	report := AttestationReport{
		Nonce: nonce,
	}
	report.Measurement = GenerateMeasurement(report.GPUId)
	report.Signature, _ = SignReport(&report)
	return report
}

func VerifyReportSignature(report *AttestationReport, pub *ecdsa.PublicKey) (bool, error) {
	h := sha256.New()
	h.Write(report.Measurement)
	h.Write(report.Nonce)
	digest := h.Sum(nil)

	// Deserialize r and s from the signature
	sigLen := len(report.Signature)
	if sigLen%2 != 0 {
		return false, errors.New("invalid ECDSA signature length")
	}
	r := new(big.Int).SetBytes(report.Signature[:sigLen/2])
	s := new(big.Int).SetBytes(report.Signature[sigLen/2:])

	if !ecdsa.Verify(pub, digest, r, s) {
		return false, errors.New("ECDSA signature verification failed")
	}
	return true, nil
}
