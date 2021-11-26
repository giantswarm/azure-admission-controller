package vmcapabilities

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type API interface {
	List(ctx context.Context, filter string) (map[string]compute.ResourceSku, error)
}

type Factory interface {
	GetClient(context.Context, client.Client, v1.ObjectMeta) (*VMSKU, error)
}
