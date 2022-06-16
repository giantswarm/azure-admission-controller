package generic

import (
	"github.com/giantswarm/apiextensions/v6/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

type GenericObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func newObjectWithOrganization(org *string) metav1.Object {
	obj := &GenericObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ab123",
			Namespace: "default",
			Labels: map[string]string{
				label.AzureOperatorVersion:        "5.0.0",
				label.Cluster:                     "ab123",
				capi.ClusterLabelName:             "ab123",
				capi.MachineControlPlaneLabelName: "true",
				label.MachinePool:                 "ab123",
				label.ReleaseVersion:              "13.0.0",
			},
		},
	}

	if org != nil {
		obj.Labels[label.Organization] = *org
	}

	return obj
}
