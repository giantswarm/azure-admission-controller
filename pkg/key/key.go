package key

import (
	"errors"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	ControlPlaneEndpointPort  = 443
	ClusterNetworkServiceCIDR = "172.31.0.0/16"
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

		if errStatus.Details == nil {
			return err
		}

		// Remove any errors for the given field.
		causes := errStatus.Details.Causes
		for i, cause := range causes {
			if cause.Field == field {
				causes[i] = causes[len(causes)-1]
				causes = causes[:len(causes)-1]
			}
		}

		if len(causes) < 1 {
			// No errors left, all clear.
			return nil
		}

		errStatus.Details.Causes = causes

		// Remove any errors for this field from the
		// aggregated message.
		messageParts := strings.Split(errStatus.Message, ", ")
		fieldPrefix := fmt.Sprintf("%s: ", field)

		for i, part := range messageParts {
			if strings.HasPrefix(part, fieldPrefix) {
				messageParts[i] = messageParts[len(messageParts)-1]
				messageParts = messageParts[:len(messageParts)-1]
			}
		}

		errStatus.Message = strings.Join(messageParts, ", ")

		return apierrors.FromObject(&errStatus)
	}

	return err
}
