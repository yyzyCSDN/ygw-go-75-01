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

func replaceMetrics(st *model.WaterState, patch MetricPatch) {
	next := model.WaterState{PondID: st.PondID}
	if patch.DO != nil {
		next.DO = *patch.DO
	}
	if patch.Ammonia != nil {
		next.Ammonia = *patch.Ammonia
	}
	if patch.Nitrite != nil {
		next.Nitrite = *patch.Nitrite
	}
	if patch.PH != nil {
		next.PH = *patch.PH
	}
	if patch.Temp != nil {
		next.Temp = *patch.Temp
	}
	*st = next
}
