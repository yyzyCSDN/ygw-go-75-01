package service

import (
	"encoding/json"
	"net/http"
	"os"

	"aquarecirc/internal/filter"
	"aquarecirc/internal/model"
)

type pondView struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Volume  float64          `json:"volume"`
	Stock   int              `json:"stock"`
	State   model.WaterState `json:"state"`
	Sensors []sensorView     `json:"sensors"`
	Balance float64          `json:"feed_balance"`
}

type sensorView struct {
	ID     string `json:"id"`
	Metric string `json:"metric"`
	Status string `json:"status"`
}

type deviceView struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	PondID  string `json:"pond_id"`
	Running bool   `json:"running"`
	Error   string `json:"error,omitempty"`
}

type aeratorView struct {
	ID         string `json:"id"`
	PondID     string `json:"pond_id"`
	StartCount int    `json:"start_count"`
	StopCount  int    `json:"stop_count"`
	Running    bool   `json:"running"`
}

type pumpHealthView struct {
	PondID      string  `json:"pond_id"`
	SampleCount int     `json:"sample_count"`
	AvgRPM      float64 `json:"avg_rpm"`
	Deviation   float64 `json:"deviation"`
	Degraded    bool    `json:"degraded"`
}

func (s *Service) HTTPServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleConsole)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/ponds", s.handlePonds)
	mux.HandleFunc("/api/readings", s.handleReadings)
	mux.HandleFunc("/api/alarms", s.handleAlarms)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/feed", s.handleFeed)
	mux.HandleFunc("/api/feed/history", s.handleFeedHistory)
	mux.HandleFunc("/api/adjust", s.handleAdjust)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/pump/restart", s.handlePumpRestart)
	mux.HandleFunc("/api/backwash", s.handleBackwash)
	mux.HandleFunc("/api/history", s.handleHistory)
	return &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5e9,
	}
}

func (s *Service) handleConsole(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := os.ReadFile("web/console.html")
	if err != nil {
		http.Error(w, "console page unavailable", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"ponds":    len(s.store.Ponds()),
		"monitors": s.store.MonitorOpen(),
		"alarms":   s.alarmBuf.Len(),
	})
}

func (s *Service) handlePonds(w http.ResponseWriter, r *http.Request) {
	views := make([]pondView, 0, len(s.store.Ponds()))
	for _, pond := range s.store.Ponds() {
		state, _ := s.store.State(pond.ID)
		view := pondView{
			ID:      pond.ID,
			Name:    pond.Name,
			Volume:  pond.Volume,
			Stock:   pond.Stock,
			State:   state,
			Balance: s.store.FeedBalance(pond.ID),
		}
		for _, sensor := range s.store.Sensors() {
			if sensor.PondID != pond.ID {
				continue
			}
			reading, err := s.store.ReadSensor(sensor.ID)
			status := "unknown"
			if err == nil {
				status = reading.Summary()
			}
			view.Sensors = append(view.Sensors, sensorView{ID: sensor.ID, Metric: string(sensor.Metric), Status: status})
		}
		views = append(views, view)
	}
	writeJSON(w, http.StatusOK, views)
}

func (s *Service) handleReadings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var reading model.SensorReading
	if err := json.NewDecoder(r.Body).Decode(&reading); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.store.ApplySensor(&reading); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accepted": true})
}

func (s *Service) handleAlarms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.bus.Recent(100))
}

func (s *Service) handleEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Events(50))
}

func (s *Service) handleFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PondID string `json:"pond_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	scheduler, ok := s.feedSchedulers[req.PondID]
	if !ok {
		http.Error(w, "pond not found", http.StatusNotFound)
		return
	}
	batch, err := scheduler.Start(r.Context(), req.PondID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (s *Service) handleAdjust(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PondID  string   `json:"pond_id"`
		DO      *float64 `json:"do"`
		Ammonia *float64 `json:"ammonia"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.DO != nil {
		unit, ok := s.aerators[req.PondID]
		if !ok {
			http.Error(w, "pond not found", http.StatusNotFound)
			return
		}
		if err := unit.AdjustDO(*req.DO); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if req.Ammonia != nil {
		unit, ok := s.filters[req.PondID]
		if !ok {
			http.Error(w, "pond not found", http.StatusNotFound)
			return
		}
		if err := unit.AdjustAmmonia(*req.Ammonia); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"adjusted": true})
}

