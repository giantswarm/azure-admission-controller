// +build liveinstallation

package env

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// EnvVarBaseDomain is the process environment variable representing the
	// E2E_BASE_DOMAIN env var which is set to k8s cluster base domain.
	// For example, if we have a cluster named "mycluster" with API Server
	// domain is api.mycluster.k8s.example.com, then the base domain is
	// k8s.example.com.
	envVarBaseDomain = "E2E_BASE_DOMAIN"

	// EnvVarE2EKubeconfig is the process environment variable representing the
	// E2E_KUBECONFIG env var.
	EnvVarE2EKubeconfig = "E2E_KUBECONFIG"
)

var (
	e2eBaseDomain string
	kubeconfig    string
)

func init() {
	var err error

	kubeconfig = os.Getenv(EnvVarE2EKubeconfig)
	if kubeconfig == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			panic(fmt.Sprintf("Env var '%s' must not be empty. Alternatively env var 'HOME' must be set and kubeconfig must be located in $HOME/.kube/config", EnvVarE2EKubeconfig))
		}

		kubeconfig = path.Join(homeDir, ".kube", "config")
	}

	e2eBaseDomain = os.Getenv(envVarBaseDomain)
	if e2eBaseDomain == "" {
		var restConfig *rest.Config
		{
			restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				panic(fmt.Sprintf("env var '%s' must not be empty, or correct kubeconfig must be specified, error %#v", envVarBaseDomain, err))
			}
		}

		url, err := url.Parse(restConfig.Host)
		if err != nil {
			panic(fmt.Sprintf("env var '%s' must not be empty, or correct kubeconfig must be specified, error %#v", envVarBaseDomain, err))
		}

		// hostname has form "g8s.<installation name>.example.com"
		hostname := url.Hostname()

		// We need "k8s.<installation name>.example.com" for base domain
		parts := strings.Split(hostname, ".")
		parts[0] = "k8s"
		e2eBaseDomain = strings.Join(parts, ".")
	}
}

func BaseDomain() string {
	return e2eBaseDomain
}

func KubeConfig() string {
	return kubeconfig
}
