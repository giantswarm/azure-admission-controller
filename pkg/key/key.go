package key

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ControlPlaneEndpointPort  = 443
	ClusterNetworkServiceCIDR = "172.31.0.0/16"
)

var (
	capiErrorMessageRegexp = regexp.MustCompile(`(.*is invalid:\s)(\[)?(.*?)(\]|$)`)
)

func GetControlPlaneEndpointHost(clusterName string, baseDomain string) string {
	return fmt.Sprintf("api.%s.%s", clusterName, baseDomain)
}

func ServiceDomain() string {
	return "cluster.local"
}

func IgnoreCAPIErrorForField(field string, err error) error {
	if status := apierrors.APIStatus(nil); errors.As(err, &status) {
		errStatus := status.Status()
		if errStatus.Reason != "Invalid" {
			return err
		}

		if errStatus.Details == nil {
			return err
		}

		// Remove any errors for the given field.
		var causes []metav1.StatusCause
		for _, cause := range errStatus.Details.Causes {
			if cause.Field != field {
				causes = append(causes, cause)
			}
		}

		if len(causes) < 1 {
			// No errors left, all clear.
			return nil
		}

		errStatus.Details.Causes = causes

		// Remove any errors for this field from the
		// aggregated message.
		errorMessageParts := capiErrorMessageRegexp.FindAllStringSubmatch(errStatus.Message, 3)[0]
		fieldPrefix := fmt.Sprintf("%s: ", field)

		var messageParts []string
		for _, part := range strings.Split(errorMessageParts[3], ", ") {
			if !strings.HasPrefix(part, fieldPrefix) {
				messageParts = append(messageParts, part)
			}
		}

		errStatus.Message = strings.Join(messageParts, ", ")
		errStatus.Message = fmt.Sprintf("%s[%s]", errorMessageParts[1], errStatus.Message)

		return apierrors.FromObject(&errStatus)
	}

	return err
}
