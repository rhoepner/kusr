package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePrivateKey(t *testing.T) {
	assertions := assert.New(t)
	ctx := &cmdctx{
		User:   "bellmanr",
		Groups: []string{"dev", "ops", "admin"},
	}
	createPrivateKey(ctx)
	assertions.NotNil(ctx.key, "PrimaryKey has not been generated")
	assertions.NotNil(ctx.keyPem, "PrimaryKey PEM has not been generated")
	assertions.NotNil(ctx.keyPemBase64, "PrimaryKey PEM Base64 has not been generated")
}

func Test_prepareCertificateSigningRequest(t *testing.T) {
	assertions := assert.New(t)

	// var clientset, err = k8s.Connect()
	// if err != nil {
	// 	panic(err)
	// }

	ctx := &cmdctx{
		User:   "bellmanr",
		Groups: []string{"dev", "ops", "admin"},
	}
	createPrivateKey(ctx)
	prepareCertificateSigningRequest(ctx)
	assertions.NotNil(ctx.csrPem, "CertificateSigningRequest PEM has not been generated")
	assertions.NotNil(ctx.csrPemBase64, "CertificateSigningRequest PEM Base64 has not been generated")
}
