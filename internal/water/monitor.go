package water

import (
	"sync"
	"time"

	"aquarecirc/internal/model"
)

type MonitorConn struct {
	id        int
	open      bool
	createdAt time.Time
	lastRead  time.Time
}

type Monitor struct {
	mu        sync.Mutex
	openConns int
	nextID    int
	conns     map[int]*MonitorConn
}

func NewMonitor() *Monitor {
	return &Monitor{
		conns: map[int]*MonitorConn{},
	}
}

func (m *Monitor) Open() (*MonitorConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	conn := &MonitorConn{id: m.nextID, open: true, createdAt: time.Now()}
	m.conns[conn.id] = conn
	m.openConns++
	return conn, nil
}

func (m *Monitor) Close(conn *MonitorConn) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conn == nil || !conn.open {
		return model.ErrClosed
	}
	conn.open = false
	delete(m.conns, conn.id)
	m.openConns--
	return nil
}

func (m *Monitor) OpenCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.openConns
}

func (m *Monitor) Read(conn *MonitorConn, store *Store, sensorID string) (*model.SensorReading, error) {
	if conn == nil || !conn.open {
		return nil, model.ErrClosed
	}
	conn.lastRead = time.Now()
	time.Sleep(2 * time.Millisecond)
	return store.ReadSensor(sensorID)
}

func (m *Monitor) CloseIdle(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().Add(-maxAge)
	closed := 0
	for _, conn := range m.conns {
		if conn.lastRead.Before(cutoff) {
			conn.open = false
			delete(m.conns, conn.id)
			m.openConns--
			closed++
		}
	}
	return closed
}