func (s *Service) handleStatus(w http.ResponseWriter, r *http.Request) {
	devices := make([]deviceView, 0, len(s.aerators)+len(s.filters)+len(s.pumps))
	aerators := make([]aeratorView, 0, len(s.aerators))
	for _, unit := range s.aerators {
		st := unit.Status()
		devices = append(devices, deviceView{ID: st.ID, Kind: string(st.Kind), PondID: st.PondID, Running: st.Running, Error: st.LastError})
		stats := unit.Stats()
		aerators = append(aerators, aeratorView{ID: st.ID, PondID: st.PondID, StartCount: stats.StartCount, StopCount: stats.StopCount, Running: st.Running})
	}
	for _, unit := range s.filters {
		st := unit.Status()
		devices = append(devices, deviceView{ID: st.ID, Kind: string(st.Kind), PondID: st.PondID, Running: st.Running, Error: st.LastError})
	}
	for _, unit := range s.pumps {
		st := unit.DeviceStatus()
		devices = append(devices, deviceView{ID: st.ID, Kind: string(st.Kind), PondID: st.PondID, Running: st.Running, Error: st.LastError})
	}
	health := make([]pumpHealthView, 0, len(s.pumps))
	for pondID, unit := range s.pumps {
		h := unit.Health()
		health = append(health, pumpHealthView{PondID: pondID, SampleCount: h.SampleCount, AvgRPM: h.AvgRPM, Deviation: h.Deviation, Degraded: h.Degraded})
	}
	backwash := map[string]filter.BackwashStatus{}
	for pondID, unit := range s.backwashes {
		backwash[pondID] = unit.Status()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"devices":     devices,
		"temp":        s.tempStatuses(),
		"aerators":    aerators,
		"pump_health": health,
		"backwash":    backwash,
	})
}

func (s *Service) handleHistory(w http.ResponseWriter, r *http.Request) {
	pondID := r.URL.Query().Get("pond_id")
	if pondID == "" {
		http.Error(w, "pond_id required", http.StatusBadRequest)
		return
	}
	metric, ok := model.ParseMetric(r.URL.Query().Get("metric"))
	if !ok {
		http.Error(w, "metric required", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, s.store.History(pondID, metric, 100))
}

func (s *Service) handleFeedHistory(w http.ResponseWriter, r *http.Request) {
	pondID := r.URL.Query().Get("pond_id")
	if pondID == "" {
		http.Error(w, "pond_id required", http.StatusBadRequest)
		return
	}
	feeder, ok := s.feeders[pondID]
	if !ok {
		http.Error(w, "pond not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, feeder.Ledger().Recent(pondID, 50))
}

func (s *Service) handleBackwash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PondID string `json:"pond_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	unit, ok := s.backwashes[req.PondID]
	if !ok {
		http.Error(w, "pond not found", http.StatusNotFound)
		return
	}
	if err := unit.Run(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, unit.Status())
}

func (s *Service) tempStatuses() map[string]model.TempRegulation {
	out := map[string]model.TempRegulation{}
	for _, pond := range s.store.Ponds() {
		if reg, ok := s.temp.Status(pond.ID); ok {
			out[pond.ID] = reg
		}
	}
	return out
}

func (s *Service) handlePumpRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PondID string `json:"pond_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	unit, ok := s.pumps[req.PondID]
	if !ok {
		http.Error(w, "pond not found", http.StatusNotFound)
		return
	}
	if err := unit.Restart(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, unit.Status())
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
