package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP    HTTPConfig
	DB      DBConfig
	Auth    AuthConfig
	Realtor RealtorConfig
	Transit TransitConfig
}

type AuthConfig struct {
	AdminUsername string
	AdminPassword string
	AdminReset    bool
	CookieSecure  bool
	TrustProxy    bool
	SessionTTL    time.Duration
}

type HTTPConfig struct {
	Addr string
}

type DBConfig struct {
	URL string
}

type TransitConfig struct {
	ReferenceKey string
	ReferenceLat float64
	ReferenceLon float64
	SnapshotKey  string
	SnapshotDay  time.Weekday
	SnapshotHour int
	NearestStops int
	WalkSpeedMps float64
	WalkDetour   float64
}

type RealtorConfig struct {
	BaseURL      string
	Mock         bool
	LatitudeMax  string
	LatitudeMin  string
	LongitudeMax string
	LongitudeMin string
}

func Load() (*Config, error) {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return nil, errors.New("POSTGRES_USER is required")
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return nil, errors.New("POSTGRES_PASSWORD is required")
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		return nil, errors.New("POSTGRES_DB is required")
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:3000"
	}

	realtorBaseURL := os.Getenv("REALTOR_BASE_URL")
	if realtorBaseURL == "" {
		realtorBaseURL = "https://api2.realtor.ca"
	}

	sslMode := ""
	if os.Getenv("POSTGRES_ENABLE_SSL") == "false" {
		sslMode = "?sslmode=disable"
	}

	transit, err := loadTransitConfig()
	if err != nil {
		return nil, err
	}

	authCfg, err := loadAuthConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		HTTP: HTTPConfig{
			Addr: addr,
		},
		DB: DBConfig{
			URL: fmt.Sprintf("postgres://%s:%s@%s:%s/%s%s", user, password, host, port, dbName, sslMode),
		},
		Auth: authCfg,
		Realtor: RealtorConfig{
			BaseURL:      realtorBaseURL,
			Mock:         os.Getenv("MOCK_REALTOR_API") == "true",
			LatitudeMax:  os.Getenv("LAT_MAX"),
			LatitudeMin:  os.Getenv("LAT_MIN"),
			LongitudeMax: os.Getenv("LON_MAX"),
			LongitudeMin: os.Getenv("LON_MIN"),
		},
		Transit: transit,
	}, nil
}

var weekdayNames = map[string]time.Weekday{
	"sun": time.Sunday,
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
}

func loadTransitConfig() (TransitConfig, error) {
	snapshot := os.Getenv("TRANSIT_SNAPSHOT")
	if snapshot == "" {
		snapshot = "tue-09"
	}
	day, hour, err := parseSnapshot(snapshot)
	if err != nil {
		return TransitConfig{}, fmt.Errorf("TRANSIT_SNAPSHOT: %w", err)
	}

	refLat, err := parseFloatEnv("TRANSIT_REF_LAT", 45.5048)
	if err != nil {
		return TransitConfig{}, err
	}
	refLon, err := parseFloatEnv("TRANSIT_REF_LON", -73.5772)
	if err != nil {
		return TransitConfig{}, err
	}
	refKey := os.Getenv("TRANSIT_REF_KEY")
	if refKey == "" {
		refKey = "mcgill"
	}

	nearest, err := parseIntEnv("TRANSIT_NEAREST_STOPS", 5)
	if err != nil {
		return TransitConfig{}, err
	}
	walkSpeed, err := parseFloatEnv("TRANSIT_WALK_SPEED_MPS", 1.39)
	if err != nil {
		return TransitConfig{}, err
	}
	walkDetour, err := parseFloatEnv("TRANSIT_WALK_DETOUR", 1.3)
	if err != nil {
		return TransitConfig{}, err
	}

	return TransitConfig{
		ReferenceKey: refKey,
		ReferenceLat: refLat,
		ReferenceLon: refLon,
		SnapshotKey:  snapshot,
		SnapshotDay:  day,
		SnapshotHour: hour,
		NearestStops: nearest,
		WalkSpeedMps: walkSpeed,
		WalkDetour:   walkDetour,
	}, nil
}

func parseSnapshot(s string) (time.Weekday, int, error) {
	parts := strings.Split(strings.ToLower(s), "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format %q, expected ddd-hh (e.g. tue-09)", s)
	}
	day, ok := weekdayNames[parts[0]]
	if !ok {
		return 0, 0, fmt.Errorf("invalid day %q", parts[0])
	}
	hour, err := strconv.Atoi(parts[1])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour %q", parts[1])
	}
	return day, hour, nil
}

func parseFloatEnv(name string, def float64) (float64, error) {
	v := os.Getenv(name)
	if v == "" {
		return def, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid float %q", name, v)
	}
	return f, nil
}

func parseIntEnv(name string, def int) (int, error) {
	v := os.Getenv(name)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer %q", name, v)
	}
	return n, nil
}

func loadAuthConfig() (AuthConfig, error) {
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		return AuthConfig{}, errors.New("ADMIN_USERNAME is required")
	}
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		return AuthConfig{}, errors.New("ADMIN_PASSWORD is required")
	}

	cookieSecure := true
	if os.Getenv("COOKIE_SECURE") == "false" {
		cookieSecure = false
	}

	sessionTTL := 720 * time.Hour
	if v := os.Getenv("SESSION_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return AuthConfig{}, fmt.Errorf("SESSION_TTL: invalid duration %q", v)
		}
		sessionTTL = d
	}

	return AuthConfig{
		AdminUsername: adminUsername,
		AdminPassword: adminPassword,
		AdminReset:    os.Getenv("ADMIN_RESET") == "true",
		CookieSecure:  cookieSecure,
		TrustProxy:    os.Getenv("TRUST_PROXY") == "true",
		SessionTTL:    sessionTTL,
	}, nil
}
