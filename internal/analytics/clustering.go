package analytics

import (
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"sort"
	"time"
)

type Cluster struct {
	ID                           string
	EventIDs                     []string
	Start, End                   time.Time
	PeakMagnitude, CentroidDepth float64
}

// ClusterEvents groups nearby events in time and depth for operator review.
func ClusterEvents(events []domain.SeismicEvent, maxGap time.Duration, maxDepthGap float64) []Cluster {
	if maxGap < 0 {
		maxGap = 5 * time.Minute
	}
	if maxDepthGap <= 0 {
		maxDepthGap = .5
	}
	ordered := append([]domain.SeismicEvent(nil), events...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].OccurredAt.Before(ordered[j].OccurredAt) })
	clusters := []Cluster{}
	for _, e := range ordered {
		if len(clusters) == 0 {
			clusters = append(clusters, newCluster(e))
			continue
		}
		last := &clusters[len(clusters)-1]
		if e.OccurredAt.Sub(last.End) <= maxGap && mathAbs(e.DepthKm-last.CentroidDepth) <= maxDepthGap {
			addCluster(last, e)
		} else {
			clusters = append(clusters, newCluster(e))
		}
	}
	return clusters
}
func newCluster(e domain.SeismicEvent) Cluster {
	return Cluster{ID: "cluster-" + e.ID, EventIDs: []string{e.ID}, Start: e.OccurredAt, End: e.OccurredAt, PeakMagnitude: e.Magnitude, CentroidDepth: e.DepthKm}
}
func addCluster(c *Cluster, e domain.SeismicEvent) {
	n := float64(len(c.EventIDs))
	c.EventIDs = append(c.EventIDs, e.ID)
	c.End = e.OccurredAt
	c.CentroidDepth = (c.CentroidDepth*n + e.DepthKm) / (n + 1)
	if e.Magnitude > c.PeakMagnitude {
		c.PeakMagnitude = e.Magnitude
	}
}
func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
