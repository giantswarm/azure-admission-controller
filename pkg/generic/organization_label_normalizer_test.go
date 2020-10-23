package generic

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/to"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/azure-admission-controller/pkg/mutator"
)

type GenericObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
}

func Test_EnsureOrganizationLabelNormalized(t *testing.T) {
	testCases := []struct {
		name          string
		input         metav1.Object
		expectedPatch *mutator.PatchOperation
		errorMatcher  func(error) bool
	}{
		{
			name:          "case 0: no need for changes",
			input:         newObjectWithOrganization(to.StringP("giantswarm")),
			expectedPatch: nil,
			errorMatcher:  nil,
		},
		{
			name:          "case 1: lowercase uppercase letters",
			input:         newObjectWithOrganization(to.StringP("GiantSwarm")),
			expectedPatch: &mutator.PatchOperation{Operation: "replace", Path: "/metadata/labels/giantswarm.io~1organization", Value: "giantswarm"},
			errorMatcher:  nil,
		},
		{
			name:          "case 2: lowercase uppercase letters combined with dashes",
			input:         newObjectWithOrganization(to.StringP("FOO-Pre-Production-Shipment-Team")),
			expectedPatch: &mutator.PatchOperation{Operation: "replace", Path: "/metadata/labels/giantswarm.io~1organization", Value: "foo-pre-production-shipment-team"},
			errorMatcher:  nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			patch, err := EnsureOrganizationLabelNormalized(context.Background(), tc.input)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !cmp.Equal(patch, tc.expectedPatch) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedPatch, patch))
			}
		})
	}
}
