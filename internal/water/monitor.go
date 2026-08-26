package water

import (
	"sync"
	"time"

	"aquarecirc/internal/model"
)

type MonitorConn struct {
	id   int
	open bool
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
	conn := &MonitorConn{id: m.nextID, open: true}
	m.conns[conn.id] = conn
	m.openConns++
	return conn, nil
}

func (m *Monitor) Close(conn *MonitorConn) error {
	if conn == nil || !conn.open {
		return model.ErrClosed
	}
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
	time.Sleep(2 * time.Millisecond)
	return store.ReadSensor(sensorID)
}

func (m *Monitor) CloseIdle(maxAge time.Duration) int {
	return 0
}

