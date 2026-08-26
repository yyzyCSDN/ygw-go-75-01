package service

import (
	"os"
	"strconv"
	"time"
)

type PondSpec struct {
	ID     string
	Name   string
	Volume float64
	Stock  int
}

type Config struct {
	Addr           string
	Ponds          []PondSpec
	SensorEvery    time.Duration
	LoopEvery      time.Duration
	SimFailAerator bool
	SimFailPump    bool
}

func LoadConfig() Config {
	port := envInt("AQUA_PORT", 8080)
	return Config{
		Addr:           ":" + strconv.Itoa(port),
		Ponds:          defaultPonds(),
		SensorEvery:    2 * time.Second,
		LoopEvery:      2 * time.Second,
		SimFailAerator: envBool("AQUA_SIM_AERATOR_FAIL", false),
		SimFailPump:    envBool("AQUA_SIM_PUMP_FAIL", false),
	}
}

func defaultPonds() []PondSpec {
	return []PondSpec{
		{ID: "p1", Name: "A区主池", Volume: 120, Stock: 1200},
		{ID: "p2", Name: "B区辅池", Volume: 80, Stock: 600},
	}
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}
