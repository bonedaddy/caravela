package partitions

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
	"sync"
)

var GlobalState = NewSystemResourcePartitions(13)

type SystemResourcePartitions struct {
	partitionsState sync.Map
	totalStats      int
}

func NewSystemResourcePartitions(totalStats int) *SystemResourcePartitions {
	return &SystemResourcePartitions{
		partitionsState: sync.Map{},
		totalStats:      totalStats,
	}
}

func (s *SystemResourcePartitions) Try(targetResPartition resources.Resources) bool {
	if partition, exist := s.partitionsState.Load(targetResPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			return partitionState.Try()
		}
	} else {
		newPartitionState := NewResourcePartitionState(s.totalStats)
		s.partitionsState.Store(targetResPartition, newPartitionState)
		return newPartitionState.Try()
	}
	return true
}

func (s *SystemResourcePartitions) Hit(resPartition resources.Resources) {
	if partition, exist := s.partitionsState.Load(resPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			partitionState.Hit()
		}
	}
}

func (s *SystemResourcePartitions) Miss(resPartition resources.Resources) {
	if partition, exist := s.partitionsState.Load(resPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			partitionState.Miss()
		}
	}
}

func (s *SystemResourcePartitions) PartitionsState() []types.PartitionState {
	res := make([]types.PartitionState, 0)
	s.partitionsState.Range(func(key, value interface{}) bool {
		partResources, _ := key.(resources.Resources)
		if partitionState, ok := value.(*ResourcePartitionState); ok {
			res = append(res, types.PartitionState{
				PartitionResources: types.Resources{
					CPUs: partResources.CPUs(),
					RAM:  partResources.RAM(),
				},
				Hits: partitionState.hits,
			})
		}
		return true
	})
	return res
}

func (s *SystemResourcePartitions) MergePartitionsState(newPartitionsState []types.PartitionState) {
	for _, newPartitionState := range newPartitionsState {
		partRes := resources.NewResources(newPartitionState.PartitionResources.CPUs, newPartitionState.PartitionResources.RAM)
		if partition, exist := s.partitionsState.Load(*partRes); exist {
			if partitionState, ok := partition.(*ResourcePartitionState); ok {
				partitionState.Merge(newPartitionState.Hits)
			}
		} else {
			unknownPartitionState := NewResourcePartitionState(s.totalStats)
			unknownPartitionState.hits = newPartitionState.Hits
			s.partitionsState.Store(*partRes, unknownPartitionState)
		}
	}
}