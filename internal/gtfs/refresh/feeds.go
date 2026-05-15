package refresh

import "github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs"

type FeedDef struct {
	Agency string
	URL    string
}

func (f FeedDef) toSource(zipPath string) gtfs.FeedSource {
	return gtfs.FeedSource{Agency: f.Agency, ZipPath: zipPath}
}

var DefaultFeeds = []FeedDef{
	{Agency: "stm", URL: "https://www.stm.info/sites/default/files/gtfs/gtfs_stm.zip"},
	{Agency: "stl", URL: "https://www.stlaval.ca/datas/opendata/GTF_STL.zip"},
	{Agency: "rtl", URL: "https://www.rtl-longueuil.qc.ca/transit/latestfeed/RTL.zip"},
	{Agency: "rem", URL: "https://gtfs.gpmmom.ca/gtfs/gtfs.zip"},
	{Agency: "exo-trains", URL: "https://exo.quebec/xdata/trains/google_transit.zip"},
	{Agency: "exo-citcrc", URL: "https://exo.quebec/xdata/citcrc/google_transit.zip"},
	{Agency: "exo-citla", URL: "https://exo.quebec/xdata/citla/google_transit.zip"},
	{Agency: "exo-citpi", URL: "https://exo.quebec/xdata/citpi/google_transit.zip"},
	{Agency: "exo-citrous", URL: "https://exo.quebec/xdata/citrous/google_transit.zip"},
	{Agency: "exo-citsv", URL: "https://exo.quebec/xdata/citsv/google_transit.zip"},
	{Agency: "exo-citso", URL: "https://exo.quebec/xdata/citso/google_transit.zip"},
	{Agency: "exo-citvr", URL: "https://exo.quebec/xdata/citvr/google_transit.zip"},
	{Agency: "exo-mrclasso", URL: "https://exo.quebec/xdata/mrclasso/google_transit.zip"},
	{Agency: "exo-mrclm", URL: "https://exo.quebec/xdata/mrclm/google_transit.zip"},
	{Agency: "exo-omitsju", URL: "https://exo.quebec/xdata/omitsju/google_transit.zip"},
	{Agency: "exo-lrrs", URL: "https://exo.quebec/xdata/lrrs/google_transit.zip"},
}
