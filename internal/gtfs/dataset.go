package gtfs

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Dataset struct {
	stops       []Stop
	stopKeyToIdx map[string]int
	connections []connection
	transfers   [][]transfer
}

func Load(sources []FeedSource, walkSpeedMps, walkDetour float64, day time.Weekday, hour int) (*Dataset, error) {
	feeds := make([]*feed, 0, len(sources))
	for _, src := range sources {
		f, err := parseFeed(src)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}

	targetDate := nextWeekday(day)

	ds := &Dataset{
		stopKeyToIdx: make(map[string]int),
	}

	feedStopOffset := make([]int, len(feeds))
	for fi, f := range feeds {
		feedStopOffset[fi] = len(ds.stops)
		for _, s := range f.stops {
			key := agencyKey(s.Agency, s.StopID)
			ds.stopKeyToIdx[key] = len(ds.stops)
			ds.stops = append(ds.stops, s)
		}
	}

	for fi, f := range feeds {
		offset := feedStopOffset[fi]
		for _, t := range f.trips {
			svc, ok := f.services[t.serviceID]
			if !ok || !svc.activeOn(targetDate) {
				continue
			}
			sts := t.stopTimes
			for i := 0; i+1 < len(sts); i++ {
				ds.connections = append(ds.connections, connection{
					depStop: offset + sts[i].stopIdx,
					arrStop: offset + sts[i+1].stopIdx,
					depSec:  sts[i].depSec,
					arrSec:  sts[i+1].arrSec,
				})
			}
		}
	}

	sort.Slice(ds.connections, func(i, j int) bool {
		return ds.connections[i].arrSec > ds.connections[j].arrSec
	})

	ds.transfers = buildTransfers(ds.stops, ds.stopKeyToIdx, feeds, walkSpeedMps, walkDetour)

	return ds, nil
}

func (d *Dataset) Stops() []Stop {
	out := make([]Stop, len(d.stops))
	copy(out, d.stops)
	return out
}

func (d *Dataset) StopCount() int        { return len(d.stops) }
func (d *Dataset) ConnectionCount() int  { return len(d.connections) }

func nextWeekday(day time.Weekday) time.Time {
	now := time.Now()
	offset := (int(day) - int(now.Weekday()) + 7) % 7
	if offset == 0 {
		offset = 7
	}
	return time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, now.Location())
}

func agencyKey(agency, stopID string) string {
	return agency + ":" + stopID
}

func PrefixedStopID(agency, stopID string) string {
	return agencyKey(agency, stopID)
}

func buildTransfers(stops []Stop, idx map[string]int, feeds []*feed, walkSpeedMps, walkDetour float64) [][]transfer {
	transfers := make([][]transfer, len(stops))

	for _, f := range feeds {
		for _, t := range f.transfers {
			from, ok1 := idx[agencyKey(f.agency, t.fromStopID)]
			to, ok2 := idx[agencyKey(f.agency, t.toStopID)]
			if !ok1 || !ok2 || from == to {
				continue
			}
			sec := t.minSec
			if sec <= 0 {
				sec = 60
			}
			transfers[from] = append(transfers[from], transfer{fromStop: from, toStop: to, seconds: sec})
		}
	}

	const walkRadiusMeters = 500.0
	walkCellSize := walkRadiusMeters / 111000.0
	cells := make(map[[2]int][]int)
	for i, s := range stops {
		cx := int(math.Floor(s.Longitude / walkCellSize))
		cy := int(math.Floor(s.Latitude / walkCellSize))
		cells[[2]int{cx, cy}] = append(cells[[2]int{cx, cy}], i)
	}

	for i, s := range stops {
		cx := int(math.Floor(s.Longitude / walkCellSize))
		cy := int(math.Floor(s.Latitude / walkCellSize))
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				for _, j := range cells[[2]int{cx + dx, cy + dy}] {
					if j == i {
						continue
					}
					m := haversineMeters(s.Latitude, s.Longitude, stops[j].Latitude, stops[j].Longitude)
					if m > walkRadiusMeters {
						continue
					}
					sec := int(m * walkDetour / walkSpeedMps)
					if sec < 30 {
						sec = 30
					}
					transfers[i] = append(transfers[i], transfer{fromStop: i, toStop: j, seconds: sec})
				}
			}
		}
	}

	return transfers
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000.0
	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	dLa := (lat2 - lat1) * math.Pi / 180
	dLo := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLa/2)*math.Sin(dLa/2) + math.Cos(la1)*math.Cos(la2)*math.Sin(dLo/2)*math.Sin(dLo/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

func (d *Dataset) FindNearestStop(lat, lon float64) (int, error) {
	if len(d.stops) == 0 {
		return -1, fmt.Errorf("dataset has no stops")
	}
	best := -1
	bestDist := math.MaxFloat64
	for i, s := range d.stops {
		dist := haversineMeters(lat, lon, s.Latitude, s.Longitude)
		if dist < bestDist {
			bestDist = dist
			best = i
		}
	}
	return best, nil
}

func (d *Dataset) FindStopsByName(query string) []int {
	out := []int{}
	for i, s := range d.stops {
		if containsFold(s.Name, query) {
			out = append(out, i)
		}
	}
	return out
}

func containsFold(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	lowerS := toLower(s)
	lowerSub := toLower(sub)
	return indexOf(lowerS, lowerSub) >= 0
}

func toLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func indexOf(s, sub string) int {
	n := len(sub)
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sub {
			return i
		}
	}
	return -1
}
