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
		{
			for _, cause := range errStatus.Details.Causes {
				if cause.Field != field {
					causes = append(causes, cause)
				}
			}

			if len(causes) < 1 {
				// No errors left, all clear.
				return nil
			}
		}

		matches := capiErrorMessageRegexp.FindAllStringSubmatch(errStatus.Message, 1)
		var messageBuilder strings.Builder
		{
			messageBuilder.WriteString(matches[0][1])
			messageBuilder.WriteString("[")

			for i, cause := range causes {
				messageBuilder.WriteString("")
				messageBuilder.WriteString(cause.Field)
				messageBuilder.WriteString(": ")
				messageBuilder.WriteString(cause.Message)

				if len(causes)-i > 1 {
					messageBuilder.WriteString(", ")
				}
			}

			messageBuilder.WriteString("]")
		}

		errStatus.Details.Causes = causes
		errStatus.Message = messageBuilder.String()

		return apierrors.FromObject(&errStatus)
	}

	return err
}
