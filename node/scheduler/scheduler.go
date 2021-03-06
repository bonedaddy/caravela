package scheduler

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/util"
	"unsafe"
)

// Scheduler is responsible for receiving local and remote requests for deploying containers
// to run in the system. It takes a request for running a container and decides where to deploy it
// in conjunction with the Discovery component.
type Scheduler struct {
	common.NodeComponent // Base component.

	config *configuration.Configuration // System's configuration.
	client userRemoteClient             // CARAVELA's remote client.

	discovery         discoveryLocal        // Local Discovery component.
	containersManager containerManagerLocal // Local Containers Manager component.
}

// NewScheduler creates a new local scheduler component.
func NewScheduler(config *configuration.Configuration, internalDisc discoveryLocal,
	containersManager containerManagerLocal, client userRemoteClient) *Scheduler {

	return &Scheduler{
		config:            config,
		client:            client,
		discovery:         internalDisc,
		containersManager: containersManager,
	}
}

// Launch is executed when a system's node wants to launch a container in this node.
func (s *Scheduler) Launch(ctx context.Context, fromBuyer *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {

	if !s.IsWorking() {
		panic(fmt.Errorf("can't launch container, scheduler not working"))
	}

	if len(containersConfigs) == 0 {
		return make([]types.ContainerStatus, 0), errors.New("no container configurations")
	}

	totalResourcesNecessary := resources.NewResources(0, 0)
	for i, contConfig := range containersConfigs {
		log.Debugf(util.LogTag("SCHEDULE")+"Launching... [%d] Img: %s, Res: <%d,%d>", i, contConfig.ImageKey,
			contConfig.Resources.CPUs, contConfig.Resources.Memory)
		totalResourcesNecessary.Add(*resources.NewResources(contConfig.Resources.CPUs, contConfig.Resources.Memory))
	}

	containerStatus, err := s.containersManager.StartContainer(fromBuyer, offer, containersConfigs, *totalResourcesNecessary)
	return containerStatus, err
}

// SubmitContainers is called when the user submits a request to the node in order to deploy a set of containers.
func (s *Scheduler) SubmitContainers(ctx context.Context, contConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	if !s.IsWorking() {
		panic(fmt.Errorf("can't run container, scheduler not working"))
	}

	resContainersStatus := make([]types.ContainerStatus, 0)

	// ================== Check for the containers group policies ==================

	coLocateTotalResources := resources.NewResourcesCPUClass(int(types.LowCPUPClass), 0, 0)
	coLocateContainers := make([]types.ContainerConfig, 0)
	spreadContainers := make([]types.ContainerConfig, 0)

	for i, contConfig := range contConfigs {
		log.Debugf(util.LogTag("SCHEDULE")+"Deploying [#%d]... Img: %s , Res: <%d;%d>, GrpPolicy: %s", i, contConfig.ImageKey,
			contConfig.Resources.CPUs, contConfig.Resources.Memory, contConfig.GroupPolicy)

		if contConfig.GroupPolicy == types.CoLocationGroupPolicy {
			coLocateContainers = append(coLocateContainers, contConfig)
			coLocateTotalResources.Add(*resources.NewResources(contConfig.Resources.CPUs, contConfig.Resources.Memory))
			coLocateTotalResources.SetCPUClass(int(contConfig.Resources.CPUClass))
		} else if contConfig.GroupPolicy == types.SpreadGroupPolicy {
			spreadContainers = append(spreadContainers, contConfig)
		}
	}

	// ================== First try launch the co-located containers ==============

	containersStatus, err := s.launchContainers(ctx, coLocateContainers, *coLocateTotalResources)
	if err != nil {
		return nil, err
	}
	resContainersStatus = append(resContainersStatus, containersStatus...)

	// ================ Then launch the containers that can be spread ==============

	for _, contConfig := range spreadContainers {
		resourcesNecessary := resources.NewResourcesCPUClass(int(contConfig.Resources.CPUClass), contConfig.Resources.CPUs, contConfig.Resources.Memory)

		containersStatus, err := s.launchContainers(ctx, []types.ContainerConfig{contConfig}, *resourcesNecessary)
		if err != nil {
			for i := range resContainersStatus { // Stop all the previous launched containers
				s.client.StopLocalContainer(ctx, &types.Node{IP: resContainersStatus[i].SupplierIP}, resContainersStatus[i].ContainerID)
			}
			return nil, err
		}

		resContainersStatus = append(resContainersStatus, containersStatus...)
	}

	return resContainersStatus, nil
}

// launchContainer launches a container in a node with the resources necessary available.
func (s *Scheduler) launchContainers(ctx context.Context, containersConfigs []types.ContainerConfig,
	resourcesNecessary resources.Resources) ([]types.ContainerStatus, error) {

	resContainersStatus := make([]types.ContainerStatus, 0)

	if len(containersConfigs) == 0 {
		return resContainersStatus, nil
	}

	offers := s.discovery.FindOffers(ctx, resourcesNecessary)
	offers = CreateSchedulePolicy(s.config).Rank(offers, resourcesNecessary) // Rank the offers according with the scheduling policy.

	if len(offers) == 0 {
		log.Debugf(util.LogTag("SCHEDULE") + "Deploy FAILED. No offers found.")
		return resContainersStatus, errors.New("no offers found to deploy")
	}

	for offerIndex, offer := range offers {
		log.Debugf(util.LogTag("SCHEDULE")+"Trying OFFER [#%d]... SuppIP: %s, Offer: %d, Amount %d, Res: <%d;%d>",
			offerIndex, offer.SupplierIP, offer.ID, offer.Amount, offer.FreeResources.CPUs, offer.FreeResources.Memory)

		containersStatus, err := s.client.LaunchContainer(
			ctx,
			&types.Node{IP: s.config.HostIP()},
			&types.Node{IP: offer.SupplierIP},
			&types.Offer{ID: offer.ID},
			containersConfigs)
		if err != nil {
			log.Debugf(util.LogTag("SCHEDULE")+"Deploy FAILED [#%d] Offer: %d error: %s", offerIndex, offer.ID, err)
			if offerIndex == (len(offers) - 1) {
				log.Debugf(util.LogTag("SCHEDULE") + "Deploy FAILED. No offers found.")
				return resContainersStatus, errors.New("all offers were reject to deploy")
			}
			continue
		}

		resContainersStatus = append(resContainersStatus, containersStatus...)
		log.Debugf(util.LogTag("SCHEDULE") + "Deploy SUCCESS")
		break
	}

	return resContainersStatus, nil
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (s *Scheduler) Start() {
	s.Started(s.config.Simulation(), func() { /* Do Nothing */ })
}

func (s *Scheduler) Stop() {
	s.Stopped(func() { /* Do Nothing */ })
}

func (s *Scheduler) IsWorking() bool {
	return s.Working()
}

// ===============================================================================
// =							    Debug Methods                                =
// ===============================================================================

func (s *Scheduler) DebugSizeBytes() int {
	schedulerSizeBytes := unsafe.Sizeof(*s)
	return int(schedulerSizeBytes)
}
