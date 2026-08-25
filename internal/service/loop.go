package service

import (
	"context"
	"sync"
	"time"

	"aquarecirc/internal/aerator"
	"aquarecirc/internal/alarm"
	"aquarecirc/internal/filter"
	"aquarecirc/internal/model"
)

func (s *Service) RunLoops(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(4 + len(s.aerators) + len(s.filters))
	go func() {
		defer wg.Done()
		s.sensorLoop(ctx)
	}()
	go func() {
		defer wg.Done()
		s.tempLoop(ctx)
	}()
	go func() {
		defer wg.Done()
		s.sampleLoop(ctx)
	}()
	go func() {
		defer wg.Done()
		s.equipmentLoop(ctx)
	}()
	for _, unit := range s.aerators {
		go func(a *aerator.Aerator) {
			defer wg.Done()
			a.RunAuto(ctx, s.cfg.LoopEvery)
		}(unit)
	}
	for _, unit := range s.filters {
		go func(f *filter.Filter) {
			defer wg.Done()
			f.RunAuto(ctx, s.cfg.LoopEvery)
		}(unit)
	}
	wg.Wait()
}

func (s *Service) sensorLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.SensorEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, sensor := range s.store.Sensors() {
				if err := s.store.PollSensor(sensor.ID); err != nil {
					s.logger.Printf("poll sensor %s: %v", sensor.ID, err)
				}
			}
		}
	}
}

func (s *Service) tempLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.LoopEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, pond := range s.store.Ponds() {
				regCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
				_ = s.temp.Regulate(regCtx, pond.ID, pond.TargetTemp)
				cancel()
			}
		}
	}
}

func (s *Service) equipmentLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.LoopEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, unit := range s.pumps {
				status := unit.Status()
				rpm := status.ActualRPM
				if status.State == model.PumpRunning {
					rpm += 15
				}
				unit.RecordRotation(rpm)
			}
			for pondID, backwash := range s.backwashes {
				backwash.Track(s.cfg.LoopEvery)
				if backwash.Due() {
					if err := backwash.Run(ctx); err != nil {
						s.logger.Printf("backwash %s: %v", pondID, err)
					}
				}
			}
		}
	}
}

func (s *Service) sampleLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.LoopEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.store.SampleAll()
			s.evaluateAlarms()
			s.store.CloseIdleConns(5 * time.Minute)
			s.bus.Prune(2 * time.Hour)
			s.bus.GateSize()
		}
	}
}

func (s *Service) evaluateAlarms() {
	for _, pond := range s.store.Ponds() {
		snap, ok := s.store.LatestSnapshot(pond.ID)
		if !ok {
			continue
		}
		if snap.Level != model.LevelCritical {
			continue
		}
		key := alarm.NewReportKey(pond.ID, "snapshot", snap.Level.String())
		if s.bus.Acquire(key) {
			_ = s.bus.Report(model.NewAlarm(pond.ID, model.MetricDevice, snap.Level, "snapshot critical"))
		}
	}
}
