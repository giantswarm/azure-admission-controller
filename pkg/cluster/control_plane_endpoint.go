package cluster

import (
	"reflect"

	"github.com/giantswarm/microerror"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/pkg/key"
)

func validateControlPlaneEndpoint(cluster capiv1alpha3.Cluster, baseDomain string) error {
	host := key.GetControlPlaneEndpointHost(cluster.Name, baseDomain)
	if cluster.Spec.ControlPlaneEndpoint.Host != host {
		return microerror.Maskf(errors.InvalidOperationError, "ControlPlaneEndpoint.Port can only be set to %s", host)
	}

	if cluster.Spec.ControlPlaneEndpoint.Port != key.ControlPlaneEndpointPort {
		return microerror.Maskf(errors.InvalidOperationError, "ControlPlaneEndpoint.Port can only be set to %d", key.ControlPlaneEndpointPort)
	}

	return nil
}

func validateControlPlaneEndpointUnchanged(old capiv1alpha3.Cluster, new capiv1alpha3.Cluster) error {
	if reflect.DeepEqual(old.Spec.ControlPlaneEndpoint, new.Spec.ControlPlaneEndpoint) {
		return microerror.Maskf(errors.InvalidOperationError, "ControlPlaneEndpoint can't be changed.")
	}

	return nil
}
