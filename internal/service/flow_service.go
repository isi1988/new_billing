package service

import (
	"fmt"
	"log"
	"new-billing/internal/config"
	"new-billing/internal/parser"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type FlowService struct {
	db  *sqlx.DB
	cfg *config.NfcapdConfig
}

func NewFlowService(db *sqlx.DB, cfg *config.NfcapdConfig) *FlowService {
	return &FlowService{db: db, cfg: cfg}
}

func (s *FlowService) StartProcessing() {
	interval, err := time.ParseDuration(s.cfg.ScanInterval)
	if err != nil {
		log.Fatalf("Invalid scan interval: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		s.processFiles()
		<-ticker.C
	}
}

func (s *FlowService) processFiles() {
	log.Println("Scanning for new nfcapd files...")
	err := filepath.Walk(s.cfg.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(info.Name(), "nfcapd.") {
			isProcessed, err := s.isFileProcessed(info.Name())
			if err != nil {
				log.Printf("Error checking if file was processed: %v", err)
				return nil
			}

			if !isProcessed {
				log.Printf("Processing new file: %s", path)
				flows, err := parser.ParseNfcapdFile(path)
				if err != nil {
					log.Printf("Error parsing file %s: %v", path, err)
					return nil
				}

				if err := s.saveFlows(flows); err != nil {
					log.Printf("Error saving flows for file %s: %v", path, err)
					return nil
				}

				if err := s.markFileAsProcessed(info.Name()); err != nil {
					log.Printf("Error marking file %s as processed: %v", path, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking through directory: %v", err)
	}
}

func (s *FlowService) isFileProcessed(fileName string) (bool, error) {
	var exists bool
	err := s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM processed_files WHERE file_name = $1)", fileName)
	return exists, err
}

func (s *FlowService) markFileAsProcessed(fileName string) error {
	_, err := s.db.Exec("INSERT INTO processed_files (file_name) VALUES ($1)", fileName)
	return err
}

// saveFlows теперь выполняет агрегацию трафика перед записью в базу данных.
func (s *FlowService) saveFlows(flows []parser.FlowData) error {
	if len(flows) == 0 {
		return nil
	}

	log.Printf("Aggregating %d raw flows into 5-minute buckets...", len(flows))

	// --- Логика агрегации ---

	// Ключ для нашей карты агрегации. Уникально идентифицирует поток в пределах временного окна.
	type aggregationKey struct {
		Timestamp time.Time // Усеченное до 5 минут время
		SrcIP     string
		DstIP     string
		SrcPort   int
		DstPort   int
		Protocol  int
	}

	// Карта для хранения агрегированных данных.
	// Ключ - уникальный поток, Значение - накопленные данные (байты/пакеты).
	aggregator := make(map[aggregationKey]*parser.FlowData)

	// Проходим по всем сырым потокам, полученным из файла
	for _, flow := range flows {
		// 1. Усекаем временную метку до ближайших 5 минут вниз.
		//    Например, 10:08:32 станет 10:05:00
		truncatedTime := flow.Timestamp.Truncate(5 * time.Minute)

		// 2. Создаем ключ для текущего потока.
		key := aggregationKey{
			Timestamp: truncatedTime,
			SrcIP:     flow.SrcIP,
			DstIP:     flow.DstIP,
			SrcPort:   flow.SrcPort,
			DstPort:   flow.DstPort,
			Protocol:  flow.Protocol,
		}

		// 3. Проверяем, есть ли уже запись для такого потока в нашем временном окне.
		if existingFlow, found := aggregator[key]; found {
			// Если есть - просто добавляем байты и пакеты.
			existingFlow.Bytes += flow.Bytes
			existingFlow.Packets += flow.Packets
		} else {
			// Если нет - создаем новую запись в карте агрегации.
			// Важно установить усеченное время, чтобы все записи были выровнены.
			flow.Timestamp = truncatedTime
			aggregator[key] = &flow
		}
	}

	// --- Запись в базу данных ---

	// Теперь у нас гораздо меньше записей для вставки.
	log.Printf("Aggregation complete. Writing %d aggregated flows to the database.", len(aggregator))

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO flows (timestamp, src_ip, dst_ip, src_port, dst_port, protocol, packets, bytes)
							 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	// Проходим по агрегированным данным и вставляем их в базу.
	for _, flow := range aggregator {
		_, err := stmt.Exec(flow.Timestamp, flow.SrcIP, flow.DstIP, flow.SrcPort, flow.DstPort, flow.Protocol, flow.Packets, flow.Bytes)
		if err != nil {
			tx.Rollback() // В случае ошибки откатываем всю транзакцию
			return fmt.Errorf("error executing insert for aggregated flow: %w", err)
		}
	}

	return tx.Commit()
}
