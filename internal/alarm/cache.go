package alarm

import (
	"sync"

	"aquarecirc/internal/model"
)

type SnapshotCache struct {
	mu     sync.Mutex
	byPond map[string]model.WaterSnapshot
}

func NewSnapshotCache() *SnapshotCache {
	return &SnapshotCache{
		byPond: map[string]model.WaterSnapshot{},
	}
}

func (c *SnapshotCache) WriteBack(pondID string, snap *model.WaterSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if snap == nil {
		return
	}
	c.byPond[pondID] = *snap
}

func (c *SnapshotCache) Latest(pondID string) (model.WaterSnapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	snap, ok := c.byPond[pondID]
	if !ok {
		return model.WaterSnapshot{}, false
	}
	return snap, true
}

func (c *SnapshotCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.byPond)
}
