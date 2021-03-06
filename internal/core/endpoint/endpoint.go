package endpoint

import (
	"errors"
	"fmt"
	"strings"

	"github.com/leon-yc/ggs/internal/core/registry"
	"github.com/leon-yc/ggs/internal/pkg/runtime"
	utiltags "github.com/leon-yc/ggs/internal/pkg/util/tags"
	"github.com/leon-yc/ggs/pkg/qlog"
)

//GetEndpointFromServiceCenter is used to get the endpoint based on appID, service name and version
//it will only return endpoints of a service
func GetEndpointFromServiceCenter(appID, microService, version string) (string, error) {
	var endPoint string

	tags := utiltags.NewDefaultTag(version, appID)
	instances, err := registry.DefaultServiceDiscoveryService.FindMicroServiceInstances(runtime.ServiceID, microService, tags)
	if err != nil {
		qlog.Warnf("Get service instance failed, for key: %s:%s:%s",
			appID, microService, version)
		return "", err
	}

	if len(instances) == 0 {
		instanceError := fmt.Sprintf("No available instance, key: %s:%s:%s",
			appID, microService, version)
		return "", errors.New(instanceError)
	}

	for _, instance := range instances {
		for _, value := range instance.EndpointsMap {
			if strings.Contains(value, "?") {
				separation := strings.Split(value, "?")
				if separation[1] == "sslEnabled=true" {
					endPoint = "https://" + separation[0]
				} else {
					endPoint = "http://" + separation[0]
				}
			} else {
				endPoint = "https://" + value
			}
		}
	}

	return endPoint, nil
}
