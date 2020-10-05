package vmcapabilities

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// CapabilitySupported is the value returned by this API from Azure when the capability is supported
	CapabilitySupported = "True"

	CapabilityAcceleratedNetworking = "AcceleratedNetworkingEnabled"

	// For internal use only.
	capabilityMemory = "MemoryGB"
	capabilityCPUs   = "vCPUs"
)

type Config struct {
	Azure  API
	Logger micrologger.Logger
}

type VMSKU struct {
	azure     API
	initMutex sync.Mutex
	logger    micrologger.Logger
	skus      map[string]cache
}

type cache map[string]compute.ResourceSku

func New(config Config) (*VMSKU, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Azure == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Azure must not be empty", config)
	}
	return &VMSKU{
		logger: config.Logger,
		azure:  config.Azure,
		skus:   make(map[string]cache),
	}, nil
}

func (v *VMSKU) CPUs(ctx context.Context, location string, vmType string) (int, error) {
	capability, err := v.getCapability(ctx, location, vmType, capabilityCPUs)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	if capability != nil {
		cpus, err := strconv.Atoi(*capability)
		if err != nil {
			return 0, microerror.Mask(invalidUpstreamResponseError)
		}

		return cpus, nil
	}

	return 0, microerror.Mask(invalidUpstreamResponseError)
}

func (v *VMSKU) HasCapability(ctx context.Context, location string, vmType string, name string) (bool, error) {
	capability, err := v.getCapability(ctx, location, vmType, name)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if capability != nil && strings.EqualFold(*capability, CapabilitySupported) {
		return true, nil
	}

	return false, nil
}

func (v *VMSKU) Memory(ctx context.Context, location string, vmType string) (int, error) {
	capability, err := v.getCapability(ctx, location, vmType, capabilityMemory)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	if capability != nil {
		mem, err := strconv.Atoi(*capability)
		if err != nil {
			return 0, microerror.Mask(invalidUpstreamResponseError)
		}

		return mem, nil
	}

	return 0, microerror.Mask(invalidUpstreamResponseError)
}

func (v *VMSKU) getCapability(ctx context.Context, location string, vmType string, name string) (*string, error) {
	if name == "" {
		return nil, microerror.Maskf(invalidRequestError, "name can't be empty")
	}
	vmsku, err := v.getSKU(ctx, location, vmType)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if vmsku.Capabilities != nil {
		for _, capability := range *vmsku.Capabilities {
			if capability.Name != nil && *capability.Name == name {
				return capability.Value, nil
			}
		}
	}

	return nil, nil
}

func (v *VMSKU) getSKU(ctx context.Context, location string, vmType string) (compute.ResourceSku, error) {
	if location == "" {
		return compute.ResourceSku{}, microerror.Maskf(invalidRequestError, "location can't be empty")
	}
	if vmType == "" {
		return compute.ResourceSku{}, microerror.Maskf(invalidRequestError, "vmType can't be empty")
	}

	if _, ok := v.skus[location]; !ok {
		err := v.initCache(ctx, location)
		if err != nil {
			return compute.ResourceSku{}, microerror.Mask(err)
		}
	}
	vmsku, found := v.skus[location][vmType]
	if !found {
		return compute.ResourceSku{}, microerror.Maskf(skuNotFoundError, vmType)
	}

	return vmsku, nil
}

func (v *VMSKU) initCache(ctx context.Context, location string) error {
	v.initMutex.Lock()
	defer v.initMutex.Unlock()
	filter := fmt.Sprintf("location eq '%s'", location)
	v.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Initializing cache for location %s with filter: %s", location, filter))
	skus, err := v.azure.List(ctx, filter)
	if err != nil {
		return microerror.Mask(err)
	}

	v.skus[location] = skus

	v.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Initialized cache. Number of SKUs in cache for location %s: '%d'", location, len(skus)))

	return nil
}
