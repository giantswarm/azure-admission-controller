package azureupdate

import (
	"context"
	"encoding/json"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/azure-admission-controller/internal/errors"
	"github.com/giantswarm/azure-admission-controller/internal/releaseversion"
)

var (
	controlPlaneName      = "gmk24"
	controlPlaneNameSpace = "default"
)

func TestAzureConfigValidate(t *testing.T) {
	releases := []string{"11.3.0", "11.3.1", "11.4.0", "12.0.0"}

	testCases := []struct {
		name         string
		ctx          context.Context
		releases     []string
		oldVersion   string
		newVersion   string
		errorMatcher func(err error) bool
	}{
		{
			name: "case 0",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.3.1",
			errorMatcher: nil,
		},
		{
			name: "case 1",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.4.0",
			errorMatcher: nil,
		},
		{
			name: "case 2",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "12.0.0",
			errorMatcher: releaseversion.IsSkippingReleaseError,
		},
		{
			name: "case 3",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.3.0",
			errorMatcher: nil,
		},
		{
			name: "case 4",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "11.4.0",
			errorMatcher: nil,
		},
		{
			name: "case 5",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "",
			errorMatcher: errors.IsParsingFailed,
		},
		{
			name: "case 6",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "",
			newVersion:   "11.3.1",
			errorMatcher: errors.IsParsingFailed,
		},
		{
			name: "case 7",
			ctx:  context.Background(),

			releases:     []string{"invalid"},
			oldVersion:   "11.3.0",
			newVersion:   "11.4.0",
			errorMatcher: errors.IsInvalidReleaseError,
		},
		{
			name: "case 8",
			ctx:  context.Background(),

			releases:     []string{"invalid"},
			oldVersion:   "11.3.0",
			newVersion:   "11.3.1",
			errorMatcher: errors.IsInvalidReleaseError,
		},
		{
			name: "case 9",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "11.3.0",
<<<<<<< HEAD
			errorMatcher: errors.IsInvalidOperationError,
=======
			conditions:   []string{},
			errorMatcher: releaseversion.IsDowngradingIsNotAllowedError,
>>>>>>> master
		},
		{
			name: "case 10",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.0.0", // does not exist
			newVersion:   "11.3.0", // exists
			errorMatcher: nil,
		},
		{
			name: "case 11",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0", // exists
			newVersion:   "11.5.0", // does not exist
<<<<<<< HEAD
			errorMatcher: errors.IsInvalidReleaseError,
=======
			conditions:   []string{},
			errorMatcher: releaseversion.IsReleaseNotFoundError,
>>>>>>> master
		},
		{
			name: "case 12",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.5.0", // does not exist
			newVersion:   "11.5.0", // does not exist
			errorMatcher: nil,
		},
		{
			name: "case 13",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0",
			newVersion:   "11.4.0",
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Create a new logger that is used by all admitters.
			var newLogger micrologger.Logger
			{
				newLogger, err = micrologger.New(micrologger.Config{})
				if err != nil {
					panic(microerror.JSON(err))
				}
			}
			fakeCtrlClient, err := getFakeCtrlClient()
			if err != nil {
				panic(microerror.JSON(err))
			}

			admit := &AzureConfigValidator{
				ctrlClient: fakeCtrlClient,
				logger:     newLogger,
			}

			// Create needed releases.
			err = ensureReleases(fakeCtrlClient, tc.releases)
			if err != nil {
				t.Fatal(err)
			}

			// Run admission request to validate AzureConfig updates.
			err = admit.Validate(tc.ctx, getAdmissionRequest(tc.oldVersion, tc.newVersion))

			// Check if the error is the expected one.
			switch {
			case err == nil && tc.errorMatcher == nil:
				// fall through
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("expected %#v got %#v", nil, err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("expected %#v got %#v", "error", nil)
			case !tc.errorMatcher(err):
				t.Fatalf("unexpected error: %#v", err)
			}
		})
	}
}

func getAdmissionRequest(oldVersion string, newVersion string) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AzureClusterUpgrade",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "provider.giantswarm.io/v1alpha1",
			Resource: "azureconfigs",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    azureConfigRawObj(newVersion),
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    azureConfigRawObj(oldVersion),
			Object: nil,
		},
	}

	return req
}

func azureConfigRawObj(version string) []byte {
	azureconfig := providerv1alpha1.AzureConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureConfig",
			APIVersion: "provider.giantswarm.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/control-plane":   controlPlaneName,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": version,
			},
		},
		Spec: providerv1alpha1.AzureConfigSpec{
			Cluster:       providerv1alpha1.Cluster{},
			Azure:         providerv1alpha1.AzureConfigSpecAzure{},
			VersionBundle: providerv1alpha1.AzureConfigSpecVersionBundle{},
		},
	}
	byt, _ := json.Marshal(azureconfig)
	return byt
}
