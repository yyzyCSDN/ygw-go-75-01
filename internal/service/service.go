package service

import (
	"errors"
	"log"
	"time"

	"aquarecirc/internal/aerator"
	"aquarecirc/internal/alarm"
	"aquarecirc/internal/feed"
	"aquarecirc/internal/filter"
	"aquarecirc/internal/model"
	"aquarecirc/internal/pump"
	"aquarecirc/internal/temp"
	"aquarecirc/internal/water"
)

type Service struct {
	cfg            Config
	logger         *log.Logger
	store          *water.Store
	bus            *alarm.Bus
	alarmBuf       *alarm.BufferSink
	aerators       map[string]*aerator.Aerator
	filters        map[string]*filter.Filter
	backwashes     map[string]*filter.FilterBackwash
	pumps          map[string]*pump.Pump
	feeders        map[string]*feed.Feeder
	feedSchedulers map[string]*feed.Scheduler
	temp           *temp.Controller
}

func New(cfg Config, logger *log.Logger) (*Service, error) {
	bus := alarm.NewBus()
	bus.AddSink(alarm.NewConsoleSink(logger.Writer()))
	alarmBuf := alarm.NewBufferSink(200)
	bus.AddSink(alarmBuf)
	store := water.NewStore(bus)
	ponds := cfg.Ponds
	if len(ponds) == 0 {
		ponds = defaultPonds()
	}
	aerators := map[string]*aerator.Aerator{}
	filters := map[string]*filter.Filter{}
	backwashes := map[string]*filter.FilterBackwash{}
	pumps := map[string]*pump.Pump{}
	feeders := map[string]*feed.Feeder{}
	feedSchedulers := map[string]*feed.Scheduler{}
	for _, spec := range ponds {
		pond := model.NewPond(spec.ID, spec.Name, spec.Volume, spec.Stock)
		if err := store.RegisterPond(pond); err != nil {
			return nil, err
		}
		_ = store.RegisterSensor(&water.Sensor{ID: spec.ID + "-do", PondID: spec.ID, Metric: model.MetricDO, Last: 6.2, Online: true})
		_ = store.RegisterSensor(&water.Sensor{ID: spec.ID + "-ammonia", PondID: spec.ID, Metric: model.MetricAmmonia, Last: 0.2, Online: true})
		_ = store.RegisterSensor(&water.Sensor{ID: spec.ID + "-ph", PondID: spec.ID, Metric: model.MetricPH, Last: 7.0, Online: true})
		_ = store.RegisterSensor(&water.Sensor{ID: spec.ID + "-temp", PondID: spec.ID, Metric: model.MetricTemp, Last: 26, Online: true})
		aeratorDriver := aerator.NewSimDriver()
		if cfg.SimFailAerator {
			aeratorDriver.SetStartError(errors.New("simulated aerator fault"))
		}
		aerators[spec.ID] = aerator.New("aer-"+spec.ID, spec.ID, store, bus, aeratorDriver)
		unit := filter.New("flt-"+spec.ID, spec.ID, store, filter.NewSimDriver())
		filters[spec.ID] = unit
		backwashes[spec.ID] = filter.NewBackwash(unit, 4*time.Hour)
		pumpDriver := pump.NewSimDriver()
		if cfg.SimFailPump {
			pumpDriver.SetStartError(errors.New("simulated pump fault"))
		}
		pumps[spec.ID] = pump.New("pmp-"+spec.ID, spec.ID, store, pumpDriver, bus)
		feeder := feed.NewFeeder(store, pumps[spec.ID])
		feeders[spec.ID] = feeder
		feedSchedulers[spec.ID] = feed.NewScheduler(feeder)
	}
	return &Service{
		cfg:            cfg,
		logger:         logger,
		store:          store,
		bus:            bus,
		alarmBuf:       alarmBuf,
		aerators:       aerators,
		filters:        filters,
		backwashes:     backwashes,
		pumps:          pumps,
		feeders:        feeders,
		feedSchedulers: feedSchedulers,
		temp:           temp.NewController(store),
	}, nil
}
