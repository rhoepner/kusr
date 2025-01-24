package cmd

import (
	"context"
	"fmt"
	"kusr/k8s"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratePrivateKey(t *testing.T) {
	pem := GenerateKey("bellmanr", 2048)
	if pem == nil {
		t.Fatalf("Failed to Generate Private Key")
	}
	t.Logf("PK %v", pem)
}

func TestGenerateCSR(t *testing.T) {
	key := GenerateKey("bellmanr", 2048)
	GenerateCSR(key, "bellmanr", []string{"dev"})
}

func TestCSR(t *testing.T) {
	var clientset, err = k8s.Connect()
	if err != nil {
		panic(err)
	}
	clusterRoleList, err := clientset.RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, clusterRole := range clusterRoleList.Items {
		fmt.Println(clusterRole.ObjectMeta.Name)
	}
}
