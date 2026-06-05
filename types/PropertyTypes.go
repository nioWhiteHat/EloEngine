package types

import (
	"context"
	

	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Builder struct{
	Db *pgx.Conn
	Properties []Property
	PropertiesModeled []PropertyModel
	Query string
}

func BuildQuery() string {
	return `
	SELECT 
    bd.construction_year,
    p.uid,
    p.subtype,
    bd.landplot_size,
    p.ad_creation_date,
    v.asking_price,
    ba.square_meters,
    COALESCE(bd.renovation_year::text, 'null') AS renovation_year,
    ba.view,
    ba.facade_type,
    bd.heat,
    bd.heat_from,
    ba.marketabillity_factor,
    ba.zone_value,
    CASE 
        WHEN bd.wc IS NOT NULL AND bd.bathrooms IS NULL THEN bd.wc
        WHEN bd.wc IS NOT NULL AND bd.bathrooms IS NOT NULL AND bd.wc > bd.bathrooms THEN bd.wc
        WHEN bd.wc IS NULL AND bd.bathrooms IS NULL THEN NULL
        ELSE 0 
    END AS wc,
    CASE 
        WHEN bd.bathrooms IS NOT NULL AND bd.wc IS NULL THEN bd.bathrooms
        WHEN bd.bathrooms IS NOT NULL AND bd.wc IS NOT NULL AND bd.bathrooms >= bd.wc THEN bd.bathrooms
        WHEN bd.wc IS NULL AND bd.bathrooms IS NULL THEN NULL
        ELSE 0 
    END AS bathrooms,
    bd.ac,
    bd.elevator,
    bd.parking,
    bd.storage_room,
    bd.aluminum_frames,
    bd.floor_level,
    bd.energy_performance_certificate,
    rs.bedrooms,
    rs.pool,
    rs.garden,
    loc.geo_lat,
    loc.geo_lon,
    ab.name AS admin_boundary_name
FROM building_details bd 
JOIN property_geo_data loc ON bd.property_uid = loc.property_uid 
JOIN admin_boundaries ab ON ab.admin_level = 6 AND ST_Contains(ab.geom, loc.geom)
JOIN base_attributes ba ON bd.property_uid = ba.property_uid
JOIN properties p ON bd.property_uid = p.uid
JOIN valuations v ON bd.property_uid = v.property_uid
JOIN residential_specifics rs ON rs.property_uid = p.uid
WHERE bd.construction_year IS NOT NULL 
    AND ba.square_meters > 10 
    AND p.transaction_type = $1
    AND p.item_type = 'residence'
    AND p.subtype IN ('apartment','detatched_house','maisonette')
    AND bd.under_construction = false
    AND bd.construction_year <= 2026
	AND bd.construction_year >= 1900
    AND v.asking_price > 1000
ORDER BY 
    bd.construction_year ASC, 
    loc.geo_lat ASC, 
    loc.geo_lon ASC;
	
	`
}

func (b *Builder) Initialize(transactionType string) (err error) {
	var properties []Property

	rows, err := b.Db.Query(context.Background(), b.Query, transactionType)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p Property
	
		

		err := rows.Scan(
			&p.Construction_year,
			&p.UID,
			&p.SubType,
			&p.LandplotSize,
			&p.AdCreationDate,
			&p.AskingPrice,
			&p.SquareMeters,
			&p.RenovationYear,
			&p.View,
			&p.FacadeType,
			&p.Heat,
			&p.HeatFrom,
			&p.MarketabillityFactor,
			&p.ZoneValue,
			&p.WC,
			&p.Bathrooms,
			&p.AC,
			&p.Elevator,
			&p.Parking,
			&p.StorageRoom,
			&p.AluminumFrames,
			&p.FloorLevel,
			&p.EnergyPerformance,
			&p.Bedrooms,
			&p.Pool,
			&p.Garden,
			&p.GeoLat,
			&p.GeoLon,
			&p.AdminName,
		)
		if err != nil {
			return err
		}

		properties = append(properties, p)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	
	b.Properties = properties
	b.PropertiesModeled = ConvertProperties(b.Properties)

	return nil
}
        


type PropertyModel struct {
	UID                    string
	
	Age                    int
	RenovAge               int
	Distance               int
	IsRenovated            int
	AgeSq                  int
	InteractionAgeDistance int
	LogInteraction         float64
	SubType                string
	              
	AdCreationAge         int
	AskingPricePerSqm      float64
	SquareMeters           int
	LandplotSize           int
	NoView                 int
	ViewForest             int
	ViewMountain           int
	ViewSea                int
	ViewUnlimited          int
	NoFacade               int
	FacadeCorner           int
	FacadeFrontage         int
	FacadeInterior         int
	FacadeThrough          int
	HeatType               string
	HeatFrom               string
	EnergyPerformance      string
	WC                     int
	AC                     int
	Elevator               int
	Parking                int
	StorageRoom            int
	AluminumFrames         int
	Bedrooms               int
	Pool                   int
	Garden                 int
	FloorLevel             int
	FloorCount             int
	GeoLat                 float64
	GeoLon                 float64
	ZoneValue			   int
	MarketabillityFactor	   float64
	AskingPrice            int
	AdminName			   string
}



type Property struct {
	Construction_year int     `json:"construction_year"`
	
	UID               *string `json:"uid"`
	SubType           *string `json:"subtype"`
	LandplotSize      *int    `json:"landplot_size"`
	
	AdCreationDate *time.Time `json:"ad_creation_date"`
	AskingPrice       *int    `json:"asking_price"`
	RealValue         *int    `json:"real_value"`
	EstimatedValue    *int    `json:"estimated_value"`
	SquareMeters      int     `json:"square_meters"`
	RenovationYear    string  `json:"renovation_year"`
	AdminName		   *string `json:"admin_boundary_name"`
	View              []string `json:"view"`
	FacadeType        []string `json:"facade_type"`
	Heat              *string  `json:"heat"`
	HeatFrom          *string  `json:"heat_from"`
	MarketabillityFactor *float64 `json:"marketabillity_factor"`
	ZoneValue *int `json:"zone_value"`
	WC                *int     `json:"wc"`
	Bathrooms         *int     `json:"bathrooms"`
	AC                *bool    `json:"ac"`
	Elevator          *bool    `json:"elevator"`
	Parking           *bool    `json:"parking"`
	StorageRoom       *bool    `json:"storage_room"`
	AluminumFrames    *bool    `json:"aluminum_frames"`
	FloorLevel        []string `json:"floor_level"`
	EnergyPerformance *string  `json:"energy_performance_certificate"`
	Bedrooms          *int     `json:"bedrooms"`
	Pool              *bool    `json:"pool"`
	Garden            *bool    `json:"garden"`
	GeoLat            *float64 `json:"geo_lat"`
	GeoLon            *float64 `json:"geo_lon"`

}

func SanitizeDynamic(props []PropertyModel) []PropertyModel {
	if len(props) == 0 {
		return props
	}

	var sumSqm, sumPrice, sumBeds float64
	for _, p := range props {
		sumSqm += float64(p.SquareMeters)
		sumPrice += float64(p.AskingPrice)
		sumBeds += float64(p.Bedrooms)
	}

	n := float64(len(props))
	meanSqm := sumSqm / n
	meanPrice := sumPrice / n
	meanBeds := sumBeds / n

	var varSqm, varPrice, varBeds float64
	for _, p := range props {
		varSqm += math.Pow(float64(p.SquareMeters)-meanSqm, 2)
		varPrice += math.Pow(float64(p.AskingPrice)-meanPrice, 2)
		varBeds += math.Pow(float64(p.Bedrooms)-meanBeds, 2)
	}

	stdSqm := math.Sqrt(varSqm / n)
	stdPrice := math.Sqrt(varPrice / n)
	stdBeds := math.Sqrt(varBeds / n)

	var clean []PropertyModel
	for _, p := range props {
		zSqm := math.Abs(float64(p.SquareMeters)-meanSqm) / stdSqm
		zPrice := math.Abs(float64(p.AskingPrice)-meanPrice) / stdPrice
		zBeds := math.Abs(float64(p.Bedrooms)-meanBeds) / stdBeds

		if zSqm > 3 || zPrice > 3 || zBeds > 3 {
			continue
		}
		clean = append(clean, p)
	}

	return clean
}

func ConvertProperties(raw []Property) []PropertyModel {
	var result []PropertyModel
	currentYear := 2026

	for _, r := range raw {
		var n PropertyModel

		if r.UID != nil {
			n.UID = *r.UID
		}

		n.Age = currentYear - r.Construction_year

		if r.RenovationYear == "null" || r.RenovationYear == "" {
			n.RenovAge = n.Age
		} else {
			if ry, err := strconv.Atoi(r.RenovationYear); err == nil {
				n.RenovAge = currentYear - ry
			} else {
				n.RenovAge = n.Age
			}
		}

		if r.SubType != nil {
			n.SubType = *r.SubType
		}

		if r.LandplotSize != nil {
			n.LandplotSize = *r.LandplotSize
		} else {
			n.LandplotSize = r.SquareMeters
		}

		n.AdCreationAge = parseMonthsSince(r.AdCreationDate)

		n.SquareMeters = r.SquareMeters

		n.ViewForest, n.ViewMountain, n.ViewSea, n.ViewUnlimited, n.NoView = parseView(r.View)
		n.FacadeCorner, n.FacadeFrontage, n.FacadeInterior, n.FacadeThrough, n.NoFacade = parseFacade(r.FacadeType)

		if r.Heat != nil {
			n.HeatType = *r.Heat
		}

		if r.HeatFrom != nil {
			n.HeatFrom = *r.HeatFrom
		}

		if r.EnergyPerformance != nil {
			n.EnergyPerformance = *r.EnergyPerformance
		}

		wcCount := 0
		if r.WC != nil {
			wcCount += *r.WC
		}
		if r.Bathrooms != nil {
			wcCount += *r.Bathrooms
		}
		n.WC = wcCount

		if r.Pool != nil && *r.Pool {
			n.Pool = 1
		} else {
			n.Pool = 0
		}

		if r.Garden != nil && *r.Garden {
			n.Garden = 1
		} else {
			n.Garden = 0
		}

		if r.Bedrooms != nil {
			n.Bedrooms = *r.Bedrooms
		} else {
			continue
		}

		if r.AC != nil && *r.AC {
			n.AC = 1
		} else {
			n.AC = 0
		}

		if r.Elevator != nil && *r.Elevator {
			n.Elevator = 1
		} else {
			n.Elevator = 0
		}

		if r.Parking != nil && *r.Parking {
			n.Parking = 1
		} else {
			n.Parking = 0
		}

		if r.StorageRoom != nil && *r.StorageRoom {
			n.StorageRoom = 1
		} else {
			n.StorageRoom = 0
		}

		if r.AluminumFrames != nil && *r.AluminumFrames {
			n.AluminumFrames = 1
		} else {
			n.AluminumFrames = 0
		}

		var FloorLevel int
		FloorLevel, n.FloorCount = parseFloors(r.FloorLevel)

		if n.Age > n.RenovAge {
			n.IsRenovated = 1
		} else {
			n.IsRenovated = 0
		}

		n.FloorLevel = FloorLevel
		n.GeoLat = *r.GeoLat
		n.GeoLon = *r.GeoLon

		if r.AskingPrice != nil && r.SquareMeters > 0 {
			n.AskingPricePerSqm = float64(*r.AskingPrice) / float64(r.SquareMeters)
		} else {
			n.AskingPricePerSqm = 0.0
		}
		if r.AdminName != nil {
			n.AdminName = *r.AdminName
		}


		n.Distance = n.Age - n.RenovAge
		n.InteractionAgeDistance = n.Age * n.Distance
		n.AgeSq = n.Age * n.Age
		n.LogInteraction = math.Log(float64(n.Age)+1.0) * float64(n.Distance)
		n.MarketabillityFactor = *r.MarketabillityFactor
		n.ZoneValue = *r.ZoneValue
		
		result = append(result, n)
	}

	return SanitizeDynamic(result)
}
func parseMonthsSince(t *time.Time) int {
	if t == nil {
		return 0
	}
	
	now := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	months := (now.Year()-t.Year())*12 + int(now.Month()-t.Month())
	
	if months < 0 {
		return 0
	}
	return months
}
func parseView(viewSlice []string) (int, int, int, int,int) {
	var forest, mountain, sea, unlimited, NoView int
	for _, rawString := range viewSlice {
		cleanStr := strings.Trim(rawString, "{}")
		traits := strings.Split(cleanStr, ",")
		for _, trait := range traits {
			switch strings.TrimSpace(trait) {
			case "forest":
				forest = 1
			case "mountain":
				mountain = 1
			case "sea":
				sea = 1
			case "unlimited":
				unlimited = 1
			default:
				NoView = 1
			}
		}
	}
	return forest, mountain, sea, unlimited, NoView
}

func parseFacade(facadeSlice []string) (int, int, int, int,int) {
	var corner, frontage, interior, through, None int
	for _, rawString := range facadeSlice {
		cleanStr := strings.Trim(rawString, "{}")
		traits := strings.Split(cleanStr, ",")
		for _, trait := range traits {
			switch strings.TrimSpace(trait) {
			case "corner":
				corner = 1
			case "frontage":
				frontage = 1
			case "interior":
				interior = 1
			case "through":
				through = 1
			default:
				None = 1
			}
		}
	}
	return corner, frontage, interior, through, None
}
func parseFloors(floors []string) (int, int) {
	if len(floors) == 0 {
		return 0, 1
	}
	minF := 2147483647
	uniqueFloors := make(map[int]struct{})
	for _, f := range floors {
		cleanF := strings.Trim(f, "{}")
		if val, err := strconv.Atoi(cleanF); err == nil {
			uniqueFloors[val] = struct{}{}
			if val < minF {
				minF = val
			}
		}
	}
	if minF == 2147483647 {
		return 0, 1
	}
	return minF, len(uniqueFloors)
}