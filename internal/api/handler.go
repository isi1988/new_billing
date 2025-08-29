package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	// Дополнительные поля для определения направления трафика
	Direction    string `json:"direction"`    // "incoming" или "outgoing"
	BytesIn      int64  `json:"bytes_in"`    // Входящие байты (если direction = incoming)
	BytesOut     int64  `json:"bytes_out"`   // Исходящие байты (если direction = outgoing)
}

type FlowSearchResult struct {
	Flows         []FlowRecord `json:"flows"`
	TotalRecords  int          `json:"total_records"`
	TotalBytesIn  int64        `json:"total_bytes_in"`
	TotalBytesOut int64        `json:"total_bytes_out"`
	TotalTraffic  int64        `json:"total_traffic"`
	Page          int          `json:"page"`
	Limit         int          `json:"limit"`
	TotalPages    int          `json:"total_pages"`
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
		limit = 25
	}
	offset := (page - 1) * limit

	// Добавляем поддержку фильтрации по датам
	var whereClauses []string
	var args []interface{}
	argIndex := 2 // $1 используется для IP-адреса в двух местах
	
	whereClauses = append(whereClauses, "(src_ip = $1 OR dst_ip = $1)")
	
	if fromDate := r.URL.Query().Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			whereClauses = append(whereClauses, "timestamp >= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		}
	}
	
	if toDate := r.URL.Query().Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			endDate := parsedDate.Add(24 * time.Hour) // до конца дня
			whereClauses = append(whereClauses, "timestamp <= $"+strconv.Itoa(argIndex))
			args = append(args, endDate)
			argIndex++
		}
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

	// Получаем общее количество записей для пагинации
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM flows 
		%s
	`, whereClause)
	
	var totalRecords int
	countArgs := append([]interface{}{ipAddress}, args...)
	err := h.db.Get(&totalRecords, countQuery, countArgs...)
	if err != nil {
		log.Printf("Error counting flows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Получаем статистику (общие суммы)
	statsQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN dst_ip = $1 THEN bytes ELSE 0 END), 0) as total_bytes_in,
			COALESCE(SUM(CASE WHEN src_ip = $1 THEN bytes ELSE 0 END), 0) as total_bytes_out
		FROM flows 
		%s
	`, whereClause)
	
	var stats struct {
		TotalBytesIn  int64 `db:"total_bytes_in"`
		TotalBytesOut int64 `db:"total_bytes_out"`
	}
	
	err = h.db.Get(&stats, statsQuery, countArgs...)
	if err != nil {
		log.Printf("Error getting flow statistics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Получаем flows с пагинацией
	flowQuery := fmt.Sprintf(`
		SELECT timestamp, src_ip, dst_ip, src_port, dst_port, protocol, packets, bytes
		FROM flows
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	
	flowArgs := append(countArgs, limit, offset)
	var flowsFromDB []FlowRecord
	err = h.db.Select(&flowsFromDB, flowQuery, flowArgs...)
	if err != nil {
		log.Printf("Error searching flows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Определяем направление трафика и заполняем поля
	var flows []FlowRecord
	for _, flow := range flowsFromDB {
		if flow.DstIP == ipAddress {
			// Входящий трафик
			flow.Direction = "incoming"
			flow.BytesIn = flow.Bytes
			flow.BytesOut = 0
		} else {
			// Исходящий трафик  
			flow.Direction = "outgoing"
			flow.BytesIn = 0
			flow.BytesOut = flow.Bytes
		}
		flows = append(flows, flow)
	}

	totalPages := (totalRecords + limit - 1) / limit
	
	result := FlowSearchResult{
		Flows:         flows,
		TotalRecords:  totalRecords,
		TotalBytesIn:  stats.TotalBytesIn,
		TotalBytesOut: stats.TotalBytesOut,
		TotalTraffic:  stats.TotalBytesIn + stats.TotalBytesOut,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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
		SELECT date_trunc('%s', timestamp) as time_period, COALESCE(SUM(bytes), 0) as total_bytes
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
