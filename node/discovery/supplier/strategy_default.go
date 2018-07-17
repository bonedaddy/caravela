package supplier

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
)

type DefaultChordOffersManager struct {
	configs          *configuration.Configuration
	resourcesMapping *resources.Mapping
	overlay          overlay.Overlay
	remoteClient     remote.Caravela
}

func (man *DefaultChordOffersManager) Init(resourcesMap *resources.Mapping, overlay overlay.Overlay,
	remoteClient remote.Caravela) {

	man.resourcesMapping = resourcesMap
	man.overlay = overlay
	man.remoteClient = remoteClient
}

func (man *DefaultChordOffersManager) FindOffers(targetResources resources.Resources) []api.Offer {
	var destinationGUID *guid.GUID = nil
	findPhase := 0
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, _ = man.resourcesMapping.RandGUID(targetResources)
		} else { // Random trader in higher resources zone
			destinationGUID, err = man.resourcesMapping.HigherRandGUID(*destinationGUID, targetResources)
			if err != nil {
				return make([]api.Offer, 0)
			} // No more resource partitions to search
		}

		res, _ := man.resourcesMapping.ResourcesByGUID(*destinationGUID)
		log.Debugf("DestinationGUIDRes: %s", res.String())

		overlayNodes, _ := man.overlay.Lookup(destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)

		for _, node := range overlayNodes {
			offers, err := man.remoteClient.GetOffers(node.IP(), guid.NewGUIDBytes(node.GUID()).String(), true, "")
			if (err == nil) && (len(offers) != 0) {
				return offers
			}
		}
		findPhase++
	}
}

func (man *DefaultChordOffersManager) AdvertiseOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error) {
	var err error
	var overlayNodes []*overlay.Node = nil
	destinationGUID, _ := man.resourcesMapping.RandGUID(availableResources)
	overlayNodes, _ = man.overlay.Lookup(destinationGUID.Bytes())
	overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)

	// .. try search nodes in the beginning of the original target resource range region
	if len(overlayNodes) == 0 {
		destinationGUID := man.resourcesMapping.FirstGUID(availableResources)
		overlayNodes, _ = man.overlay.Lookup(destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// ... try search for random nodes that handle less powerful resource combinations
	for len(overlayNodes) == 0 {
		destinationGUID, err = man.resourcesMapping.LowerRandGUID(*destinationGUID, availableResources)
		if err != nil {
			log.Errorf(util.LogTag("Supplier")+"NO NODES to handle resources offer: %s, error: %s",
				availableResources.String(), err)
			return nil, errors.New("no nodes available to accept offer") // Wait fot the next tick to try supply resources
		}
		overlayNodes, _ = man.overlay.Lookup(destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// Chose the first node returned by the overlay API
	chosenNode := overlayNodes[0]
	chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

	err = man.remoteClient.CreateOffer(man.configs.HostIP(), "", chosenNode.IP(),
		chosenNodeGUID.String(), newOfferID, 1, availableResources.CPUs(),
		availableResources.RAM())

	if err == nil {
		return newSupplierOffer(common.OfferID(newOfferID), 1, availableResources, chosenNode.IP(), *chosenNodeGUID), nil
	}

	return nil, errors.New("impossible advertise offer")
}

// Remove nodes that do not belong to that target GUID partition. (Probably because we were target a frontier node)
func (man *DefaultChordOffersManager) removeNonTargetNodes(remoteNodes []*overlay.Node, targetGuid guid.GUID) []*overlay.Node {
	resultNodes := make([]*overlay.Node, 0)
	targetGuidResources, _ := man.resourcesMapping.ResourcesByGUID(targetGuid)
	for _, remoteNode := range remoteNodes {
		remoteNodeResources, _ := man.resourcesMapping.ResourcesByGUID(*guid.NewGUIDBytes(remoteNode.GUID()))
		if targetGuidResources.Equals(*remoteNodeResources) {
			resultNodes = append(resultNodes, remoteNode)
		}
	}
	return resultNodes
}