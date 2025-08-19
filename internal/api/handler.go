package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type APIHandler struct {
	db *sqlx.DB
}

func NewAPIHandler(db *sqlx.DB) *APIHandler {
	return &APIHandler{db: db}
}

type FlowRecord struct {
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	SrcIP     string    `db:"src_ip" json:"src_ip"`
	DstIP     string    `db:"dst_ip" json:"dst_ip"`
	SrcPort   int       `db:"src_port" json:"src_port"`
	DstPort   int       `db:"dst_port" json:"dst_port"`
	Protocol  int       `db:"protocol" json:"protocol"`
	Packets   int64     `db:"packets" json:"packets"`
	Bytes     int64     `db:"bytes" json:"bytes"`
}

type AggregationResult struct {
	TimePeriod time.Time `db:"time_period" json:"time_period"`
	TotalBytes int64     `db:"total_bytes" json:"total_bytes"`
}

func (h *APIHandler) SearchFlows(w http.ResponseWriter, r *http.Request) {
	ipAddress := r.URL.Query().Get("ip")
	if ipAddress == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := `
		SELECT timestamp, src_ip, dst_ip, src_port, dst_port, protocol, packets, bytes
		FROM flows
		WHERE src_ip = $1 OR dst_ip = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	var flows []FlowRecord
	err := h.db.Select(&flows, query, ipAddress, limit, offset)
	if err != nil {
		log.Printf("Error searching flows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Определение входящего/исходящего трафика
	// Это лучше делать на клиенте или добавить поле в ответ

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flows)
}

func (h *APIHandler) AggregateFlows(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	granularity := r.URL.Query().Get("granularity") // minute, hour, day, month

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start_time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end_time format", http.StatusBadRequest)
		return
	}

	//dateFormat := "YYYY-MM-DD HH24:MI:00"
	//switch granularity {
	//case "hour":
	//	dateFormat = "YYYY-MM-DD HH24:00:00"
	//case "day":
	//	dateFormat = "YYYY-MM-DD 00:00:00"
	//case "month":
	//	dateFormat = "YYYY-MM-01 00:00:00"
	//}

	query := fmt.Sprintf(`
		SELECT date_trunc('%s', timestamp) as time_period, SUM(bytes) as total_bytes
		FROM flows
		WHERE timestamp BETWEEN $1 AND $2
		GROUP BY time_period
		ORDER BY time_period
	`, granularity)

	var results []AggregationResult
	err = h.db.Select(&results, query, startTime, endTime)
	if err != nil {
		log.Printf("Error aggregating flows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
