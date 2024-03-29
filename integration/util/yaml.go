package util

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	corev1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capzexp "sigs.k8s.io/cluster-api-provider-azure/exp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func CreateCRsInFolder(ctx context.Context, client client.Client, crsFolder string) error {
	return filepath.Walk(crsFolder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
				return nil
			}

			bs, err := ioutil.ReadFile(path)
			if err != nil {
				return microerror.Mask(err)
			}

			o, err := unmarshal(bs)
			if err != nil {
				return microerror.Mask(err)
			}

			err = client.Create(ctx, o)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	})
}

func DeleteCRsInFolder(ctx context.Context, client client.Client, crsFolder string) error {
	return filepath.Walk(crsFolder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
				return nil
			}

			bs, err := ioutil.ReadFile(path)
			if err != nil {
				return microerror.Mask(err)
			}

			o, err := unmarshal(bs)
			if err != nil {
				return microerror.Mask(err)
			}

			err = client.Delete(ctx, o)
			if apierrors.IsNotFound(err) {
				// Ok
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	})
}

func unmarshal(bs []byte) (client.Object, error) {
	var err error
	var obj client.Object

	// First parse kind.
	typeMeta := &metav1.TypeMeta{}
	err = yaml.Unmarshal(bs, typeMeta)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Then construct correct CR object.
	switch typeMeta.Kind {
	case "Namespace":
		obj = new(corev1.Namespace)
	case "Organization":
		obj = new(securityv1alpha1.Organization)
	case "Cluster":
		obj = new(capi.Cluster)
	case "MachinePool":
		obj = new(capiexp.MachinePool)
	case "AzureCluster":
		obj = new(capz.AzureCluster)
	case "AzureMachine":
		obj = new(capz.AzureMachine)
	case "AzureMachinePool":
		obj = new(capzexp.AzureMachinePool)
	case "Release":
		obj = new(releasev1alpha1.Release)
	case "Spark":
		obj = new(corev1alpha1.Spark)
	default:
		return nil, microerror.Maskf(unknownKindError, "error while unmarshalling the CR read from file, kind: %s", typeMeta.Kind)
	}

	// ...and unmarshal the whole object.
	err = yaml.Unmarshal(bs, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return obj, nil
}
