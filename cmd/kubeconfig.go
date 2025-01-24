package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/spf13/cobra"
	certsv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const PREAMBLE_CERTIFICATE_REQUEST = "CERTIFICATE REQUEST"

var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Generate a Kube-Config file for a user with a given name",
	Long:  ``,
	Run:   execute,
}

func execute(cmd *cobra.Command, args []string) {
	fmt.Println("kubeconfig called")
}

func init() {
	rootCmd.AddCommand(kubeconfigCmd)
	// kubeconfigCmd.PersistentFlags().
}

type cmdctx struct {
	User         string
	Groups       []string
	key          *rsa.PrivateKey
	keyPem       []byte
	keyPemBase64 string
	csrPem       []byte
	csrPemBase64 string
}

func createPrivateKey(ctx *cmdctx) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	ctx.key = key

	// pub := key.Public()

	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	ctx.keyPem = keyPem
	ctx.keyPemBase64 = base64.StdEncoding.EncodeToString(keyPem)

	fmt.Println("base64 Encoded RSA Private Key")
	fmt.Println(ctx.keyPemBase64)

	// pubPem := pem.EncodeToMemory(
	// 	&pem.Block{
	// 		Type:  "RSA PUBLIC KEY",
	// 		Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
	// 	},
	// )

	// Write private key to file
	// if err := os.WriteFile(user+".rsa", privPem, 0700); err != nil {
	// 	panic(err)
	// }
	// Write public key to file
	// if err := os.WriteFile(user+".rsa.pub", pubPem, 0755); err != nil {
	// 	panic(err)
	// }
	// return key
}

func prepareCertificateSigningRequest(ctx *cmdctx) {
	subj := pkix.Name{
		CommonName:   ctx.User,
		Organization: ctx.Groups,
	}

	rawSubj := subj.ToRDNSequence()
	asn1Subj, _ := asn1.Marshal(rawSubj)

	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, ctx.key)

	// csrFile, err := os.Create("bellman.csr")
	// if err != nil {
	// 	panic(err)
	// }
	// defer csrFile.Close()

	ctx.csrPem = pem.EncodeToMemory(&pem.Block{Type: PREAMBLE_CERTIFICATE_REQUEST, Bytes: csrBytes})
	ctx.csrPemBase64 = base64.StdEncoding.EncodeToString(ctx.csrPem)
	fmt.Println("base64 Encoded Certificate Signing Request")
	fmt.Println(ctx.csrPemBase64)
}

func applyCertificateSigningRequest(clientset *kubernetes.Clientset, ctx cmdctx) {

	csr := &certsv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: ctx.User,
		},
		Spec: certsv1.CertificateSigningRequestSpec{
			Groups:     ctx.Groups,
			SignerName: "kubernetes.io/kube-apiserver-client",
			Usages:     []certsv1.KeyUsage{certsv1.UsageClientAuth},
			Request:    ctx.csrPem,
		},
	}
	clientset.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, metav1.CreateOptions{})
}

//
