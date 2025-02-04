package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a Key-Pair to be used along with kubecfg",
	Long:  ``,
	Run:   executeInit,
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().Int("bitsize", 2048, "Specify the bitsize to be used for the Key-Pair. Defaults to 2048")
	// kubeconfigCmd.Flags().Bool("noproxy", false, "Set to true to not write the innovas proxy into kubeconfig")
}

func executeInit(cmd *cobra.Command, args []string) {
	bitsize, _ := cmd.Flags().GetInt("bitsize")
	key, err := rsa.GenerateKey(rand.Reader, bitsize)
	if err != nil {
		panic(err)
	}

	// Extract public component.
	pub := key.Public()

	// Encode private key to PKCS#1 ASN.1 PEM
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	// Encode public key to PKCS#1 ASN.1 PEM.
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)

	// ctx.keyPemBase64 = base64.StdEncoding.EncodeToString(keyPem)

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	keyfile := filepath.Join(home, ".kube", "kubecfg.key")

	if err := os.WriteFile(keyfile, keyPEM, 0700); err != nil {
		panic(err)
	}
	cmd.Printf("RSA Private Key Created at %v", keyfile)

	pubfile := filepath.Join(home, ".kube", "kubecfg.pub")
	if err := os.WriteFile(pubfile, pubPEM, 0700); err != nil {
		panic(err)
	}
	cmd.Printf("RSA Private Key Created at %v", keyfile)

}
