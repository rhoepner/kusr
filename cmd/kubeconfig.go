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
	"kusr/k8s"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	certsv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const PREAMBLE_CERTIFICATE_REQUEST = "CERTIFICATE REQUEST"
const NS_KUBE_PUBLIC = "kube-public"
const CM_KUBE_ROOT_CA_CRT = "kube-root-ca.crt"
const NAMED = "pske-non-prod"

var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Generate a Kube-Config file for a user with a given name",
	Long:  ``,
	Run:   execute,
}

func execute(cmd *cobra.Command, args []string) {
	var name = args[0]
	var clientset, err = k8s.Connect()
	if err != nil {
		panic(err)
	}
	ctx := &cmdctx{
		User:   name,
		Groups: []string{"dev"},
	}

	preparePrivateKey(cmd, ctx)
	prepareCertificateSigningRequest(cmd, ctx)
	prepareCertificateAuthorityData(clientset, ctx)
	prepareCertificate(cmd, clientset, ctx)
	createKubeConfig(cmd, ctx)
}

func init() {
	RootCmd.AddCommand(kubeconfigCmd)
	kubeconfigCmd.Flags().Bool("trace", false, "Set to true to write intermediate files to disks")
	kubeconfigCmd.Flags().Bool("noproxy", false, "Set to true to not write the innovas proxy into kubeconfig")
}

type cmdctx struct {
	User                       string
	Groups                     []string
	key                        *rsa.PrivateKey
	keyPem                     []byte
	keyPemBase64               string
	csrPem                     []byte
	csrPemBase64               string
	crtPem                     []byte
	crtPemBase64               string
	certificateAuthority       string
	certificateAuthorityBase64 string
}

func preparePrivateKey(cmd *cobra.Command, ctx *cmdctx) {
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

	trace, _ := cmd.Flags().GetBool("trace")

	if trace {
		if err := os.WriteFile(ctx.User+".key", ctx.keyPem, 0660); err != nil {
			panic(err)
		}
	}

	fmt.Println("RSA Private Key Created!")
}

func prepareCertificateSigningRequest(cmd *cobra.Command, ctx *cmdctx) {
	subj := pkix.Name{
		CommonName: ctx.User,
	}

	rawSubj := subj.ToRDNSequence()
	asn1Subj, _ := asn1.Marshal(rawSubj)

	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, ctx.key)

	ctx.csrPem = pem.EncodeToMemory(&pem.Block{Type: PREAMBLE_CERTIFICATE_REQUEST, Bytes: csrBytes})
	ctx.csrPemBase64 = base64.StdEncoding.EncodeToString(ctx.csrPem)

	trace, _ := cmd.Flags().GetBool("trace")

	if trace {
		if err := os.WriteFile(ctx.User+".csr", ctx.keyPem, 0660); err != nil {
			panic(err)
		}
	}
	fmt.Println("Certificate Signing Request created!")
}

func prepareCertificate(cmd *cobra.Command, clientset *kubernetes.Clientset, ctx *cmdctx) {

	var expirationSeconds int32 = 31557600

	csr := &certsv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: ctx.User,
		},
		Spec: certsv1.CertificateSigningRequestSpec{
			SignerName:        "kubernetes.io/kube-apiserver-client",
			Usages:            []certsv1.KeyUsage{certsv1.UsageClientAuth},
			Request:           ctx.csrPem,
			ExpirationSeconds: &expirationSeconds,
		},
	}
	csrBeforeApproval, err := clientset.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	csrBeforeApproval.Status.Conditions = append(csrBeforeApproval.Status.Conditions, certsv1.CertificateSigningRequestCondition{
		Type:           certsv1.CertificateApproved,
		Status:         v1.ConditionTrue,
		Reason:         "Give Developer Access to Cluster",
		Message:        "Approced by kusr",
		LastUpdateTime: metav1.Now(),
	})

	_, err = clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.TODO(), csr.ObjectMeta.Name, csrBeforeApproval, metav1.UpdateOptions{})
	if err != nil {
		panic(err)
	}
	csrWithCert, err := clientset.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), ctx.User, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	ctx.crtPem = csrWithCert.Status.Certificate
	ctx.crtPemBase64 = base64.StdEncoding.EncodeToString(ctx.crtPem)

	trace, _ := cmd.Flags().GetBool("trace")

	if trace {
		if err := os.WriteFile(ctx.User+".crt", ctx.keyPem, 0660); err != nil {
			panic(err)
		}
	}
	fmt.Println("Certificate created!")

}

func prepareCertificateAuthorityData(clientset *kubernetes.Clientset, ctx *cmdctx) {
	configMap, err := clientset.CoreV1().ConfigMaps(NS_KUBE_PUBLIC).Get(context.TODO(), CM_KUBE_ROOT_CA_CRT, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	ctx.certificateAuthority = configMap.Data["ca.crt"]
	ctx.certificateAuthorityBase64 = base64.StdEncoding.EncodeToString([]byte(ctx.certificateAuthority))
}

func createKubeConfig(cmd *cobra.Command, ctx *cmdctx) {
	kubeconfig := KCConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Clusters: []KCNamedCluster{
			{
				Name: NAMED,
				Cluster: KCCluster{
					CertificateAuthorityData: ctx.certificateAuthorityBase64,
					Server:                   "https://api.non-prod.426110.projects.prod.gardener.get-cloud.io",
				},
			},
		},
		Contexts: []KCNamedContext{
			{
				Name: NAMED,
				Context: KCContext{
					Cluster: NAMED,
					User:    ctx.User,
				},
			},
		},
		CurrentContext: NAMED,
		Users: []KCNamedUser{
			{
				Name: ctx.User,
				User: KCUser{
					ClientCertificateData: ctx.crtPemBase64,
					ClientKeyData:         ctx.keyPemBase64,
				},
			},
		},
	}

	noproxy, _ := cmd.Flags().GetBool("noproxy")

	if !noproxy {
		kubeconfig.Clusters[0].Cluster.ProxyUrl = "http://proxy-k.innovas.de:3128"
	}

	buf, err := yaml.Marshal(kubeconfig)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(ctx.User+"--non-prod.yaml", buf, 0700); err != nil {
		panic(err)
	}
}

type KCConfig struct {
	ApiVersion     string           `yaml:"apiVersion,omitempty"`
	Kind           string           `yaml:"kind,omitempty"`
	Clusters       []KCNamedCluster `yaml:"clusters,omitempty"`
	Contexts       []KCNamedContext `yaml:"contexts,omitempty"`
	Users          []KCNamedUser    `yaml:"users,omitempty"`
	CurrentContext string           `yaml:"current-context"`
}

type KCNamedCluster struct {
	// Name is the nickname for this Cluster
	Name string `yaml:"name"`
	// Cluster holds the cluster information
	Cluster KCCluster `yaml:"cluster"`
}

type KCCluster struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
	ProxyUrl                 string `yaml:"proxy-url,omitempty"`
}

type KCNamedContext struct {
	Name string `yaml:"name"`
	// Cluster holds the cluster information
	Context KCContext `yaml:"context"`
}

type KCContext struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type KCNamedUser struct {
	Name string `yaml:"name"`
	User KCUser `yaml:"user"`
}

type KCUser struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}
