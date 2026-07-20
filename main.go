package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// 認証用のIDとパスワード（お好みの文字列に変更してください）
const basicUser = "admin"
const basicPass = "password123"

// Basic認証用のミドルウェア関数
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != basicUser || pass != basicPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// Charger represents a Tesla Supercharger
type Charger struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	City     string `json:"city"`
	GPS      string `json:"gps"`
	Lat      string `json:"lat"`
	Lng      string `json:"lng"`
	Stalls   string `json:"stalls"`
	Power    string `json:"power"`
	AreaType string `json:"AreaType"`
}

// Hotel represents a Marriott Hotel
type Hotel struct {
	Brand   string `json:"brand"`
	Name    string `json:"name"`
	Address string `json:"address"`
	GPS     string `json:"gps"`
	Lat     string `json:"lat"`
	Lng     string `json:"lng"`
}

// BaseParam represents vehicle parameters from base.csv
type BaseParam struct {
	BattFull float64 `json:"batt_full"`
	BattBuff float64 `json:"batt_buff"`
	SoCStart float64 `json:"soc_start"`
	SoCMid   float64 `json:"soc_mid"`
	SoCDest  float64 `json:"soc_dest"`
}

type SimulationRequest struct {
	Coordinates []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinates"`
	TotalDistance float64   `json:"total_distance"`
	TargetSpeed   float64   `json:"target_speed"`
	DepTime       string    `json:"dep_time"`
	BattFull      float64   `json:"batt_full"`
	BattBuff      float64   `json:"batt_buff"`
	SoCStart      float64   `json:"soc_start"`
	Elevations    []float64 `json:"elevations"`
	Temperatures  []float64 `json:"temperatures"`
}

type SimulationSample struct {
	Dist float64 `json:"dist"`
	Time string  `json:"time"`
	Temp float64 `json:"temp"`
	Elev float64 `json:"elev"`
	EC   float64 `json:"ec"`
	SoC  float64 `json:"soc"`
}

type SimulationResponse struct {
	TotalDist  float64            `json:"total_dist"`
	AvgEC      float64            `json:"avg_ec"`
	ElevEnergy float64            `json:"elev_energy"`
	FinalSoC   float64            `json:"final_soc"`
	Samples    []SimulationSample `json:"samples"`
}

type GMKeyResponse struct {
	Available bool   `json:"available"`
	Key       string `json:"key"`
}

func parseGPS(gps string) (string, string) {
	gps = strings.Trim(gps, "\" ")
	parts := strings.Split(gps, ",")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", ""
}

func chargersHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("sc.csv")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	_, _ = reader.Read()

	var chargers []Charger
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		lat, lng := parseGPS(record[6])
		chargers = append(chargers, Charger{
			Name:     record[0],
			Address:  record[1],
			City:     record[2],
			GPS:      record[6],
			Lat:      lat,
			Lng:      lng,
			Stalls:   record[8],
			Power:    record[9],
			AreaType: record[12],
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chargers)
}

func mariottoHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("mariotto.csv")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	_, _ = reader.Read()

	var hotels []Hotel
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		lat, lng := parseGPS(record[3])
		hotels = append(hotels, Hotel{
			Brand:   record[0],
			Name:    record[1],
			Address: record[2],
			GPS:     record[3],
			Lat:     lat,
			Lng:     lng,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hotels)
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("base.csv")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	var params BaseParam
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 2 {
			continue
		}
		key := record[0]
		var val float64
		fmt.Sscanf(record[1], "%f", &val)

		switch key {
		case "batt_full":
			params.BattFull = val
		case "batt_buff":
			params.BattBuff = val
		case "soc_start":
			params.SoCStart = val
		case "soc_mid":
			params.SoCMid = val
		case "soc_dest":
			params.SoCDest = val
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params)
}

func ecHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("ec.csv")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func gmkeyHandler(w http.ResponseWriter, r *http.Request) {
	var res GMKeyResponse
	res.Available = false
	res.Key = ""

	f, err := os.Open("gmkey.csv")
	if err == nil {
		defer f.Close()
		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err == nil && len(records) > 0 && len(records[0]) > 0 {
			key := strings.TrimSpace(records[0][0])
			if key != "" {
				res.Available = true
				res.Key = key
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func GetInterpolatedEC(matrix [][]string, temp float64, speed float64) float64 {
	if len(matrix) < 2 {
		return 0.0
	}
	header := matrix[0]
	var speeds []float64
	for i := 1; i < len(header); i++ {
		var s float64
		fmt.Sscanf(header[i], "%f", &s)
		speeds = append(speeds, s)
	}
	var sIdx1 int = 0
	var sIdx2 int = 0
	if speed <= speeds[0] {
		sIdx1 = 0
		sIdx2 = 0
	} else if speed >= speeds[len(speeds)-1] {
		sIdx1 = len(speeds) - 1
		sIdx2 = len(speeds) - 1
	} else {
		for i := 0; i < len(speeds)-1; i++ {
			if speed >= speeds[i] && speed <= speeds[i+1] {
				sIdx1 = i
				sIdx2 = i + 1
				break
			}
		}
	}
	rows := matrix[1:]
	var temps []float64
	for _, row := range rows {
		var t float64
		fmt.Sscanf(row[0], "%f", &t)
		temps = append(temps, t)
	}
	var tIdx1 int = 0
	var tIdx2 int = 0
	maxT := temps[0]
	minT := temps[len(temps)-1]
	if temp >= maxT {
		tIdx1 = 0
		tIdx2 = 0
	} else if temp <= minT {
		tIdx1 = len(temps) - 1
		tIdx2 = len(temps) - 1
	} else {
		for i := 0; i < len(temps)-1; i++ {
			if temp <= temps[i] && temp >= temps[i+1] {
				tIdx1 = i
				tIdx2 = i + 1
				break
			}
		}
	}
	var v11 float64 = 0.0
	var v12 float64 = 0.0
	var v21 float64 = 0.0
	var v22 float64 = 0.0
	fmt.Sscanf(rows[tIdx1][sIdx1+1], "%f", &v11)
	fmt.Sscanf(rows[tIdx1][sIdx2+1], "%f", &v12)
	fmt.Sscanf(rows[tIdx2][sIdx1+1], "%f", &v21)
	fmt.Sscanf(rows[tIdx2][sIdx2+1], "%f", &v22)
	var sx float64 = 0.0
	var tx float64 = 0.0
	if sIdx1 != sIdx2 {
		sx = (speed - speeds[sIdx1]) / (speeds[sIdx2] - speeds[sIdx1])
	}
	if tIdx1 != tIdx2 {
		tx = (temp - temps[tIdx1]) / (temps[tIdx2] - temps[tIdx1])
	}
	return (v11 + sx*(v12-v11)) + tx*((v21+sx*(v22-v21))-(v11+sx*(v12-v11)))
}

func HandleBackendSimulation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req SimulationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f, err := os.Open("ec.csv")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	matrix, err := reader.ReadAll()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var layout string = "2006-01-02T15:04"
	importTime, err := time.Parse(layout, req.DepTime)
	if err != nil {
		importTime = time.Now()
	}
	var currentDist float64 = 0.0
	var totalEnergyFlat float64 = 0.0
	var totalEnergyElev float64 = 0.0
	var samples []SimulationSample
	usableKWh := req.BattFull - req.BattBuff

	if len(req.Elevations) > 0 {
		samples = append(samples, SimulationSample{
			Dist: 0.0,
			Time: importTime.Format(layout),
			Temp: req.Temperatures[0],
			Elev: req.Elevations[0],
			EC:   0.0,
			SoC:  req.SoCStart,
		})
	}

	if len(req.Elevations) > 1 {
		stepCount := len(req.Elevations) - 1
		actualInterval := req.TotalDistance / float64(stepCount)

		for s := 1; s <= stepCount; s++ {
			if s >= len(req.Elevations) || s >= len(req.Temperatures) {
				break
			}
			targetDist := float64(s) * actualInterval
			if targetDist > req.TotalDistance {
				targetDist = req.TotalDistance
			}
			segmentDist := targetDist - currentDist
			elapsedHours := targetDist / req.TargetSpeed
			arrivalTime := importTime.Add(time.Duration(elapsedHours * float64(time.Hour)))
			temp := req.Temperatures[s]
			ec := GetInterpolatedEC(matrix, temp, req.TargetSpeed)
			segmentFlatWh := segmentDist * ec

			deltaH := req.Elevations[s] - req.Elevations[s-1]
			mghWh := (2200.0 * 9.81 * deltaH) / 3600.0
			var segmentElevWh float64 = 0.0
			if deltaH > 0 {
				segmentElevWh = mghWh / 0.9
			} else {
				segmentElevWh = mghWh * 0.8
			}

			totalEnergyFlat += segmentFlatWh
			totalEnergyElev += segmentElevWh
			currentDist = targetDist

			currentSoC := req.SoCStart - (((totalEnergyFlat+totalEnergyElev)/1000.0)/usableKWh)*100.0

			samples = append(samples, SimulationSample{
				Dist: currentDist,
				Time: arrivalTime.Format(layout),
				Temp: temp,
				Elev: req.Elevations[s],
				EC:   ec,
				SoC:  currentSoC,
			})
		}
	}
	var avgECFlat float64 = 0.0
	if req.TotalDistance > 0 {
		avgECFlat = totalEnergyFlat / req.TotalDistance
	}
	var finalSoC float64 = req.SoCStart
	if len(samples) > 0 {
		finalSoC = samples[len(samples)-1].SoC
	}
	res := SimulationResponse{
		TotalDist:  req.TotalDistance,
		AvgEC:      avgECFlat,
		ElevEnergy: totalEnergyElev,
		FinalSoC:   finalSoC,
		Samples:    samples,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	http.HandleFunc("/api/chargers", basicAuth(chargersHandler))
	http.HandleFunc("/api/mariotto", basicAuth(mariottoHandler))
	http.HandleFunc("/api/base", basicAuth(baseHandler))
	http.HandleFunc("/api/ec", basicAuth(ecHandler))
	http.HandleFunc("/api/gmkey", basicAuth(gmkeyHandler))
	http.HandleFunc("/api/simulation", basicAuth(HandleBackendSimulation))

	http.Handle("/", basicAuth(http.FileServer(http.Dir(".")).ServeHTTP))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
