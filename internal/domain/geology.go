package domain

import (
	"fmt"
	"strings"
)

type RockLayer struct {
	DepthTop, DepthBottom   float64
	Formation, Permeability string
	FractureRisk            float64
}

func (l RockLayer) Validate() error {
	if l.DepthTop < 0 || l.DepthBottom <= l.DepthTop {
		return FieldError{"depth", "layer bounds invalid"}
	}
	if strings.TrimSpace(l.Formation) == "" {
		return FieldError{"formation", "is required"}
	}
	if l.FractureRisk < 0 || l.FractureRisk > 1 {
		return FieldError{"fracture_risk", "must be between zero and one"}
	}
	return nil
}
func (l RockLayer) Contains(depth float64) bool { return depth >= l.DepthTop && depth < l.DepthBottom }
func (l RockLayer) Thickness() float64          { return l.DepthBottom - l.DepthTop }
func ResolveLayer(layers []RockLayer, depth float64) (RockLayer, error) {
	for _, layer := range layers {
		if layer.Contains(depth) {
			return layer, nil
		}
	}
	return RockLayer{}, ConflictError{"rock_layer", fmt.Sprintf("no layer at %.2fkm", depth)}
}
func SortLayers(layers []RockLayer) []RockLayer {
	out := append([]RockLayer(nil), layers...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].DepthTop < out[i].DepthTop {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
func LayersCover(layers []RockLayer, depth float64) bool {
	sorted := SortLayers(layers)
	cursor := 0.0
	for _, layer := range sorted {
		if layer.DepthTop > cursor {
			return false
		}
		if layer.DepthBottom > cursor {
			cursor = layer.DepthBottom
		}
		if cursor >= depth {
			return true
		}
	}
	return cursor >= depth
}
func RiskFromLayers(layers []RockLayer) float64 {
	if len(layers) == 0 {
		return 0
	}
	risk := 0.0
	for _, layer := range layers {
		if layer.FractureRisk > risk {
			risk = layer.FractureRisk
		}
	}
	return risk
}
