package node

import (
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/scheduler"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/overlay/chord"
)

/*
Top level structure that contains all the modules/objects that manages the Caravela node.
*/
type Node struct {
	discovery         *discovery.Discovery
	scheduler         *scheduler.Scheduler
	containersManager *containers.Manager
	config            *configuration.Configuration
	overlay           overlay.Overlay
	dockerClient      docker.Client
}

func NewNode(config *configuration.Configuration) *Node {
	res := &Node{}

	res.config = config

	// Global GUID size initialization
	guid.InitializeGUID(config.ChordHashSizeBits())

	// Create Overlay struct (Chord overlay initial)
	res.overlay = chord.NewChordOverlay(guid.SizeBytes(), config.HostIP(), config.OverlayPort(),
		config.ChordVirtualNodes(), config.ChordNumSuccessors(), config.ChordTimeout())

	// Create CARAVELA's remote client
	caravelaCli := remote.NewHttpClient(config)

	// Create resources mapping based on the configurations
	resourcesMap := resources.NewResourcesMap(resources.GetCpuCoresPartitions(config.CpuCoresPartitions()),
		resources.GetRamPartitions(config.RamPartitions()))
	resourcesMap.Print()

	// Create Docker client and obtain the maximum resources Docker Engine has available
	res.dockerClient = docker.NewDockerClient()
	res.dockerClient.Initialize(config.DockerAPIVersion())
	maxCPUs, maxRAM := res.dockerClient.GetDockerCPUAndRAM()
	maxResources := resources.NewResources(maxCPUs, maxRAM)

	res.discovery = discovery.NewDiscovery(config, res.overlay, caravelaCli, resourcesMap, *maxResources)

	// Create the containers manager
	res.containersManager = containers.NewManager(config, res.dockerClient, res.discovery)

	res.scheduler = scheduler.NewScheduler(config, res.discovery, res.containersManager, caravelaCli)

	return res
}

func (node *Node) Start(join bool, joinIP string) {
	if join {
		node.overlay.Join(joinIP, node.config.OverlayPort(), node)
	} else {
		node.overlay.Create(node)
	}

	node.discovery.Start()
	node.containersManager.Start()
}

func (node *Node) AddTrader(guidBytes []byte) {
	guidRes := guid.NewGUIDBytes(guidBytes)
	node.discovery.AddTrader(*guidRes)
}

/* ================================== NodeRemote ============================= */

func (node *Node) Discovery() nodeAPI.Discovery {
	return node.discovery
}

func (node *Node) Scheduler() nodeAPI.Scheduler {
	return node.scheduler
}
