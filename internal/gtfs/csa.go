package gtfs

import (
	"fmt"
	"math"
)

const minTransferSec = 60

func (d *Dataset) ComputeBackward(targetStopIdxs []int, arriveBySec int) []Result {
	n := len(d.stops)
	latestDep := make([]int, n)
	for i := range latestDep {
		latestDep[i] = math.MinInt32
	}

	for _, t := range targetStopIdxs {
		if t < 0 || t >= n {
			continue
		}
		if arriveBySec > latestDep[t] {
			latestDep[t] = arriveBySec
		}
	}

	updated := true
	for updated {
		updated = false
		for _, t := range targetStopIdxs {
			if propagateWalkBackward(t, latestDep, d.transfers) {
				updated = true
			}
		}
	}

	for _, c := range d.connections {
		if c.arrSec > latestDep[c.arrStop] {
			continue
		}
		bestDep := c.depSec
		if bestDep > latestDep[c.depStop] {
			latestDep[c.depStop] = bestDep
			propagateWalkBackward(c.depStop, latestDep, d.transfers)
		}
	}

	results := make([]Result, n)
	for i, s := range d.stops {
		results[i] = Result{Stop: s}
		if latestDep[i] > math.MinInt32 && latestDep[i] <= arriveBySec {
			secs := arriveBySec - latestDep[i]
			results[i].CommuteSecondsToTarget = &secs
		}
	}
	return results
}

func propagateWalkBackward(stopIdx int, latestDep []int, transfers [][]transfer) bool {
	any := false
	queue := []int{stopIdx}
	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]
		for _, tr := range transfers[s] {
			candidate := latestDep[s] - tr.seconds
			if candidate > latestDep[tr.toStop] {
				latestDep[tr.toStop] = candidate
				queue = append(queue, tr.toStop)
				any = true
			}
		}
	}
	return any
}

func (d *Dataset) ResolveTarget(refLat, refLon float64, nameQuery string) ([]int, error) {
	if nameQuery != "" {
		matches := d.FindStopsByName(nameQuery)
		if len(matches) > 0 {
			return matches, nil
		}
	}
	idx, err := d.FindNearestStop(refLat, refLon)
	if err != nil {
		return nil, err
	}
	if idx < 0 {
		return nil, fmt.Errorf("no stop near %.4f,%.4f", refLat, refLon)
	}
	return []int{idx}, nil
}
