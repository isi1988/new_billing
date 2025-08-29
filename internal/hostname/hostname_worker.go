package hostname

import (
	"database/sql"
	"log"
	"net"
	"time"
	"new-billing/internal/models"

	"github.com/jmoiron/sqlx"
)

type HostnameWorker struct {
	db *sqlx.DB
}

func NewHostnameWorker(db *sqlx.DB) *HostnameWorker {
	return &HostnameWorker{
		db: db,
	}
}

// StartWorker запускает воркер для резолвинга IP адресов в хосты
func (hw *HostnameWorker) StartWorker() {
	log.Println("Starting hostname resolution worker...")
	
	// Сразу выполняем первичную обработку
	hw.processIPs()
	
	// Затем запускаем периодическую обработку каждые 5 дней
	ticker := time.NewTicker(5 * 24 * time.Hour) // 5 дней
	go func() {
		for {
			select {
			case <-ticker.C:
				hw.processIPs()
			}
		}
	}()
}

// processIPs получает все уникальные IP адреса из flows и разрешает их в хосты
func (hw *HostnameWorker) processIPs() {
	log.Println("Processing IP addresses for hostname resolution...")
	
	// Получаем все уникальные IP адреса из flows
	uniqueIPs, err := hw.getUniqueIPs()
	if err != nil {
		log.Printf("Error getting unique IPs: %v", err)
		return
	}
	
	log.Printf("Found %d unique IP addresses to process", len(uniqueIPs))
	
	processed := 0
	skipped := 0
	
	for _, ip := range uniqueIPs {
		// Проверяем, нужно ли обновить информацию об этом IP
		shouldUpdate, err := hw.shouldUpdateIP(ip)
		if err != nil {
			log.Printf("Error checking if should update IP %s: %v", ip, err)
			continue
		}
		
		if !shouldUpdate {
			skipped++
			continue
		}
		
		// Разрешаем IP в хост
		hostname, err := hw.resolveIP(ip)
		if err != nil {
			// Если не удалось разрешить, используем сам IP как хост
			hostname = ip
		}
		
		// Сохраняем результат в базу
		err = hw.saveHostname(ip, hostname)
		if err != nil {
			log.Printf("Error saving hostname for IP %s: %v", ip, err)
			continue
		}
		
		processed++
		
		// Небольшая задержка, чтобы не перегружать DNS сервер
		time.Sleep(100 * time.Millisecond)
	}
	
	log.Printf("Hostname resolution completed: processed %d, skipped %d IPs", processed, skipped)
}

// getUniqueIPs получает все уникальные IP адреса из таблицы flows
func (hw *HostnameWorker) getUniqueIPs() ([]string, error) {
	query := `
		SELECT DISTINCT src_ip FROM flows
		UNION
		SELECT DISTINCT dst_ip FROM flows
	`
	
	var ips []string
	err := hw.db.Select(&ips, query)
	return ips, err
}

// shouldUpdateIP проверяет, нужно ли обновить информацию об IP адресе
func (hw *HostnameWorker) shouldUpdateIP(ip string) (bool, error) {
	var updatedAt *time.Time
	err := hw.db.Get(&updatedAt, 
		"SELECT updated_at FROM ip_hostnames WHERE ip_address = $1", ip)
	
	if err == sql.ErrNoRows {
		// IP не найден в базе, нужно добавить
		return true, nil
	}
	if err != nil {
		return false, err
	}
	
	// Если прошло больше 5 дней с последнего обновления, обновляем
	if updatedAt != nil {
		return time.Since(*updatedAt) > 5*24*time.Hour, nil
	}
	
	return true, nil
}

// resolveIP разрешает IP адрес в имя хоста
func (hw *HostnameWorker) resolveIP(ip string) (string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}
	
	if len(names) > 0 {
		return names[0], nil
	}
	
	return ip, nil
}

// saveHostname сохраняет соответствие IP -> hostname в базу данных
func (hw *HostnameWorker) saveHostname(ip, hostname string) error {
	now := time.Now()
	
	// Пытаемся обновить существующую запись
	result, err := hw.db.Exec(`
		UPDATE ip_hostnames 
		SET hostname = $2, updated_at = $3 
		WHERE ip_address = $1
	`, ip, hostname, now)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	// Если строка не была обновлена, создаем новую
	if rowsAffected == 0 {
		_, err = hw.db.Exec(`
			INSERT INTO ip_hostnames (ip_address, hostname, resolved_at, updated_at)
			VALUES ($1, $2, $3, $3)
		`, ip, hostname, now)
	}
	
	return err
}

// GetIPInfo получает информацию об IP адресе включая hostname
func (hw *HostnameWorker) GetIPInfo(ip string) (*models.IPHostname, error) {
	var ipInfo models.IPHostname
	err := hw.db.Get(&ipInfo, 
		"SELECT * FROM ip_hostnames WHERE ip_address = $1", ip)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return &ipInfo, err
}

// GetConnectionInfo получает информацию о подключении по IP адресу
func (hw *HostnameWorker) GetConnectionInfo(ip string) (map[string]interface{}, error) {
	info := make(map[string]interface{})
	
	// Получаем информацию о подключении, договоре и клиенте
	query := `
		SELECT 
			c.id as connection_id,
			c.address,
			c.connection_type,
			co.number as contract_number,
			co.id as contract_id,
			cl.id as client_id,
			CASE 
				WHEN cl.client_type = 'individual' THEN 
					COALESCE(cl.last_name || ' ' || cl.first_name, cl.email, 'Клиент ID: ' || cl.id::text)
				ELSE 
					COALESCE(cl.short_name, cl.full_name, cl.email, 'Клиент ID: ' || cl.id::text)
			END as client_name,
			cl.client_type,
			t.name as tariff_name
		FROM connections c
		JOIN contracts co ON c.contract_id = co.id
		JOIN clients cl ON co.client_id = cl.id
		JOIN tariffs t ON c.tariff_id = t.id
		WHERE c.ip_address = $1
	`
	
	type ConnectionInfo struct {
		ConnectionID   int    `db:"connection_id"`
		Address        string `db:"address"`
		ConnectionType string `db:"connection_type"`
		ContractNumber string `db:"contract_number"`
		ContractID     int    `db:"contract_id"`
		ClientID       int    `db:"client_id"`
		ClientName     string `db:"client_name"`
		ClientType     string `db:"client_type"`
		TariffName     string `db:"tariff_name"`
	}
	
	var connInfo ConnectionInfo
	err := hw.db.Get(&connInfo, query, ip)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	info["connection"] = connInfo
	return info, nil
}