package k8s

import (
	"context"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCSR(t *testing.T) {
	var clientset, err = Connect()
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
