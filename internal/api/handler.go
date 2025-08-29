package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
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

	// Добавляем поддержку фильтрации по датам и сетям
	var whereClauses []string
	var args []interface{}
	argIndex := 1
	
	// Определяем тип фильтрации по IP
	var ipClause string
	if strings.Contains(ipAddress, "/") {
		// CIDR notation (e.g., 192.168.1.0/24)
		_, ipNet, err := net.ParseCIDR(ipAddress)
		if err != nil {
			http.Error(w, "Invalid CIDR notation: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Используем PostgreSQL оператор << для проверки вхождения в сеть
		ipClause = "(src_ip << $" + strconv.Itoa(argIndex) + " OR dst_ip << $" + strconv.Itoa(argIndex) + ")"
		args = append(args, ipNet.String())
		argIndex++
	} else if strings.Contains(ipAddress, "*") {
		// Wildcard notation (e.g., 192.168.1.*)
		likePattern := strings.ReplaceAll(ipAddress, "*", "%")
		ipClause = "(src_ip LIKE $" + strconv.Itoa(argIndex) + " OR dst_ip LIKE $" + strconv.Itoa(argIndex) + ")"
		args = append(args, likePattern)
		argIndex++
	} else {
		// Точное совпадение IP
		ipClause = "(src_ip = $" + strconv.Itoa(argIndex) + " OR dst_ip = $" + strconv.Itoa(argIndex) + ")"
		args = append(args, ipAddress)
		argIndex++
	}
	
	whereClauses = append(whereClauses, ipClause)
	
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
	err := h.db.Get(&totalRecords, countQuery, args...)
	if err != nil {
		log.Printf("Error counting flows: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Получаем статистику (общие суммы)
	statsQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(bytes), 0) as total_bytes
		FROM flows 
		%s
	`, whereClause)
	
	var stats struct {
		TotalBytes int64 `db:"total_bytes"`
	}
	
	err = h.db.Get(&stats, statsQuery, args...)
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
	
	flowArgs := append(args, limit, offset)
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
		// Для сетевых масок и wildcard поиска сложнее определить направление,
		// поэтому просто помечаем как mixed если это не точное совпадение IP
		if strings.Contains(ipAddress, "/") || strings.Contains(ipAddress, "*") {
			flow.Direction = "mixed"
			flow.BytesIn = flow.Bytes / 2  // Примерное разделение
			flow.BytesOut = flow.Bytes / 2
		} else if flow.DstIP == ipAddress {
			// Входящий трафик (точное совпадение)
			flow.Direction = "incoming"
			flow.BytesIn = flow.Bytes
			flow.BytesOut = 0
		} else {
			// Исходящий трафик (точное совпадение)
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
		TotalBytesIn:  stats.TotalBytes / 2, // Примерное разделение для сетей
		TotalBytesOut: stats.TotalBytes / 2,
		TotalTraffic:  stats.TotalBytes,
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
