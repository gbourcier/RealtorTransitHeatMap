package gtfs

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func parseFeed(src FeedSource) (*feed, error) {
	rc, err := zip.OpenReader(src.ZipPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src.ZipPath, err)
	}
	defer rc.Close()

	files := make(map[string]*zip.File)
	for _, f := range rc.File {
		files[strings.ToLower(f.Name)] = f
	}

	f := &feed{
		agency:   src.Agency,
		trips:    make(map[string]*trip),
		services: make(map[string]*service),
	}

	stopIndexByID := make(map[string]int)
	if err := readCSV(files, "stops.txt", func(row map[string]string) error {
		latS := row["stop_lat"]
		lonS := row["stop_lon"]
		if latS == "" || lonS == "" {
			return nil
		}
		lat, err := strconv.ParseFloat(latS, 64)
		if err != nil {
			return nil
		}
		lon, err := strconv.ParseFloat(lonS, 64)
		if err != nil {
			return nil
		}
		stop := Stop{
			Agency:    src.Agency,
			StopID:    row["stop_id"],
			Name:      row["stop_name"],
			Latitude:  lat,
			Longitude: lon,
		}
		stopIndexByID[stop.StopID] = len(f.stops)
		f.stops = append(f.stops, stop)
		return nil
	}); err != nil {
		return nil, err
	}

	if err := readCSV(files, "calendar.txt", func(row map[string]string) error {
		sid := row["service_id"]
		s, ok := f.services[sid]
		if !ok {
			s = &service{addedDates: map[string]bool{}, removedDays: map[string]bool{}}
			f.services[sid] = s
		}
		days := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
		for i, d := range days {
			if row[d] == "1" {
				s.weekdays[i] = true
			}
		}
		s.startDate, _ = time.Parse("20060102", row["start_date"])
		s.endDate, _ = time.Parse("20060102", row["end_date"])
		return nil
	}); err != nil {
		return nil, err
	}

	if err := readCSV(files, "calendar_dates.txt", func(row map[string]string) error {
		sid := row["service_id"]
		s, ok := f.services[sid]
		if !ok {
			s = &service{addedDates: map[string]bool{}, removedDays: map[string]bool{}}
			f.services[sid] = s
		}
		date := row["date"]
		switch row["exception_type"] {
		case "1":
			s.addedDates[date] = true
		case "2":
			s.removedDays[date] = true
		}
		return nil
	}); err != nil {
		return nil, err
	}

	tripService := make(map[string]string)
	if err := readCSV(files, "trips.txt", func(row map[string]string) error {
		tripService[row["trip_id"]] = row["service_id"]
		return nil
	}); err != nil {
		return nil, err
	}

	if err := readCSV(files, "stop_times.txt", func(row map[string]string) error {
		tid := row["trip_id"]
		sid, ok := tripService[tid]
		if !ok {
			return nil
		}
		stopIdx, ok := stopIndexByID[row["stop_id"]]
		if !ok {
			return nil
		}
		arr, ok1 := parseGTFSTime(row["arrival_time"])
		dep, ok2 := parseGTFSTime(row["departure_time"])
		if !ok1 || !ok2 {
			return nil
		}
		seq, _ := strconv.Atoi(row["stop_sequence"])

		t, exists := f.trips[tid]
		if !exists {
			t = &trip{serviceID: sid}
			f.trips[tid] = t
		}
		t.stopTimes = append(t.stopTimes, stopTime{
			stopIdx: stopIdx,
			arrSec:  arr,
			depSec:  dep,
			seq:     seq,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	for _, t := range f.trips {
		stops := t.stopTimes
		for i := 1; i < len(stops); i++ {
			j := i
			for j > 0 && stops[j-1].seq > stops[j].seq {
				stops[j-1], stops[j] = stops[j], stops[j-1]
				j--
			}
		}
	}

	if err := readCSV(files, "transfers.txt", func(row map[string]string) error {
		minSec, _ := strconv.Atoi(row["min_transfer_time"])
		f.transfers = append(f.transfers, rawTransfer{
			fromStopID: row["from_stop_id"],
			toStopID:   row["to_stop_id"],
			minSec:     minSec,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return f, nil
}

func readCSV(files map[string]*zip.File, name string, fn func(map[string]string) error) error {
	f, ok := files[name]
	if !ok {
		return nil
	}
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = true

	rawHeader, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header %s: %w", name, err)
	}
	header := make([]string, len(rawHeader))
	for i, h := range rawHeader {
		header[i] = strings.TrimSpace(strings.TrimPrefix(h, "\ufeff"))
	}

	row := make(map[string]string, len(header))
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		for k := range row {
			delete(row, k)
		}
		for i, v := range rec {
			if i < len(header) {
				row[header[i]] = v
			}
		}
		if err := fn(row); err != nil {
			return err
		}
	}
	return nil
}

func parseGTFSTime(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	sec, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, false
	}
	return h*3600 + m*60 + sec, true
}

func (s *service) activeOn(t time.Time) bool {
	dateStr := t.Format("20060102")
	if s.removedDays[dateStr] {
		return false
	}
	if s.addedDates[dateStr] {
		return true
	}
	if !s.startDate.IsZero() && t.Before(s.startDate) {
		return false
	}
	if !s.endDate.IsZero() && t.After(s.endDate.Add(24*time.Hour)) {
		return false
	}
	return s.weekdays[int(t.Weekday())]
}
