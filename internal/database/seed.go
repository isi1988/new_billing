package database

import (
	"fmt"
	"log"
	"math/rand"
	"new-billing/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

func SeedTestData(db *sqlx.DB) {
	log.Println("Seeding test data...")

	// Проверяем, есть ли уже тестовые данные в таблице traffic
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM traffic")
	if err != nil {
		log.Printf("Error checking traffic data: %v", err)
		return
	}

	// Если данные уже есть, пропускаем заполнение
	if count > 0 {
		log.Println("Test data already exists, skipping seeding")
		return
	}

	// Получаем все подключения для генерации трафика
	var connections []struct {
		ID       int `db:"id"`
		ClientID int `db:"client_id"`
	}

	err = db.Select(&connections, `
		SELECT c.id, ct.client_id 
		FROM connections c 
		JOIN contracts ct ON c.contract_id = ct.id
	`)

	if err != nil {
		log.Printf("Error getting connections: %v", err)
		return
	}

	if len(connections) == 0 {
		log.Println("No connections found, cannot seed traffic data")
		return
	}

	// Генерируем тестовые данные трафика за последние 30 дней
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	log.Printf("Generating traffic data for %d connections...", len(connections))

	// Для каждого подключения генерируем случайные записи трафика
	for _, conn := range connections {
		// Генерируем от 50 до 200 записей за последние 30 дней
		recordCount := rand.Intn(150) + 50

		for i := 0; i < recordCount; i++ {
			// Случайное время за последние 30 дней
			hoursBack := rand.Intn(30 * 24) // 30 дней * 24 часа
			timestamp := now.Add(-time.Duration(hoursBack) * time.Hour)

			// Добавляем случайные минуты и секунды для разнообразия
			minutesOffset := rand.Intn(60)
			secondsOffset := rand.Intn(60)
			timestamp = timestamp.Add(time.Duration(minutesOffset)*time.Minute + time.Duration(secondsOffset)*time.Second)

			// Генерируем реалистичные значения трафика
			// Входящий трафик: от 100KB до 500MB
			bytesIn := int64(rand.Intn(500*1024*1024-100*1024) + 100*1024)
			// Исходящий трафик: обычно меньше входящего
			bytesOut := int64(rand.Intn(int(bytesIn/2)) + 1024)

			// Пакеты пропорционально трафику
			packetsIn := bytesIn / int64(rand.Intn(1400)+500) // средний размер пакета 500-1900 байт
			packetsOut := bytesOut / int64(rand.Intn(1400)+500)

			// Вставляем запись
			_, err := db.Exec(`
				INSERT INTO traffic (connection_id, client_id, timestamp, bytes_in, bytes_out, packets_in, packets_out)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, conn.ID, conn.ClientID, timestamp, bytesIn, bytesOut, packetsIn, packetsOut)

			if err != nil {
				log.Printf("Error inserting traffic data: %v", err)
				continue
			}
		}
	}

	// Подсчитываем сколько записей создали
	err = db.Get(&count, "SELECT COUNT(*) FROM traffic")
	if err == nil {
		log.Printf("Successfully seeded %d traffic records", count)
	}
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
		signDate := time.Now().AddDate(0, -rand.Intn(12), 0) // Договор подписан в течение последнего года

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

	// Создаем подключения
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
