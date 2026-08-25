package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"aquarecirc/internal/model"
)

func (s *Service) Probe() error {
	if len(s.store.Ponds()) == 0 {
		return errors.New("no ponds registered")
	}
	if len(s.aerators) != len(s.store.Ponds()) {
		return errors.New("aerator count mismatch")
	}
	if len(s.filters) != len(s.store.Ponds()) {
		return errors.New("filter count mismatch")
	}
	if s.store.MonitorOpen() != 0 {
		return errors.New("monitor connections leaked")
	}
	if s.store.SnapshotCacheSize() == 0 {
		return errors.New("snapshot cache empty")
	}
	if s.bus.GateSize() > 10000 {
		return errors.New("report gate too large")
	}
	if s.bus.SuppressedTotal() > 100000 {
		return errors.New("suppressed alarm count too large")
	}
	if s.store.LastSampleAt().IsZero() {
		return errors.New("no sample recorded")
	}
	for _, pond := range s.store.Ponds() {
		state, ok := s.store.State(pond.ID)
		if !ok {
			return fmt.Errorf("pond %s missing state", pond.ID)
		}
		if state.Level == model.LevelNormal && state.DO == 0 {
			return fmt.Errorf("pond %s has empty metrics", pond.ID)
		}
	}
	return nil
}

func (s *Service) WaitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := s.Probe(); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return s.Probe()
}

func (s *Service) HTTPProbe(addr string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + addr + "/api/health")
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health probe status %d", resp.StatusCode)
	}
	return nil
}
