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
	
	// Получаем маску, если передана
	maskStr := r.URL.Query().Get("mask")
	var subnet string

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
	if maskStr != "" {
		// Формируем CIDR из IP и маски
		mask, err := strconv.Atoi(maskStr)
		if err != nil || mask < 0 || mask > 32 {
			http.Error(w, "Invalid mask: must be between 0 and 32", http.StatusBadRequest)
			return
		}
		subnet = fmt.Sprintf("%s/%d", ipAddress, mask)
		// Используем PostgreSQL оператор << для проверки вхождения в сеть с приведением типов к inet
		ipClause = "(src_ip::inet << $" + strconv.Itoa(argIndex) + "::inet OR dst_ip::inet << $" + strconv.Itoa(argIndex) + "::inet)"
		args = append(args, subnet)
		argIndex++
	} else if strings.Contains(ipAddress, "/") {
		// CIDR notation (e.g., 192.168.1.0/24)
		_, ipNet, err := net.ParseCIDR(ipAddress)
		if err != nil {
			http.Error(w, "Invalid CIDR notation: "+err.Error(), http.StatusBadRequest)
			return
		}
		subnet = ipNet.String()
		// Используем PostgreSQL оператор << для проверки вхождения в сеть с приведением типов к inet
		ipClause = "(src_ip::inet << $" + strconv.Itoa(argIndex) + "::inet OR dst_ip::inet << $" + strconv.Itoa(argIndex) + "::inet)"
		args = append(args, subnet)
		argIndex++
	} else if strings.Contains(ipAddress, "*") {
		// Wildcard notation (e.g., 192.168.1.*)
		likePattern := strings.ReplaceAll(ipAddress, "*", "%")
		ipClause = "(src_ip LIKE $" + strconv.Itoa(argIndex) + " OR dst_ip LIKE $" + strconv.Itoa(argIndex) + ")"
		args = append(args, likePattern)
		argIndex++
	} else {
		// Ищем подключения с данным IP и используем их подсети для поиска
		var connectionSubnets []string
		err := h.db.Select(&connectionSubnets, `
			SELECT DISTINCT ip_address || '/' || mask::text as subnet 
			FROM connections 
			WHERE ip_address = $1
		`, ipAddress)
		
		if err == nil && len(connectionSubnets) > 0 {
			// Если нашли подключения с таким IP, ищем по их подсетям
			var subnetClauses []string
			for _, subnet := range connectionSubnets {
				subnetClauses = append(subnetClauses, 
					"(src_ip::inet << $"+strconv.Itoa(argIndex)+"::inet OR dst_ip::inet << $"+strconv.Itoa(argIndex)+"::inet)")
				args = append(args, subnet)
				argIndex++
			}
			ipClause = "(" + strings.Join(subnetClauses, " OR ") + ")"
		} else {
			// Если подключений не найдено, используем точное совпадение IP
			ipClause = "(src_ip = $" + strconv.Itoa(argIndex) + " OR dst_ip = $" + strconv.Itoa(argIndex) + ")"
			args = append(args, ipAddress)
			argIndex++
		}
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

	// Определяем направление трафика и заполняем поля для flows на текущей странице
	var flows []FlowRecord
	var pageBytesIn, pageBytesOut int64
	
	for _, flow := range flowsFromDB {
		// Правильная логика определения направления как в ConnectionStats
		if flow.SrcIP == ipAddress {
			// Исходящий трафик - трафик ИЗ этого IP
			flow.Direction = "outgoing"
			flow.BytesIn = 0
			flow.BytesOut = flow.Bytes
			pageBytesOut += flow.Bytes
		} else if flow.DstIP == ipAddress {
			// Входящий трафик - трафик К этому IP
			flow.Direction = "incoming"
			flow.BytesIn = flow.Bytes
			flow.BytesOut = 0
			pageBytesIn += flow.Bytes
		} else {
			// Если используется подсеть/маска, определяем по IP подключения
			// В этом случае нужно проверить, что поиск шел именно по подсети
			if subnet != "" || strings.Contains(ipAddress, "/") || strings.Contains(ipAddress, "*") {
				// Для подсетевого поиска считаем трафик по факту IP адресов
				if flow.SrcIP == ipAddress {
					flow.Direction = "outgoing"
					flow.BytesIn = 0
					flow.BytesOut = flow.Bytes
					pageBytesOut += flow.Bytes
				} else {
					flow.Direction = "incoming" 
					flow.BytesIn = flow.Bytes
					flow.BytesOut = 0
					pageBytesIn += flow.Bytes
				}
			} else {
				// Fallback для точного поиска
				flow.Direction = "mixed"
				flow.BytesIn = flow.Bytes / 2
				flow.BytesOut = flow.Bytes / 2
				pageBytesIn += flow.Bytes / 2
				pageBytesOut += flow.Bytes / 2
			}
		}
		flows = append(flows, flow)
	}

	// Отдельный запрос для получения всего входящего, исходящего и общего трафика
	totalQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN dst_ip = $%d THEN bytes ELSE 0 END), 0) as total_bytes_in,
			COALESCE(SUM(CASE WHEN src_ip = $%d THEN bytes ELSE 0 END), 0) as total_bytes_out,
			COALESCE(SUM(bytes), 0) as total_traffic
		FROM flows
		%s
	`, argIndex, argIndex+1, whereClause)
	
	totalArgs := make([]interface{}, len(args))
	copy(totalArgs, args)
	totalArgs = append(totalArgs, ipAddress, ipAddress)
	var totalResult struct {
		TotalBytesIn  int64 `db:"total_bytes_in"`
		TotalBytesOut int64 `db:"total_bytes_out"`
		TotalTraffic  int64 `db:"total_traffic"`
	}
	
	err = h.db.Get(&totalResult, totalQuery, totalArgs...)
	if err != nil {
		log.Printf("Error getting total traffic stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	totalBytesIn := totalResult.TotalBytesIn
	totalBytesOut := totalResult.TotalBytesOut

	totalPages := (totalRecords + limit - 1) / limit
	
	result := FlowSearchResult{
		Flows:         flows,
		TotalRecords:  totalRecords,
		TotalBytesIn:  totalBytesIn,
		TotalBytesOut: totalBytesOut,
		TotalTraffic:  totalResult.TotalTraffic,
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
