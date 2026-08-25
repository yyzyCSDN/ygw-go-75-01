package water

import "aquarecirc/internal/model"

func classify(st *model.WaterState) model.Level {
	level := model.LevelNormal
	if st.DO < 3 || st.Ammonia > 1.5 || st.Nitrite > 0.8 {
		level = model.LevelCritical
	} else if st.DO < 5 || st.Ammonia > 0.7 || st.Nitrite > 0.4 {
		level = model.LevelWarning
	}
	if st.PH < 6 || st.PH > 9 || st.Temp < 15 || st.Temp > 34 {
		level = model.LevelCritical
	}
	return level
}

// applyMetrics merges the non-nil fields of patch into st in place.
//
// Each device only reports the subset of metrics it owns (the aerator writes DO,
// the filter writes Ammonia, etc.), so a patch must update only its own fields
// and leave the others untouched. A full replace here would let two concurrent
// patches clobber each other and drop whatever the slower writer had just set.
func applyMetrics(st *model.WaterState, patch MetricPatch) {
	if patch.DO != nil {
		st.DO = *patch.DO
	}
	if patch.Ammonia != nil {
		st.Ammonia = *patch.Ammonia
	}
	if patch.Nitrite != nil {
		st.Nitrite = *patch.Nitrite
	}
	if patch.PH != nil {
		st.PH = *patch.PH
	}
	if patch.Temp != nil {
		st.Temp = *patch.Temp
	}
}
