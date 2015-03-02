package local

import (
	"crypto/x509"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/cfssl/signer"
)

const (
	testCaFile         = "testdata/ca.pem"
	testCaKeyFile      = "testdata/ca_key.pem"
	testECDSACaFile    = "testdata/ecdsa256_ca.pem"
	testECDSACaKeyFile = "testdata/ecdsa256_ca_key.pem"
)

var expiry = 1 * time.Minute

// Start a signer with the testing RSA CA cert and key.
func newTestSigner(t *testing.T) (s *Signer) {
	s, err := NewSignerFromFile(testCaFile, testCaKeyFile, nil)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestNewSignerFromFilePolicy(t *testing.T) {
	var CAConfig = &config.Config{
		Signing: &config.Signing{
			Profiles: map[string]*config.SigningProfile{
				"signature": &config.SigningProfile{
					Usage:  []string{"digital signature"},
					Expiry: expiry,
				},
			},
			Default: &config.SigningProfile{
				Usage:        []string{"cert sign", "crl sign"},
				ExpiryString: "43800h",
				Expiry:       expiry,
				CA:           true,
			},
		},
	}
	_, err := NewSignerFromFile(testCaFile, testCaKeyFile, CAConfig.Signing)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewSignerFromFileInvalidPolicy(t *testing.T) {
	var invalidConfig = &config.Config{
		Signing: &config.Signing{
			Profiles: map[string]*config.SigningProfile{
				"invalid": &config.SigningProfile{
					Usage:  []string{"wiretapping"},
					Expiry: expiry,
				},
				"empty": &config.SigningProfile{},
			},
			Default: &config.SigningProfile{
				Usage:  []string{"digital signature"},
				Expiry: expiry,
			},
		},
	}
	_, err := NewSignerFromFile(testCaFile, testCaKeyFile, invalidConfig.Signing)
	if err == nil {
		t.Fatal(err)
	}

	if !strings.Contains(err.Error(), `"code":5200`) {
		t.Fatal(err)
	}
}

func TestNewSignerFromFileNoUsageInPolicy(t *testing.T) {
	var invalidConfig = &config.Config{
		Signing: &config.Signing{
			Profiles: map[string]*config.SigningProfile{
				"invalid": &config.SigningProfile{
					Usage:  []string{},
					Expiry: expiry,
				},
				"empty": &config.SigningProfile{},
			},
			Default: &config.SigningProfile{
				Usage:  []string{"digital signature"},
				Expiry: expiry,
			},
		},
	}
	_, err := NewSignerFromFile(testCaFile, testCaKeyFile, invalidConfig.Signing)
	if err == nil {
		t.Fatal("expect InvalidPolicy error")
	}

	if !strings.Contains(err.Error(), `"code":5200`) {
		t.Fatal(err)
	}
}

func newCustomSigner(t *testing.T, testCaFile, testCaKeyFile string) (s *Signer) {
	s, err := NewSignerFromFile(testCaFile, testCaKeyFile, nil)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestNewSignerFromFile(t *testing.T) {
	newTestSigner(t)
}

const (
	testHostName = "localhost"
)

func testSignFile(t *testing.T, certFile string) ([]byte, error) {
	s := newTestSigner(t)

	pem, err := ioutil.ReadFile(certFile)
	if err != nil {
		t.Fatal(err)
	}

	return s.Sign(signer.SignRequest{Hostname: testHostName, Request: string(pem)})
}

type csrTest struct {
	file    string
	keyAlgo string
	keyLen  int
	// Error checking function
	errorCallback func(*testing.T, error)
}

// A helper function that returns a errorCallback function which expects an error.
func ExpectError() func(*testing.T, error) {
	return func(t *testing.T, err error) {
		if err == nil {
			t.Fatal("Expected error. Got nothing.")
		}
	}
}

var csrTests = []csrTest{
	{
		file:          "testdata/rsa2048.csr",
		keyAlgo:       "rsa",
		keyLen:        2048,
		errorCallback: nil,
	},
	{
		file:          "testdata/rsa3072.csr",
		keyAlgo:       "rsa",
		keyLen:        3072,
		errorCallback: nil,
	},
	{
		file:          "testdata/rsa4096.csr",
		keyAlgo:       "rsa",
		keyLen:        4096,
		errorCallback: nil,
	},
	{
		file:          "testdata/ecdsa256.csr",
		keyAlgo:       "ecdsa",
		keyLen:        256,
		errorCallback: nil,
	},
	{
		file:          "testdata/ecdsa384.csr",
		keyAlgo:       "ecdsa",
		keyLen:        384,
		errorCallback: nil,
	},
	{
		file:          "testdata/ecdsa521.csr",
		keyAlgo:       "ecdsa",
		keyLen:        521,
		errorCallback: nil,
	},
}

func TestSignCSRs(t *testing.T) {
	s := newTestSigner(t)
	hostname := "cloudflare.com"
	for _, test := range csrTests {
		csr, err := ioutil.ReadFile(test.file)
		if err != nil {
			t.Fatal("CSR loading error:", err)
		}
		// It is possible to use different SHA2 algorithm with RSA CA key.
		rsaSigAlgos := []x509.SignatureAlgorithm{x509.SHA1WithRSA, x509.SHA256WithRSA, x509.SHA384WithRSA, x509.SHA512WithRSA}
		for _, sigAlgo := range rsaSigAlgos {
			s.sigAlgo = sigAlgo
			certBytes, err := s.Sign(signer.SignRequest{Hostname: hostname, Request: string(csr)})
			if test.errorCallback != nil {
				test.errorCallback(t, err)
			} else {
				if err != nil {
					t.Fatalf("Expected no error. Got %s. Param %s %d", err.Error(), test.keyAlgo, test.keyLen)
				}
				cert, _ := helpers.ParseCertificatePEM(certBytes)
				if cert.SignatureAlgorithm != s.SigAlgo() {
					t.Fatal("Cert Signature Algorithm does not match the issuer.")
				}
			}
		}
	}
}

func TestECDSASigner(t *testing.T) {
	s := newCustomSigner(t, testECDSACaFile, testECDSACaKeyFile)
	hostname := "cloudflare.com"
	for _, test := range csrTests {
		csr, err := ioutil.ReadFile(test.file)
		if err != nil {
			t.Fatal("CSR loading error:", err)
		}
		// Try all ECDSA SignatureAlgorithm
		SigAlgos := []x509.SignatureAlgorithm{x509.ECDSAWithSHA1, x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512}
		for _, sigAlgo := range SigAlgos {
			s.sigAlgo = sigAlgo
			certBytes, err := s.Sign(signer.SignRequest{Hostname: hostname, Request: string(csr)})
			if test.errorCallback != nil {
				test.errorCallback(t, err)
			} else {
				if err != nil {
					t.Fatalf("Expected no error. Got %s. Param %s %d", err.Error(), test.keyAlgo, test.keyLen)
				}
				cert, _ := helpers.ParseCertificatePEM(certBytes)
				if cert.SignatureAlgorithm != s.SigAlgo() {
					t.Fatal("Cert Signature Algorithm does not match the issuer.")
				}
			}
		}
	}
}

const (
	ecdsaInterCSR = "testdata/ecdsa256-inter.csr"
	ecdsaInterKey = "testdata/ecdsa256-inter.key"
	rsaInterCSR   = "testdata/rsa2048-inter.csr"
	rsaInterKey   = "testdata/rsa2048-inter.key"
)

func TestCAIssuing(t *testing.T) {
	var caCerts = []string{testCaFile, testECDSACaFile}
	var caKeys = []string{testCaKeyFile, testECDSACaKeyFile}
	var interCSRs = []string{ecdsaInterCSR, rsaInterCSR}
	var interKeys = []string{ecdsaInterKey, rsaInterKey}
	var CAPolicy = &config.Signing{
		Default: &config.SigningProfile{
			Usage:        []string{"cert sign", "crl sign"},
			ExpiryString: "1h",
			Expiry:       1 * time.Hour,
			CA:           true,
		},
	}
	var hostname = "cloudflare-inter.com"
	// Each RSA or ECDSA root CA issues two intermediate CAs (one ECDSA and one RSA).
	// For each intermediate CA, use it to issue additional RSA and ECDSA intermediate CSRs.
	for i, caFile := range caCerts {
		caKeyFile := caKeys[i]
		s := newCustomSigner(t, caFile, caKeyFile)
		s.policy = CAPolicy
		for j, csr := range interCSRs {
			csrBytes, _ := ioutil.ReadFile(csr)
			certBytes, err := s.Sign(signer.SignRequest{Hostname: hostname, Request: string(csrBytes)})
			if err != nil {
				t.Fatal(err)
			}
			interCert, err := helpers.ParseCertificatePEM(certBytes)
			if err != nil {
				t.Fatal(err)
			}
			keyBytes, _ := ioutil.ReadFile(interKeys[j])
			interKey, _ := helpers.ParsePrivateKeyPEM(keyBytes)
			interSigner := &Signer{interCert, interKey, CAPolicy, signer.DefaultSigAlgo(interKey)}
			for _, anotherCSR := range interCSRs {
				anotherCSRBytes, _ := ioutil.ReadFile(anotherCSR)
				bytes, err := interSigner.Sign(signer.SignRequest{Hostname: hostname, Request: string(anotherCSRBytes)})
				if err != nil {
					t.Fatal(err)
				}
				cert, err := helpers.ParseCertificatePEM(bytes)
				if err != nil {
					t.Fatal(err)
				}
				if cert.SignatureAlgorithm != interSigner.SigAlgo() {
					t.Fatal("Cert Signature Algorithm does not match the issuer.")
				}
			}
		}
	}

}

const testCSR = "testdata/ecdsa256.csr"

func TestOverrideSubject(t *testing.T) {
	csrPEM, err := ioutil.ReadFile(testCSR)
	if err != nil {
		t.Fatalf("%v", err)
	}

	req := &signer.Subject{
		Hosts: []string{"127.0.0.1"},
		Names: []csr.Name{
			{O: "example.net"},
		},
	}

	s := newCustomSigner(t, testECDSACaFile, testECDSACaKeyFile)

	certPEM, err := s.Sign(signer.SignRequest{Hostname: "localhost", Request: string(csrPEM), Subject: req})
	if err != nil {
		t.Fatalf("%v", err)
	}

	cert, err := helpers.ParseCertificatePEM(certPEM)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if cert.Subject.Organization[0] != "example.net" {
		t.Fatalf("Failed to override subject: want example.net but have %s", cert.Subject.Organization[0])
	}

	log.Info("Overrode subject info")
}
