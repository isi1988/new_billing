package database

import (
	"fmt"
	"log"
	"new-billing/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

// ProcessIPTraffic обрабатывает трафик по IP и связывает с клиентом если возможно
func ProcessIPTraffic(db *sqlx.DB, srcIP, dstIP string, bytes, packets int64, timestamp time.Time) error {
	// Ищем подключение по IP (источник или назначение)
	var connection struct {
		ID        int    `db:"id"`
		ClientID  int    `db:"client_id"`
		IPAddress string `db:"ip_address"`
	}

	// Сначала проверяем, является ли источник нашим клиентом
	err := db.Get(&connection, `
		SELECT c.id, ct.client_id, c.ip_address
		FROM connections c 
		JOIN contracts ct ON c.contract_id = ct.id
		WHERE c.ip_address = $1
	`, srcIP)

	if err != nil {
		// Если источник не найден, проверяем назначение
		err = db.Get(&connection, `
			SELECT c.id, ct.client_id, c.ip_address
			FROM connections c 
			JOIN contracts ct ON c.contract_id = ct.id
			WHERE c.ip_address = $1
		`, dstIP)

		if err != nil {
			// Если ни источник, ни назначение не являются нашими клиентами, игнорируем
			return nil
		}
	}

	// Определяем направление трафика
	var bytesIn, bytesOut, packetsIn, packetsOut int64
	if connection.IPAddress == dstIP {
		// Трафик идет К нашему клиенту (входящий)
		bytesIn = bytes
		packetsIn = packets
		bytesOut = 0
		packetsOut = 0
	} else {
		// Трафик идет ОТ нашего клиента (исходящий)
		bytesOut = bytes
		packetsOut = packets
		bytesIn = 0
		packetsIn = 0
	}

	// Записываем данные трафика
	_, err = db.Exec(`
		INSERT INTO traffic (connection_id, client_id, timestamp, bytes_in, bytes_out, packets_in, packets_out)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, connection.ID, connection.ClientID, timestamp, bytesIn, bytesOut, packetsIn, packetsOut)

	return err
}

func SeedBasicData(db *sqlx.DB) {
	log.Println("Seeding basic data if needed...")

	// Проверяем наличие базовых данных
	var clientCount int
	err := db.Get(&clientCount, "SELECT COUNT(*) FROM clients")
	if err != nil || clientCount > 0 {
		return // Данные уже есть или ошибка
	}

	log.Println("Creating basic test data...")

	// Создаем тестовых клиентов
	testClients := []struct {
		Type      models.ClientType
		Email     string
		Phone     string
		FirstName *string
		LastName  *string
		FullName  *string
	}{
		{
			Type:      models.IndividualType,
			Email:     "ivan.petrov@example.com",
			Phone:     "+7 (900) 123-45-67",
			FirstName: stringPtr("Иван"),
			LastName:  stringPtr("Петров"),
		},
		{
			Type:      models.IndividualType,
			Email:     "maria.sidorova@example.com",
			Phone:     "+7 (900) 234-56-78",
			FirstName: stringPtr("Мария"),
			LastName:  stringPtr("Сидорова"),
		},
		{
			Type:     models.LegalEntityType,
			Email:    "info@techcorp.ru",
			Phone:    "+7 (495) 123-45-67",
			FullName: stringPtr("ООО \"ТехКорп\""),
		},
	}

	clientIDs := make([]int, 0, len(testClients))
	for _, client := range testClients {
		var id int
		err := db.QueryRow(`
			INSERT INTO clients (client_type, email, phone, is_blocked, first_name, last_name, full_name)
			VALUES ($1, $2, $3, false, $4, $5, $6) RETURNING id
		`, client.Type, client.Email, client.Phone, client.FirstName, client.LastName, client.FullName).Scan(&id)

		if err != nil {
			log.Printf("Error creating client: %v", err)
			continue
		}
		clientIDs = append(clientIDs, id)
	}

	// Создаем тестовые тарифы
	testTariffs := []struct {
		Name             string
		PaymentType      models.PaymentType
		IsForIndividuals bool
		MaxSpeedIn       int
		MaxSpeedOut      int
		MaxTrafficIn     int64
		MaxTrafficOut    int64
	}{
		{
			Name:             "Домашний 100",
			PaymentType:      models.Postpaid,
			IsForIndividuals: true,
			MaxSpeedIn:       100000, // 100 Мбит/с
			MaxSpeedOut:      100000,
			MaxTrafficIn:     1000000000000, // 1 ТБ
			MaxTrafficOut:    1000000000000,
		},
		{
			Name:             "Бизнес 1000",
			PaymentType:      models.Postpaid,
			IsForIndividuals: false,
			MaxSpeedIn:       1000000, // 1 Гбит/с
			MaxSpeedOut:      1000000,
			MaxTrafficIn:     10000000000000, // 10 ТБ
			MaxTrafficOut:    10000000000000,
		},
	}

	tariffIDs := make([]int, 0, len(testTariffs))
	for _, tariff := range testTariffs {
		var id int
		err := db.QueryRow(`
			INSERT INTO tariffs (name, is_archived, payment_type, is_for_individuals, max_speed_in, max_speed_out, max_traffic_in, max_traffic_out)
			VALUES ($1, false, $2, $3, $4, $5, $6, $7) RETURNING id
		`, tariff.Name, tariff.PaymentType, tariff.IsForIndividuals, tariff.MaxSpeedIn, tariff.MaxSpeedOut, tariff.MaxTrafficIn, tariff.MaxTrafficOut).Scan(&id)

		if err != nil {
			log.Printf("Error creating tariff: %v", err)
			continue
		}
		tariffIDs = append(tariffIDs, id)
	}

	// Создаем тестовое оборудование
	testEquipment := []struct {
		Model       string
		Description string
		MACAddress  string
	}{
		{
			Model:       "Cisco ISR4321",
			Description: "Маршрутизатор для малого офиса",
			MACAddress:  "00:1B:44:11:3A:B7",
		},
		{
			Model:       "TP-Link Archer C7",
			Description: "WiFi роутер для дома",
			MACAddress:  "00:1B:44:11:3A:B8",
		},
		{
			Model:       "Mikrotik hEX S",
			Description: "Профессиональный маршрутизатор",
			MACAddress:  "00:1B:44:11:3A:B9",
		},
	}

	equipmentIDs := make([]int, 0, len(testEquipment))
	for _, equip := range testEquipment {
		var id int
		err := db.QueryRow(`
			INSERT INTO equipment (model, description, mac_address)
			VALUES ($1, $2, $3) RETURNING id
		`, equip.Model, equip.Description, equip.MACAddress).Scan(&id)

		if err != nil {
			log.Printf("Error creating equipment: %v", err)
			continue
		}
		equipmentIDs = append(equipmentIDs, id)
	}

	// Создаем договоры
	contractIDs := make([]int, 0, len(clientIDs))
	for i, clientID := range clientIDs {
		contractNumber := generateContractNumber(i + 1)
		signDate := time.Now().AddDate(0, -1, 0) // Договор подписан месяц назад

		var id int
		err := db.QueryRow(`
			INSERT INTO contracts (client_id, "number", sign_date, is_blocked)
			VALUES ($1, $2, $3, false) RETURNING id
		`, clientID, contractNumber, signDate).Scan(&id)

		if err != nil {
			log.Printf("Error creating contract: %v", err)
			continue
		}
		contractIDs = append(contractIDs, id)
	}

	// Создаем подключения с реальными IP адресами
	for i, contractID := range contractIDs {
		if i >= len(equipmentIDs) || i >= len(tariffIDs) {
			break
		}

		ipAddress := generateIPAddress(i + 1)

		_, err := db.Exec(`
			INSERT INTO connections (equipment_id, contract_id, address, connection_type, tariff_id, ip_address, mask)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, equipmentIDs[i%len(equipmentIDs)], contractID, generateAddress(i+1), "ethernet", tariffIDs[i%len(tariffIDs)], ipAddress, 24)

		if err != nil {
			log.Printf("Error creating connection: %v", err)
			continue
		}
	}

	log.Println("Basic test data seeded successfully")
	log.Println("IP addresses assigned:")
	log.Println("- 192.168.1.101 -> Иван Петров")
	log.Println("- 192.168.1.102 -> Мария Сидорова")
	log.Println("- 192.168.1.103 -> ООО \"ТехКорп\"")
}

func stringPtr(s string) *string {
	return &s
}

func generateContractNumber(index int) string {
	return "DOG-" + time.Now().Format("2006") + "-" + fmt.Sprintf("%04d", index)
}

func generateIPAddress(index int) string {
	return "192.168.1." + fmt.Sprintf("%d", 100+index)
}

func generateAddress(index int) string {
	addresses := []string{
		"г. Москва, ул. Ленина, д. 1",
		"г. Москва, ул. Пушкина, д. 10, кв. 5",
		"г. Москва, пр. Мира, д. 25, офис 301",
	}
	return addresses[index%len(addresses)]
}
