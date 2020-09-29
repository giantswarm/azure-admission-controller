package vmcapabilities

type Capabilities struct {
	SupportsAcceleratedNetworking bool
}

var (
	instanceTypes = map[string]Capabilities{
		"Standard_D4_v3":   {SupportsAcceleratedNetworking: true},
		"Standard_D4s_v3":  {SupportsAcceleratedNetworking: true},
		"Standard_D8_v3":   {SupportsAcceleratedNetworking: true},
		"Standard_D8s_v3":  {SupportsAcceleratedNetworking: true},
		"Standard_D16_v3":  {SupportsAcceleratedNetworking: true},
		"Standard_D16s_v3": {SupportsAcceleratedNetworking: true},
		"Standard_D32_v3":  {SupportsAcceleratedNetworking: true},
		"Standard_D32s_v3": {SupportsAcceleratedNetworking: true},
		"Standard_E4s_v3":  {SupportsAcceleratedNetworking: true},
		"Standard_E8a_v4":  {SupportsAcceleratedNetworking: true},
		"Standard_E8as_v4": {SupportsAcceleratedNetworking: true},
		"Standard_E8s_v3":  {SupportsAcceleratedNetworking: true},
		"Standard_E16s_v3": {SupportsAcceleratedNetworking: true},
		"Standard_E32s_v3": {SupportsAcceleratedNetworking: true},
		"Standard_F4s_v2":  {SupportsAcceleratedNetworking: true},
		"Standard_F8s_v2":  {SupportsAcceleratedNetworking: true},
		"Standard_F16s_v2": {SupportsAcceleratedNetworking: true},
		"Standard_F32s_v2": {SupportsAcceleratedNetworking: true},
	}
)

func FromInstanceType(instanceType string) (*Capabilities, error) {
	capabilities, found := instanceTypes[instanceType]
	if !found {
		return nil, unknownVMTypeError
	}

	return &capabilities, nil
}
