package alarm

import (
	"sync"

	"aquarecirc/internal/model"
)

type SnapshotCache struct {
	mu     sync.Mutex
	byPond map[string]*model.WaterSnapshot
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
	// Copy the snapshot into a per-pond slot so each pond owns an
	// independent value. Sharing a single backing variable would let a
	// later pond's write overwrite an earlier pond's cached reading.
	c.byPond[pondID] = &model.WaterSnapshot{
		PondID:  snap.PondID,
		DO:      snap.DO,
		Ammonia: snap.Ammonia,
		Nitrite: snap.Nitrite,
		PH:      snap.PH,
		Temp:    snap.Temp,
		Level:   snap.Level,
	}
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
