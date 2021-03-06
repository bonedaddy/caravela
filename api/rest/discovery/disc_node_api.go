package discovery

import (
	"context"
	"github.com/strabox/caravela/api/types"
)

// Discovery API necessary to forward the REST calls
type Discovery interface {
	CreateOffer(ctx context.Context, fromNode, toNode *types.Node, offer *types.Offer)
	RefreshOffer(ctx context.Context, fromTrader *types.Node, offer *types.Offer) bool
	UpdateOffer(ctx context.Context, fromSupplier, toTrader *types.Node, offer *types.Offer)
	RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer)
	GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer
	AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node)
}
