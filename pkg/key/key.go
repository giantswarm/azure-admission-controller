package key

import "fmt"

const (
	ControlPlaneEndpointPort = 443
)

func GetControlPlaneEndpointHost(clusterName string, baseDomain string) string {
	return fmt.Sprintf("api.%s.%s", clusterName, baseDomain)
}
