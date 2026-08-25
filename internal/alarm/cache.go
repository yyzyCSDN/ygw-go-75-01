package alarm

import (
	"sync"

	"aquarecirc/internal/model"
)

type SnapshotCache struct {
	mu     sync.Mutex
	byPond map[string]*model.WaterSnapshot
	shared model.WaterSnapshot
}

func NewSnapshotCache() *SnapshotCache {
	return &SnapshotCache{
		byPond: map[string]*model.WaterSnapshot{},
	}
}

func (c *SnapshotCache) WriteBack(pondID string, snap *model.WaterSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if snap == nil {
		return
	}
	c.shared = *snap
	c.byPond[pondID] = &c.shared
}

func (c *SnapshotCache) Latest(pondID string) (model.WaterSnapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	snap, ok := c.byPond[pondID]
	if !ok || snap == nil {
		return model.WaterSnapshot{}, false
	}
	return *snap, true
}

func (c *SnapshotCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.byPond)
}
