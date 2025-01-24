package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"os"
)

func GenerateKey(user string, bitSize int) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		panic(err)
	}

	pub := key.Public()

	privPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	pubPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)

	// Write private key to file
	if err := os.WriteFile(user+".rsa", privPem, 0700); err != nil {
		panic(err)
	}
	// Write public key to file
	if err := os.WriteFile(user+".rsa.pub", pubPem, 0755); err != nil {
		panic(err)
	}
	return key
}

func GenerateCSR(pk *rsa.PrivateKey, user string, groups []string) {
	subj := pkix.Name{
		CommonName:   user,
		Organization: groups,
	}

	rawSubj := subj.ToRDNSequence()
	asn1Subj, _ := asn1.Marshal(rawSubj)

	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, pk)

	csrFile, err := os.Create("bellman.csr")
	if err != nil {
		panic(err)
	}
	defer csrFile.Close()

	pem.Encode(csrFile, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
}
