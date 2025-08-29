package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"new-billing/internal/models"
	"new-billing/internal/telegram"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type BillingHandler struct {
	DB              *sqlx.DB
	TelegramService *telegram.TelegramService
}

// --- Хелпер для получения количества измененных строк ---
func mustRowsAffected(res sql.Result) int64 {
	count, err := res.RowsAffected()
	if err != nil {
		return 0
	}
	return count
}

//================================================================================
// CRUD: USERS (Пользователи)
//================================================================================

// @Summary      Создать пользователя
// @Description  Создает нового пользователя (доступно только администраторам)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user body object{username=string,password=string,role=string} true "Данные нового пользователя"
// @Success      201  {object}  models.User
// @Failure      400  {object}  map[string]string "Invalid request"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /users [post]
// @Security     BearerAuth
func (h *BillingHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string      `json:"username"`
		Password string      `json:"password"`
		Role     models.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Password == "" {
		http.Error(w, "Invalid request body or empty password", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	var newUser models.User
	query := `INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) RETURNING id, username, role, created_at`
	err = h.DB.QueryRowx(query, payload.Username, string(hashedPassword), payload.Role).StructScan(&newUser)
	if err != nil {
		http.Error(w, "Could not create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

// @Summary      Получить список пользователей
// @Description  Возвращает список всех пользователей
// @Tags         Users
// @Produce      json
// @Success      200  {array}   models.User
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /users [get]
// @Security     BearerAuth
func (h *BillingHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users := []models.User{}
	if err := h.DB.Select(&users, "SELECT id, username, role, created_at FROM users ORDER BY id"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// @Summary      Получить пользователя по ID
// @Description  Возвращает одного пользователя по его ID
// @Tags         Users
// @Produce      json
// @Param        id   path      int  true  "ID Пользователя"
// @Success      200  {object}  models.User
// @Failure      404  {object}  map[string]string "User not found"
// @Router       /users/{id} [get]
// @Security     BearerAuth
func (h *BillingHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	user := models.User{}
	if err := h.DB.Get(&user, "SELECT id, username, role, created_at FROM users WHERE id=$1", id); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary      Обновить пользователя
// @Description  Обновляет данные пользователя (кроме пароля)
// @Tags         Users
// @Accept       json
// @Param        id   path      int          true  "ID Пользователя"
// @Param        user body      models.User  true  "Обновленные данные"
// @Success      200  {string}  string "OK"
// @Failure      404  {object}  map[string]string "User not found"
// @Router       /users/{id} [put]
// @Security     BearerAuth
func (h *BillingHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE users SET username=$1, role=$2 WHERE id=$3`
	res, err := h.DB.Exec(query, user.Username, user.Role, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "User not found or not updated", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary      Удалить пользователя
// @Description  Удаляет пользователя по ID (нельзя удалить 'admin')
// @Tags         Users
// @Param        id   path      int  true  "ID Пользователя"
// @Success      204  {string}  string "No Content"
// @Failure      404  {object}  map[string]string "User not found"
// @Router       /users/{id} [delete]
// @Security     BearerAuth
func (h *BillingHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("DELETE FROM users WHERE id=$1 AND username != 'admin'", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "User not found or cannot be deleted", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//================================================================================
// CRUD: CLIENTS (Клиенты)
//================================================================================

// @Summary      Создать клиента
// @Description  Создает нового клиента (физ. или юр. лицо)
// @Tags         Clients
// @Accept       json
// @Produce      json
// @Param        client body models.Client true "Объект нового клиента"
// @Success      201  {object}  models.Client
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /clients [post]
// @Security     BearerAuth
func (h *BillingHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var client models.Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO clients (client_type, email, phone, is_blocked, first_name, last_name, patronymic, passport_number, passport_issued_by, passport_issue_date, registration_address, birth_date, inn, kpp, full_name, short_name, ogrn, ogrn_date, legal_address, actual_address, bank_name, bank_account, bank_bik, bank_correspondent, ceo, accountant) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26) RETURNING id`
	err := h.DB.QueryRow(query, client.ClientType, client.Email, client.Phone, client.IsBlocked, client.FirstName, client.LastName, client.Patronymic, client.PassportNumber, client.PassportIssuedBy, client.PassportIssueDate, client.RegistrationAddress, client.BirthDate, client.INN, client.KPP, client.FullName, client.ShortName, client.OGRN, client.OGRNDate, client.LegalAddress, client.ActualAddress, client.BankName, client.BankAccount, client.BankBIK, client.BankCorrespondent, client.CEO, client.Accountant).Scan(&client.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

// @Summary      Получить список клиентов
// @Description  Возвращает список всех клиентов
// @Tags         Clients
// @Produce      json
// @Success      200  {array}   models.Client
// @Failure      500  {object}  map[string]string
// @Router       /clients [get]
// @Security     BearerAuth
func (h *BillingHandler) GetClients(w http.ResponseWriter, r *http.Request) {
	type ClientWithDetails struct {
		models.Client
		ContractsCount int `json:"contracts_count" db:"contracts_count"`
	}

	var clients []ClientWithDetails
	query := `
		SELECT 
			c.*,
			COALESCE(COUNT(ct.id), 0) as contracts_count
		FROM clients c
		LEFT JOIN contracts ct ON c.id = ct.client_id
		GROUP BY c.id, c.client_type, c.email, c.phone, c.is_blocked, 
			c.first_name, c.last_name, c.patronymic, c.passport_number, 
			c.passport_issued_by, c.passport_issue_date, c.registration_address, 
			c.birth_date, c.inn, c.kpp, c.full_name, c.short_name, c.ogrn, 
			c.ogrn_date, c.legal_address, c.actual_address, c.bank_name, 
			c.bank_account, c.bank_bik, c.bank_correspondent, c.ceo, c.accountant
		ORDER BY c.id
	`

	if err := h.DB.Select(&clients, query); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

// @Summary      Получить клиента по ID
// @Description  Возвращает одного клиента по его ID
// @Tags         Clients
// @Produce      json
// @Param        id   path      int  true  "ID Клиента"
// @Success      200  {object}  models.Client
// @Failure      404  {object}  map[string]string
// @Router       /clients/{id} [get]
// @Security     BearerAuth
func (h *BillingHandler) GetClientByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	client := models.Client{}
	if err := h.DB.Get(&client, "SELECT * FROM clients WHERE id=$1", id); err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

// @Summary      Обновить клиента
// @Description  Обновляет данные клиента по ID
// @Tags         Clients
// @Accept       json
// @Param        id     path      int            true  "ID Клиента"
// @Param        client body      models.Client  true  "Обновленные данные"
// @Success      200    {string}  string "OK"
// @Failure      404    {object}  map[string]string
// @Router       /clients/{id} [put]
// @Security     BearerAuth
func (h *BillingHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var client models.Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE clients SET client_type=$1, email=$2, phone=$3, is_blocked=$4, first_name=$5, last_name=$6, patronymic=$7, passport_number=$8, passport_issued_by=$9, passport_issue_date=$10, registration_address=$11, birth_date=$12, inn=$13, kpp=$14, full_name=$15, short_name=$16, ogrn=$17, ogrn_date=$18, legal_address=$19, actual_address=$20, bank_name=$21, bank_account=$22, bank_bik=$23, bank_correspondent=$24, ceo=$25, accountant=$26 WHERE id=$27`
	res, err := h.DB.Exec(query, client.ClientType, client.Email, client.Phone, client.IsBlocked, client.FirstName, client.LastName, client.Patronymic, client.PassportNumber, client.PassportIssuedBy, client.PassportIssueDate, client.RegistrationAddress, client.BirthDate, client.INN, client.KPP, client.FullName, client.ShortName, client.OGRN, client.OGRNDate, client.LegalAddress, client.ActualAddress, client.BankName, client.BankAccount, client.BankBIK, client.BankCorrespondent, client.CEO, client.Accountant, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Client not found or not updated", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary      Удалить клиента
// @Description  Удаляет клиента и все связанные с ним данные (договоры, подключения)
// @Tags         Clients
// @Param        id   path      int  true  "ID Клиента"
// @Success      204  {string}  string "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /clients/{id} [delete]
// @Security     BearerAuth
func (h *BillingHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("DELETE FROM clients WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//================================================================================
// CRUD: EQUIPMENT (Оборудование)
//================================================================================

// @Summary      Добавить оборудование
// @Description  Создает новую запись об оборудовании
// @Tags         Equipment
// @Accept       json
// @Produce      json
// @Param        equipment body models.Equipment true "Объект нового оборудования"
// @Success      201  {object}  models.Equipment
// @Failure      500  {object}  map[string]string
// @Router       /equipment [post]
// @Security     BearerAuth
func (h *BillingHandler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	var equip models.Equipment
	if err := json.NewDecoder(r.Body).Decode(&equip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO equipment (model, description, mac_address) VALUES ($1, $2, $3) RETURNING id`
	err := h.DB.QueryRow(query, equip.Model, equip.Description, equip.MACAddress).Scan(&equip.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(equip)
}

// @Summary      Получить список оборудования
// @Description  Возвращает все оборудование
// @Tags         Equipment
// @Produce      json
// @Success      200  {array}   models.Equipment
// @Failure      500  {object}  map[string]string
// @Router       /equipment [get]
// @Security     BearerAuth
func (h *BillingHandler) GetAllEquipment(w http.ResponseWriter, r *http.Request) {
	equipment := []models.Equipment{}
	if err := h.DB.Select(&equipment, "SELECT * FROM equipment ORDER BY id"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(equipment)
}

// @Summary      Получить оборудование по ID
// @Description  Возвращает одну единицу оборудования по ID
// @Tags         Equipment
// @Produce      json
// @Param        id   path      int  true  "ID Оборудования"
// @Success      200  {object}  models.Equipment
// @Failure      404  {object}  map[string]string
// @Router       /equipment/{id} [get]
// @Security     BearerAuth
func (h *BillingHandler) GetEquipmentByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	equip := models.Equipment{}
	if err := h.DB.Get(&equip, "SELECT * FROM equipment WHERE id=$1", id); err != nil {
		http.Error(w, "Equipment not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(equip)
}

// @Summary      Обновить оборудование
// @Description  Обновляет данные оборудования по ID
// @Tags         Equipment
// @Accept       json
// @Param        id        path      int              true  "ID Оборудования"
// @Param        equipment body      models.Equipment true  "Обновленные данные"
// @Success      200       {string}  string "OK"
// @Failure      404       {object}  map[string]string
// @Router       /equipment/{id} [put]
// @Security     BearerAuth
func (h *BillingHandler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var equip models.Equipment
	if err := json.NewDecoder(r.Body).Decode(&equip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE equipment SET model=$1, description=$2, mac_address=$3 WHERE id=$4`
	res, err := h.DB.Exec(query, equip.Model, equip.Description, equip.MACAddress, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Equipment not found or not updated", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary      Удалить оборудование
// @Description  Удаляет оборудование по ID
// @Tags         Equipment
// @Param        id   path      int  true  "ID Оборудования"
// @Success      204  {string}  string "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /equipment/{id} [delete]
// @Security     BearerAuth
func (h *BillingHandler) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("DELETE FROM equipment WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Equipment not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//================================================================================
// CRUD: TARIFFS (Тарифы)
//================================================================================

// @Summary      Создать тариф
// @Description  Создает новый тариф в системе
// @Tags         Tariffs
// @Accept       json
// @Produce      json
// @Param        tariff body models.Tariff true "Объект нового тарифа"
// @Success      201  {object}  models.Tariff
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /tariffs [post]
// @Security     BearerAuth
func (h *BillingHandler) CreateTariff(w http.ResponseWriter, r *http.Request) {
	var tariff models.Tariff
	if err := json.NewDecoder(r.Body).Decode(&tariff); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO tariffs (name, is_archived, payment_type, is_for_individuals, max_speed_in, max_speed_out, max_traffic_in, max_traffic_out) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := h.DB.QueryRow(query, tariff.Name, tariff.IsArchived, tariff.PaymentType, tariff.IsForIndividuals, tariff.MaxSpeedIn, tariff.MaxSpeedOut, tariff.MaxTrafficIn, tariff.MaxTrafficOut).Scan(&tariff.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tariff)
}

// @Summary      Получить список тарифов
// @Description  Возвращает список всех тарифов
// @Tags         Tariffs
// @Produce      json
// @Success      200  {array}  models.Tariff
// @Failure      500  {object}  map[string]string
// @Router       /tariffs [get]
// @Security     BearerAuth
func (h *BillingHandler) GetTariffs(w http.ResponseWriter, r *http.Request) {
	tariffs := []models.Tariff{}
	if err := h.DB.Select(&tariffs, "SELECT * FROM tariffs ORDER BY id"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tariffs)
}

// @Summary      Получить тариф по ID
// @Description  Возвращает один тариф по его уникальному идентификатору
// @Tags         Tariffs
// @Produce      json
// @Param        id   path      int  true  "ID Тарифа"
// @Success      200  {object}  models.Tariff
// @Failure      404  {object}  map[string]string
// @Router       /tariffs/{id} [get]
// @Security     BearerAuth
func (h *BillingHandler) GetTariffByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	tariff := models.Tariff{}
	if err := h.DB.Get(&tariff, "SELECT * FROM tariffs WHERE id=$1", id); err != nil {
		http.Error(w, "Tariff not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tariff)
}

// @Summary      Обновить тариф
// @Description  Обновляет существующий тариф по его ID
// @Tags         Tariffs
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID Тарифа для обновления"
// @Param        tariff body models.Tariff true "Объект с обновленными данными тарифа"
// @Success      200  {string}  string "OK"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /tariffs/{id} [put]
// @Security     BearerAuth
func (h *BillingHandler) UpdateTariff(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var tariff models.Tariff
	if err := json.NewDecoder(r.Body).Decode(&tariff); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE tariffs SET name=$1, is_archived=$2, payment_type=$3, is_for_individuals=$4, max_speed_in=$5, max_speed_out=$6, max_traffic_in=$7, max_traffic_out=$8 WHERE id=$9`
	res, err := h.DB.Exec(query, tariff.Name, tariff.IsArchived, tariff.PaymentType, tariff.IsForIndividuals, tariff.MaxSpeedIn, tariff.MaxSpeedOut, tariff.MaxTrafficIn, tariff.MaxTrafficOut, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Tariff not found or not updated", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary      Удалить тариф
// @Description  Удаляет тариф по его ID
// @Tags         Tariffs
// @Param        id   path      int  true  "ID Тарифа для удаления"
// @Success      204  {string}  string "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /tariffs/{id} [delete]
// @Security     BearerAuth
func (h *BillingHandler) DeleteTariff(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("DELETE FROM tariffs WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Tariff not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//================================================================================
// CRUD: CONTRACTS (Договоры)
//================================================================================

// @Summary      Создать договор
// @Description  Создает новый договор для клиента
// @Tags         Contracts
// @Accept       json
// @Produce      json
// @Param        contract body models.Contract true "Объект нового договора"
// @Success      201  {object}  models.Contract
// @Failure      500  {object}  map[string]string
// @Router       /contracts [post]
// @Security     BearerAuth
func (h *BillingHandler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var contract models.Contract
	if err := json.NewDecoder(r.Body).Decode(&contract); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO contracts (client_id, "number", sign_date) VALUES ($1, $2, $3) RETURNING id`
	err := h.DB.QueryRow(query, contract.ClientID, contract.Number, contract.SignDate).Scan(&contract.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return the full contract details including client information
	type ContractWithDetails struct {
		ID               int               `json:"id" db:"id"`
		Number           string            `json:"number" db:"number"`
		SignDate         models.CustomDate `json:"sign_date" db:"sign_date"`
		ClientID         int               `json:"client_id" db:"client_id"`
		IsBlocked        bool              `json:"is_blocked" db:"is_blocked"`
		ConnectionsCount int               `json:"connections_count" db:"connections_count"`
		ClientName       string            `json:"client_name" db:"client_name"`
		ClientEmail      string            `json:"client_email" db:"client_email"`
		ClientType       string            `json:"client_type" db:"client_type"`
	}
	
	var createdContract ContractWithDetails
	detailsQuery := `
		SELECT 
			c.id, c.number, c.sign_date, c.client_id, c.is_blocked,
			COUNT(DISTINCT conn.id) as connections_count,
			COALESCE(
				CASE 
					WHEN cl.client_type = 'individual' THEN 
						COALESCE(cl.last_name || ' ' || cl.first_name, cl.email)
					ELSE 
						COALESCE(cl.short_name, cl.full_name, cl.email)
				END,
				'Клиент #' || cl.id
			) as client_name,
			COALESCE(cl.email, '') as client_email,
			cl.client_type
		FROM contracts c
		LEFT JOIN connections conn ON c.id = conn.contract_id
		LEFT JOIN clients cl ON c.client_id = cl.id
		WHERE c.id = $1
		GROUP BY c.id, c.number, c.sign_date, c.client_id, c.is_blocked, 
				 cl.id, cl.client_type, cl.last_name, cl.first_name, cl.email, 
				 cl.short_name, cl.full_name
	`
	
	err = h.DB.QueryRow(detailsQuery, contract.ID).Scan(
		&createdContract.ID,
		&createdContract.Number,
		&createdContract.SignDate,
		&createdContract.ClientID,
		&createdContract.IsBlocked,
		&createdContract.ConnectionsCount,
		&createdContract.ClientName,
		&createdContract.ClientEmail,
		&createdContract.ClientType,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdContract)
}

// @Summary      Получить список договоров
// @Description  Возвращает все договоры
// @Tags         Contracts
// @Produce      json
// @Success      200  {array}   models.Contract
// @Failure      500  {object}  map[string]string
// @Router       /contracts [get]
// @Security     BearerAuth
func (h *BillingHandler) GetContracts(w http.ResponseWriter, r *http.Request) {
	type ContractWithDetails struct {
		ID               int               `json:"id" db:"id"`
		Number           string            `json:"number" db:"number"`
		SignDate         models.CustomDate `json:"sign_date" db:"sign_date"`
		ClientID         int               `json:"client_id" db:"client_id"`
		IsBlocked        bool              `json:"is_blocked" db:"is_blocked"`
		ConnectionsCount int               `json:"connections_count" db:"connections_count"`
		ClientName       string            `json:"client_name" db:"client_name"`
		ClientEmail      string            `json:"client_email" db:"client_email"`
		ClientType       string            `json:"client_type" db:"client_type"`
	}

	var contractsWithDetails []ContractWithDetails

	query := `
		SELECT 
			c.id, c.number, c.sign_date, c.client_id, c.is_blocked,
			COUNT(DISTINCT conn.id) as connections_count,
			COALESCE(
				CASE 
					WHEN cl.client_type = 'individual' THEN 
						COALESCE(cl.last_name || ' ' || cl.first_name, cl.email)
					ELSE 
						COALESCE(cl.short_name, cl.full_name, cl.email)
				END,
				'Клиент #' || cl.id
			) as client_name,
			COALESCE(cl.email, '') as client_email,
			cl.client_type
		FROM contracts c
		LEFT JOIN connections conn ON c.id = conn.contract_id
		LEFT JOIN clients cl ON c.client_id = cl.id
		GROUP BY c.id, c.number, c.sign_date, c.client_id, c.is_blocked, 
				 cl.id, cl.client_type, cl.last_name, cl.first_name, cl.email, 
				 cl.short_name, cl.full_name
		ORDER BY c.id
	`

	if err := h.DB.Select(&contractsWithDetails, query); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contractsWithDetails)
}

// @Summary      Получить договор по ID
// @Description  Возвращает один договор по ID
// @Tags         Contracts
// @Produce      json
// @Param        id   path      int  true  "ID Договора"
// @Success      200  {object}  models.Contract
// @Failure      404  {object}  map[string]string
// @Router       /contracts/{id} [get]
// @Security     BearerAuth
func (h *BillingHandler) GetContractByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	contract := models.Contract{}
	if err := h.DB.Get(&contract, `SELECT * FROM contracts WHERE id=$1`, id); err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contract)
}

// @Summary      Обновить договор
// @Description  Обновляет данные договора по ID
// @Tags         Contracts
// @Accept       json
// @Param        id       path      int             true  "ID Договора"
// @Param        contract body      models.Contract true  "Обновленные данные"
// @Success      200      {string}  string "OK"
// @Failure      404      {object}  map[string]string
// @Router       /contracts/{id} [put]
// @Security     BearerAuth
func (h *BillingHandler) UpdateContract(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var contract models.Contract
	if err := json.NewDecoder(r.Body).Decode(&contract); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE contracts SET client_id=$1, "number"=$2, sign_date=$3 WHERE id=$4`
	res, err := h.DB.Exec(query, contract.ClientID, contract.Number, contract.SignDate, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Contract not found or not updated", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary      Удалить договор
// @Description  Удаляет договор по ID
// @Tags         Contracts
// @Param        id   path      int  true  "ID Договора"
// @Success      204  {string}  string "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /contracts/{id} [delete]
// @Security     BearerAuth
func (h *BillingHandler) DeleteContract(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("DELETE FROM contracts WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//================================================================================
// CRUD: CONNECTIONS (Подключения)
//================================================================================

// @Summary      Создать подключение
// @Description  Создает новую запись о подключении услуги
// @Tags         Connections
// @Accept       json
// @Produce      json
// @Param        connection body models.Connection true "Объект нового подключения"
// @Success      201  {object}  models.Connection
// @Failure      500  {object}  map[string]string
// @Router       /connections [post]
// @Security     BearerAuth
func (h *BillingHandler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	var conn models.Connection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO connections (equipment_id, contract_id, address, connection_type, tariff_id, ip_address, mask, is_blocked) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := h.DB.QueryRow(query, conn.EquipmentID, conn.ContractID, conn.Address, conn.ConnectionType, conn.TariffID, conn.IPAddress, conn.Mask, conn.IsBlocked).Scan(&conn.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conn)
}

// @Summary      Получить список подключений
// @Description  Возвращает все подключения
// @Tags         Connections
// @Produce      json
// @Success      200  {array}   models.Connection
// @Failure      500  {object}  map[string]string
// @Router       /connections [get]
// @Security     BearerAuth
func (h *BillingHandler) GetConnections(w http.ResponseWriter, r *http.Request) {
	type ConnectionWithDetails struct {
		ID             int    `json:"id" db:"id"`
		EquipmentID    int    `json:"equipment_id" db:"equipment_id"`
		ContractID     int    `json:"contract_id" db:"contract_id"`
		Address        string `json:"address" db:"address"`
		ConnectionType string `json:"connection_type" db:"connection_type"`
		TariffID       int    `json:"tariff_id" db:"tariff_id"`
		IPAddress      string `json:"ip_address" db:"ip_address"`
		Mask           int    `json:"mask" db:"mask"`
		IsBlocked      bool   `json:"is_blocked" db:"is_blocked"`
		ContractNumber string `json:"contract_number" db:"contract_number"`
		TariffName     string `json:"tariff_name" db:"tariff_name"`
		EquipmentModel string `json:"equipment_model" db:"equipment_model"`
	}

	var connectionsWithDetails []ConnectionWithDetails

	query := `
		SELECT 
			c.id, c.equipment_id, c.contract_id, c.address, c.connection_type,
			c.tariff_id, c.ip_address, c.mask, c.is_blocked,
			cont.number as contract_number,
			t.name as tariff_name,
			e.model as equipment_model
		FROM connections c
		LEFT JOIN contracts cont ON c.contract_id = cont.id
		LEFT JOIN tariffs t ON c.tariff_id = t.id
		LEFT JOIN equipment e ON c.equipment_id = e.id
		ORDER BY c.id
	`

	if err := h.DB.Select(&connectionsWithDetails, query); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectionsWithDetails)
}

// @Summary      Получить подключение по ID
// @Description  Возвращает одно подключение по ID
// @Tags         Connections
// @Produce      json
// @Param        id   path      int  true  "ID Подключения"
// @Success      200  {object}  models.Connection
// @Failure      404  {object}  map[string]string
// @Router       /connections/{id} [get]
// @Security     BearerAuth
func (h *BillingHandler) GetConnectionByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	conn := models.Connection{}
	if err := h.DB.Get(&conn, "SELECT * FROM connections WHERE id=$1", id); err != nil {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn)
}

// @Summary      Обновить подключение
// @Description  Обновляет данные подключения по ID
// @Tags         Connections
// @Accept       json
// @Param        id         path      int               true  "ID Подключения"
// @Param        connection body      models.Connection true  "Обновленные данные"
// @Success      200        {string}  string "OK"
// @Failure      404        {object}  map[string]string
// @Router       /connections/{id} [put]
// @Security     BearerAuth
func (h *BillingHandler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var conn models.Connection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `UPDATE connections SET equipment_id=$1, contract_id=$2, address=$3, connection_type=$4, tariff_id=$5, ip_address=$6, mask=$7, is_blocked=$8 WHERE id=$9`
	res, err := h.DB.Exec(query, conn.EquipmentID, conn.ContractID, conn.Address, conn.ConnectionType, conn.TariffID, conn.IPAddress, conn.Mask, conn.IsBlocked, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Connection not found or not updated", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary      Удалить подключение
// @Description  Удаляет подключение по ID
// @Tags         Connections
// @Param        id   path      int  true  "ID Подключения"
// @Success      204  {string}  string "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /connections/{id} [delete]
// @Security     BearerAuth
func (h *BillingHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("DELETE FROM connections WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Получить подключения по договору
// @Description  Возвращает все подключения для указанного договора
// @Tags         Connections
// @Produce      json
// @Param        contract_id   path      int  true  "ID Договора"
// @Success      200  {array}   models.Connection
// @Failure      500  {object}  map[string]string
// @Router       /contracts/{contract_id}/connections [get]
// @Security     BearerAuth
func (h *BillingHandler) GetConnectionsByContract(w http.ResponseWriter, r *http.Request) {
	contractID := mux.Vars(r)["contract_id"]

	type ConnectionWithDetails struct {
		ID             int    `json:"id" db:"id"`
		EquipmentID    int    `json:"equipment_id" db:"equipment_id"`
		ContractID     int    `json:"contract_id" db:"contract_id"`
		Address        string `json:"address" db:"address"`
		ConnectionType string `json:"connection_type" db:"connection_type"`
		TariffID       int    `json:"tariff_id" db:"tariff_id"`
		IPAddress      string `json:"ip_address" db:"ip_address"`
		Mask           int    `json:"mask" db:"mask"`
		IsBlocked      bool   `json:"is_blocked" db:"is_blocked"`
		ContractNumber string `json:"contract_number" db:"contract_number"`
		TariffName     string `json:"tariff_name" db:"tariff_name"`
		EquipmentModel string `json:"equipment_model" db:"equipment_model"`
	}

	var connections []ConnectionWithDetails

	query := `
		SELECT 
			c.id, c.equipment_id, c.contract_id, c.address, c.connection_type,
			c.tariff_id, c.ip_address, c.mask, c.is_blocked,
			cont.number as contract_number,
			t.name as tariff_name,
			e.model as equipment_model
		FROM connections c
		LEFT JOIN contracts cont ON c.contract_id = cont.id
		LEFT JOIN tariffs t ON c.tariff_id = t.id
		LEFT JOIN equipment e ON c.equipment_id = e.id
		WHERE c.contract_id = $1
		ORDER BY c.id
	`

	if err := h.DB.Select(&connections, query, contractID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

// @Summary      Получить договоры по клиенту
// @Description  Возвращает список всех договоров для конкретного клиента
// @Tags         Contracts
// @Produce      json
// @Param        client_id path int true "ID клиента"
// @Success      200  {array}   object{id=int,number=string,sign_date=string,is_blocked=bool,connections_count=int}
// @Failure      500  {object}  map[string]string
// @Router       /clients/{client_id}/contracts [get]
// @Security     BearerAuth
func (h *BillingHandler) GetContractsByClient(w http.ResponseWriter, r *http.Request) {
	clientID := mux.Vars(r)["client_id"]

	type ContractWithDetails struct {
		ID               int    `json:"id" db:"id"`
		Number           string `json:"number" db:"number"`
		SignDate         string `json:"sign_date" db:"sign_date"`
		IsBlocked        bool   `json:"is_blocked" db:"is_blocked"`
		ConnectionsCount int    `json:"connections_count" db:"connections_count"`
	}

	var contracts []ContractWithDetails

	query := `
		SELECT 
			c.id, c.number, c.sign_date, c.is_blocked,
			COALESCE(COUNT(conn.id), 0) as connections_count
		FROM contracts c
		LEFT JOIN connections conn ON c.id = conn.contract_id
		WHERE c.client_id = $1
		GROUP BY c.id, c.number, c.sign_date, c.is_blocked
		ORDER BY c.id
	`

	if err := h.DB.Select(&contracts, query, clientID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contracts)
}

// @Summary      Заблокировать подключение
// @Description  Блокирует указанное подключение
// @Tags         Connections
// @Produce      json
// @Param        id   path      int  true  "ID Подключения"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /connections/{id}/block [post]
// @Security     BearerAuth
func (h *BillingHandler) BlockConnection(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("UPDATE connections SET is_blocked = true WHERE id = $1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection blocked successfully"})
}

// @Summary      Разблокировать подключение
// @Description  Разблокирует указанное подключение
// @Tags         Connections
// @Produce      json
// @Param        id   path      int  true  "ID Подключения"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /connections/{id}/unblock [post]
// @Security     BearerAuth
func (h *BillingHandler) UnblockConnection(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("UPDATE connections SET is_blocked = false WHERE id = $1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection unblocked successfully"})
}

//================================================================================
// BLOCKING: Методы блокировки/разблокировки
//================================================================================

// BlockClient блокирует клиента и все его договоры
// @Summary      Заблокировать клиента
// @Description  Блокирует клиента и автоматически блокирует все его договоры
// @Tags         Clients
// @Param        id   path      int  true  "ID Клиента"
// @Success      200  {string}  string "OK"
// @Failure      404  {object}  map[string]string
// @Router       /clients/{id}/block [post]
// @Security     BearerAuth
func (h *BillingHandler) BlockClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Блокируем клиента
	res, err := tx.Exec("UPDATE clients SET is_blocked=true WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Блокируем все договоры клиента
	_, err = tx.Exec("UPDATE contracts SET is_blocked=true WHERE client_id=$1", id)
	if err != nil {
		http.Error(w, "Failed to block contracts", http.StatusInternalServerError)
		return
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UnblockClient разблокирует клиента (договоры остаются заблокированными)
// @Summary      Разблокировать клиента
// @Description  Разблокирует клиента (договоры нужно разблокировать отдельно)
// @Tags         Clients
// @Param        id   path      int  true  "ID Клиента"
// @Success      200  {string}  string "OK"
// @Failure      404  {object}  map[string]string
// @Router       /clients/{id}/unblock [post]
// @Security     BearerAuth
func (h *BillingHandler) UnblockClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("UPDATE clients SET is_blocked=false WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// BlockContract блокирует договор
// @Summary      Заблокировать договор
// @Description  Блокирует конкретный договор
// @Tags         Contracts
// @Param        id   path      int  true  "ID Договора"
// @Success      200  {string}  string "OK"
// @Failure      404  {object}  map[string]string
// @Router       /contracts/{id}/block [post]
// @Security     BearerAuth
func (h *BillingHandler) BlockContract(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("UPDATE contracts SET is_blocked=true WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UnblockContract разблокирует договор
// @Summary      Разблокировать договор
// @Description  Разблокирует конкретный договор
// @Tags         Contracts
// @Param        id   path      int  true  "ID Договора"
// @Success      200  {string}  string "OK"
// @Failure      404  {object}  map[string]string
// @Router       /contracts/{id}/unblock [post]
// @Security     BearerAuth
func (h *BillingHandler) UnblockContract(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	res, err := h.DB.Exec("UPDATE contracts SET is_blocked=false WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//================================================================================
// TRAFFIC DASHBOARD
//================================================================================

// @Summary      Получить данные трафика с фильтрацией
// @Description  Возвращает данные трафика с возможностью фильтрации по клиенту, IP, временному интервалу
// @Tags         Traffic
// @Produce      json
// @Param        client_id query int false "ID клиента"
// @Param        ip_address query string false "IP адрес"
// @Param        from query string false "Дата начала (YYYY-MM-DD HH:MM:SS)"
// @Param        to query string false "Дата окончания (YYYY-MM-DD HH:MM:SS)"
// @Param        limit query int false "Лимит записей (по умолчанию 100)"
// @Param        offset query int false "Смещение записей (по умолчанию 0)"
// @Success      200  {array}   models.TrafficResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /traffic [get]
// @Security     BearerAuth
func (h *BillingHandler) GetTrafficData(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT 
			t.id,
			t.connection_id,
			t.client_id,
			COALESCE(cl.first_name || ' ' || cl.last_name, cl.full_name, cl.short_name, cl.email) as client_name,
			cl.email as client_email,
			c.ip_address,
			t.timestamp,
			t.bytes_in,
			t.bytes_out,
			t.packets_in,
			t.packets_out,
			(t.bytes_in + t.bytes_out) as total_traffic
		FROM traffic t
		LEFT JOIN clients cl ON t.client_id = cl.id
		LEFT JOIN connections c ON t.connection_id = c.id
	`

	if clientID := queryParams.Get("client_id"); clientID != "" {
		if id, err := strconv.Atoi(clientID); err == nil {
			whereClauses = append(whereClauses, "t.client_id = $"+strconv.Itoa(argIndex))
			args = append(args, id)
			argIndex++
		}
	}

	if ipAddress := queryParams.Get("ip_address"); ipAddress != "" {
		whereClauses = append(whereClauses, "c.ip_address ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+ipAddress+"%")
		argIndex++
	}

	if fromDate := queryParams.Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02 15:04:05", fromDate); err == nil {
			whereClauses = append(whereClauses, "t.timestamp >= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		} else if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			whereClauses = append(whereClauses, "t.timestamp >= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		}
	}

	if toDate := queryParams.Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02 15:04:05", toDate); err == nil {
			whereClauses = append(whereClauses, "t.timestamp <= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		} else if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			// Добавляем день к дате в Go, а не в SQL
			endDate := parsedDate.Add(24 * time.Hour)
			whereClauses = append(whereClauses, "t.timestamp <= $"+strconv.Itoa(argIndex))
			args = append(args, endDate)
			argIndex++
		}
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	baseQuery += " ORDER BY t.timestamp DESC"

	limit := 100
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := queryParams.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	baseQuery += " LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	var traffic []models.TrafficResponse
	err := h.DB.Select(&traffic, baseQuery, args...)
	if err != nil {
		http.Error(w, "Failed to fetch traffic data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(traffic)
}

// @Summary      Получить статистику трафика
// @Description  Возвращает агрегированную статистику по трафику
// @Tags         Traffic
// @Produce      json
// @Param        client_id query int false "ID клиента"
// @Param        from query string false "Дата начала (YYYY-MM-DD)"
// @Param        to query string false "Дата окончания (YYYY-MM-DD)"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /traffic/stats [get]
// @Security     BearerAuth
func (h *BillingHandler) GetTrafficStats(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT 
			COUNT(*) as total_records,
			COALESCE(SUM(t.bytes_in), 0) as total_bytes_in,
			COALESCE(SUM(t.bytes_out), 0) as total_bytes_out,
			COALESCE(SUM(t.bytes_in + t.bytes_out), 0) as total_traffic,
			COALESCE(AVG(t.bytes_in + t.bytes_out), 0) as avg_traffic,
			COALESCE(MAX(t.bytes_in + t.bytes_out), 0) as max_traffic,
			COALESCE(MIN(t.bytes_in + t.bytes_out), 0) as min_traffic
		FROM traffic t
		LEFT JOIN connections c ON t.connection_id = c.id
	`

	if clientID := queryParams.Get("client_id"); clientID != "" {
		if id, err := strconv.Atoi(clientID); err == nil {
			whereClauses = append(whereClauses, "t.client_id = $"+strconv.Itoa(argIndex))
			args = append(args, id)
			argIndex++
		}
	}

	if ipAddress := queryParams.Get("ip_address"); ipAddress != "" {
		whereClauses = append(whereClauses, "c.ip_address ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+ipAddress+"%")
		argIndex++
	}

	if fromDate := queryParams.Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			whereClauses = append(whereClauses, "t.timestamp >= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		}
	}

	if toDate := queryParams.Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			// Добавляем день к дате в Go, а не в SQL
			endDate := parsedDate.Add(24 * time.Hour)
			whereClauses = append(whereClauses, "t.timestamp <= $"+strconv.Itoa(argIndex))
			args = append(args, endDate)
			argIndex++
		}
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var stats struct {
		TotalRecords  int     `db:"total_records"`
		TotalBytesIn  int64   `db:"total_bytes_in"`
		TotalBytesOut int64   `db:"total_bytes_out"`
		TotalTraffic  int64   `db:"total_traffic"`
		AvgTraffic    float64 `db:"avg_traffic"`
		MaxTraffic    int64   `db:"max_traffic"`
		MinTraffic    int64   `db:"min_traffic"`
	}

	err := h.DB.Get(&stats, baseQuery, args...)
	if err != nil {
		http.Error(w, "Failed to fetch traffic stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total_records":   stats.TotalRecords,
		"total_bytes_in":  stats.TotalBytesIn,
		"total_bytes_out": stats.TotalBytesOut,
		"total_traffic":   stats.TotalTraffic,
		"avg_traffic":     stats.AvgTraffic,
		"max_traffic":     stats.MaxTraffic,
		"min_traffic":     stats.MinTraffic,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary      Получить IP-адреса клиента
// @Description  Возвращает все IP-адреса подключений клиента
// @Tags         Clients
// @Produce      json
// @Param        id   path      int  true  "ID клиента"
// @Success      200  {array}   string
// @Failure      404  {object}  map[string]string
// @Router       /clients/{id}/ips [get]
// @Security     BearerAuth
func (h *BillingHandler) GetClientIPs(w http.ResponseWriter, r *http.Request) {
	clientID := mux.Vars(r)["id"]
	
	var ips []string
	query := `
		SELECT DISTINCT c.ip_address
		FROM connections c
		JOIN contracts ct ON c.contract_id = ct.id
		WHERE ct.client_id = $1 AND c.is_blocked = false
		ORDER BY c.ip_address
	`
	
	err := h.DB.Select(&ips, query, clientID)
	if err != nil {
		http.Error(w, "Client not found or no connections", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"client_id": clientID,
		"ip_addresses": ips,
	})
}

// @Summary      Экспорт данных трафика в CSV
// @Description  Экспортирует данные трафика в CSV формате с возможностью фильтрации
// @Tags         Traffic
// @Produce      text/csv
// @Param        client_id query int false "ID клиента"
// @Param        ip_address query string false "IP адрес"
// @Param        from query string false "Дата начала (YYYY-MM-DD HH:MM:SS)"
// @Param        to query string false "Дата окончания (YYYY-MM-DD HH:MM:SS)"
// @Success      200  {string}  string "CSV data"
// @Failure      500  {object}  map[string]string
// @Router       /traffic/export [get]
// @Security     BearerAuth
func (h *BillingHandler) ExportTrafficCSV(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT 
			t.id,
			t.connection_id,
			t.client_id,
			COALESCE(cl.first_name || ' ' || cl.last_name, cl.full_name, cl.short_name, cl.email) as client_name,
			cl.email as client_email,
			c.ip_address,
			t.timestamp,
			t.bytes_in,
			t.bytes_out,
			t.packets_in,
			t.packets_out,
			(t.bytes_in + t.bytes_out) as total_traffic
		FROM traffic t
		LEFT JOIN clients cl ON t.client_id = cl.id
		LEFT JOIN connections c ON t.connection_id = c.id
	`

	// Применяем те же фильтры что и в GetTrafficData
	if clientID := queryParams.Get("client_id"); clientID != "" {
		if id, err := strconv.Atoi(clientID); err == nil {
			whereClauses = append(whereClauses, "t.client_id = $"+strconv.Itoa(argIndex))
			args = append(args, id)
			argIndex++
		}
	}

	if ipAddress := queryParams.Get("ip_address"); ipAddress != "" {
		whereClauses = append(whereClauses, "c.ip_address ILIKE $"+strconv.Itoa(argIndex))
		args = append(args, "%"+ipAddress+"%")
		argIndex++
	}

	if fromDate := queryParams.Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02 15:04:05", fromDate); err == nil {
			whereClauses = append(whereClauses, "t.timestamp >= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		} else if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			whereClauses = append(whereClauses, "t.timestamp >= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		}
	}

	if toDate := queryParams.Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02 15:04:05", toDate); err == nil {
			whereClauses = append(whereClauses, "t.timestamp <= $"+strconv.Itoa(argIndex))
			args = append(args, parsedDate)
			argIndex++
		} else if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			// Добавляем день к дате в Go, а не в SQL
			endDate := parsedDate.Add(24 * time.Hour)
			whereClauses = append(whereClauses, "t.timestamp <= $"+strconv.Itoa(argIndex))
			args = append(args, endDate)
			argIndex++
		}
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	baseQuery += " ORDER BY t.timestamp DESC"

	var traffic []models.TrafficResponse
	err := h.DB.Select(&traffic, baseQuery, args...)
	if err != nil {
		http.Error(w, "Failed to fetch traffic data for export: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки для CSV
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=traffic_export_"+time.Now().Format("2006-01-02_15-04-05")+".csv")

	// Добавляем BOM для корректного отображения в Excel
	w.Write([]byte("\xEF\xBB\xBF"))

	// Записываем заголовки CSV
	csvHeader := "ID,ID Подключения,ID Клиента,Имя Клиента,Email Клиента,IP Адрес,Время,Входящий Трафик (байт),Исходящий Трафик (байт),Входящие Пакеты,Исходящие Пакеты,Общий Трафик (байт)\n"
	w.Write([]byte(csvHeader))

	// Записываем данные
	for _, item := range traffic {
		clientName := item.ClientName
		if clientName == "" {
			clientName = "N/A"
		}

		csvLine := fmt.Sprintf("%d,%d,%d,\"%s\",\"%s\",\"%s\",\"%s\",%d,%d,%d,%d,%d\n",
			item.ID,
			item.ConnectionID,
			item.ClientID,
			clientName,
			item.ClientEmail,
			item.IPAddress,
			item.Timestamp.Format("2006-01-02 15:04:05"),
			item.BytesIn,
			item.BytesOut,
			item.PacketsIn,
			item.PacketsOut,
			item.TotalTraffic,
		)
		w.Write([]byte(csvLine))
	}
}

// @Summary      Получить статистику по договору
// @Description  Возвращает детальную статистику трафика по конкретному договору
// @Tags         Contracts
// @Produce      json
// @Param        id path int true "ID договора"
// @Param        from query string false "Дата начала (YYYY-MM-DD)"
// @Param        to query string false "Дата окончания (YYYY-MM-DD)"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /contracts/{id}/stats [get]
// @Security     BearerAuth
func (h *BillingHandler) GetContractStats(w http.ResponseWriter, r *http.Request) {
	contractID := mux.Vars(r)["id"]
	queryParams := r.URL.Query()

	// Проверяем существование договора
	var contractExists bool
	err := h.DB.Get(&contractExists, "SELECT EXISTS(SELECT 1 FROM contracts WHERE id=$1)", contractID)
	if err != nil || !contractExists {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	var args []interface{}
	argIndex := 1

	// Основной запрос для статистики по договору
	baseQuery := `
		SELECT 
			COUNT(*) as total_records,
			COALESCE(SUM(t.bytes_in), 0) as total_bytes_in,
			COALESCE(SUM(t.bytes_out), 0) as total_bytes_out,
			COALESCE(SUM(t.bytes_in + t.bytes_out), 0) as total_traffic,
			COALESCE(AVG(t.bytes_in + t.bytes_out), 0) as avg_traffic,
			COALESCE(MAX(t.bytes_in + t.bytes_out), 0) as max_traffic,
			COALESCE(MIN(t.bytes_in + t.bytes_out), 0) as min_traffic,
			COUNT(DISTINCT DATE(t.timestamp)) as active_days
		FROM traffic t
		JOIN connections c ON t.connection_id = c.id
		JOIN contracts ct ON c.contract_id = ct.id
		WHERE ct.id = $` + strconv.Itoa(argIndex)

	args = append(args, contractID)
	argIndex++

	if fromDate := queryParams.Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			baseQuery += " AND t.timestamp >= $" + strconv.Itoa(argIndex)
			args = append(args, parsedDate)
			argIndex++
		}
	}

	if toDate := queryParams.Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			// Добавляем день к дате в Go, а не в SQL
			endDate := parsedDate.Add(24 * time.Hour)
			baseQuery += " AND t.timestamp <= $" + strconv.Itoa(argIndex)
			args = append(args, endDate)
			argIndex++
		}
	}

	var stats struct {
		TotalRecords  int     `db:"total_records"`
		TotalBytesIn  int64   `db:"total_bytes_in"`
		TotalBytesOut int64   `db:"total_bytes_out"`
		TotalTraffic  int64   `db:"total_traffic"`
		AvgTraffic    float64 `db:"avg_traffic"`
		MaxTraffic    int64   `db:"max_traffic"`
		MinTraffic    int64   `db:"min_traffic"`
		ActiveDays    int     `db:"active_days"`
	}

	err = h.DB.Get(&stats, baseQuery, args...)
	if err != nil {
		http.Error(w, "Failed to fetch contract stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем информацию о договоре и клиенте
	contractInfoQuery := `
		SELECT 
			ct.id,
			ct.number,
			ct.sign_date,
			ct.is_blocked,
			cl.id as client_id,
			COALESCE(cl.first_name || ' ' || cl.last_name, cl.full_name, cl.short_name, cl.email) as client_name,
			cl.email as client_email,
			COUNT(c.id) as connections_count
		FROM contracts ct
		JOIN clients cl ON ct.client_id = cl.id
		LEFT JOIN connections c ON ct.id = c.contract_id
		WHERE ct.id = $1
		GROUP BY ct.id, ct.number, ct.sign_date, ct.is_blocked, cl.id, cl.first_name, cl.last_name, cl.full_name, cl.short_name, cl.email
	`

	var contractInfo struct {
		ID               int       `db:"id"`
		Number           string    `db:"number"`
		SignDate         time.Time `db:"sign_date"`
		IsBlocked        bool      `db:"is_blocked"`
		ClientID         int       `db:"client_id"`
		ClientName       string    `db:"client_name"`
		ClientEmail      string    `db:"client_email"`
		ConnectionsCount int       `db:"connections_count"`
	}

	err = h.DB.Get(&contractInfo, contractInfoQuery, contractID)
	if err != nil {
		http.Error(w, "Failed to fetch contract info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем топ дней по трафику
	topDaysQuery := `
		SELECT 
			DATE(t.timestamp) as date,
			COALESCE(SUM(t.bytes_in + t.bytes_out), 0) as daily_traffic
		FROM traffic t
		JOIN connections c ON t.connection_id = c.id
		JOIN contracts ct ON c.contract_id = ct.id
		WHERE ct.id = $1
	`
	topDaysArgs := []interface{}{contractID}
	argIdx := 2

	if fromDate := queryParams.Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			topDaysQuery += " AND t.timestamp >= $" + strconv.Itoa(argIdx)
			topDaysArgs = append(topDaysArgs, parsedDate)
			argIdx++
		}
	}

	if toDate := queryParams.Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			// Добавляем день к дате в Go, а не в SQL
			endDate := parsedDate.Add(24 * time.Hour)
			topDaysQuery += " AND t.timestamp <= $" + strconv.Itoa(argIdx)
			topDaysArgs = append(topDaysArgs, endDate)
			argIdx++
		}
	}

	topDaysQuery += `
		GROUP BY DATE(t.timestamp)
		ORDER BY daily_traffic DESC
		LIMIT 5
	`

	var topDays []struct {
		Date         time.Time `db:"date"`
		DailyTraffic int64     `db:"daily_traffic"`
	}

	err = h.DB.Select(&topDays, topDaysQuery, topDaysArgs...)
	if err != nil {
		// log.Printf("Error fetching top days: %v", err)
		topDays = []struct {
			Date         time.Time `db:"date"`
			DailyTraffic int64     `db:"daily_traffic"`
		}{}
	}

	response := map[string]interface{}{
		"contract": map[string]interface{}{
			"id":                contractInfo.ID,
			"number":            contractInfo.Number,
			"sign_date":         contractInfo.SignDate.Format("2006-01-02"),
			"is_blocked":        contractInfo.IsBlocked,
			"client_id":         contractInfo.ClientID,
			"client_name":       contractInfo.ClientName,
			"client_email":      contractInfo.ClientEmail,
			"connections_count": contractInfo.ConnectionsCount,
		},
		"traffic_stats": map[string]interface{}{
			"total_records":   stats.TotalRecords,
			"total_bytes_in":  stats.TotalBytesIn,
			"total_bytes_out": stats.TotalBytesOut,
			"total_traffic":   stats.TotalTraffic,
			"avg_traffic":     stats.AvgTraffic,
			"max_traffic":     stats.MaxTraffic,
			"min_traffic":     stats.MinTraffic,
			"active_days":     stats.ActiveDays,
		},
		"top_days": topDays,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//================================================================================
// CONNECTION STATS
//================================================================================

// @Summary      Получить статистику по подключению
// @Description  Возвращает агрегированную статистику трафика по подключению
// @Tags         Connections
// @Produce      json
// @Param        id path string true "ID подключения"
// @Param        from query string false "Дата начала (YYYY-MM-DD)"
// @Param        to query string false "Дата окончания (YYYY-MM-DD)"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /connections/{id}/stats [get]
// @Security     BearerAuth
func (h *BillingHandler) GetConnectionStats(w http.ResponseWriter, r *http.Request) {
	connectionID := mux.Vars(r)["id"]
	queryParams := r.URL.Query()

	// Проверяем существование подключения
	var connectionExists bool
	err := h.DB.Get(&connectionExists, "SELECT EXISTS(SELECT 1 FROM connections WHERE id=$1)", connectionID)
	if err != nil || !connectionExists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	var args []interface{}
	argIndex := 1

	// Основной запрос для статистики по подключению
	baseQuery := `
		SELECT 
			COUNT(*) as total_records,
			COALESCE(SUM(t.bytes_in), 0) as total_bytes_in,
			COALESCE(SUM(t.bytes_out), 0) as total_bytes_out,
			COALESCE(SUM(t.bytes_in + t.bytes_out), 0) as total_traffic,
			COALESCE(AVG(t.bytes_in + t.bytes_out), 0) as avg_traffic,
			COALESCE(MAX(t.bytes_in + t.bytes_out), 0) as max_traffic,
			COALESCE(MIN(t.bytes_in + t.bytes_out), 0) as min_traffic,
			COUNT(DISTINCT DATE(t.timestamp)) as active_days
		FROM traffic t
		WHERE t.connection_id = $` + strconv.Itoa(argIndex)

	args = append(args, connectionID)
	argIndex++

	if fromDate := queryParams.Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			baseQuery += " AND t.timestamp >= $" + strconv.Itoa(argIndex)
			args = append(args, parsedDate)
			argIndex++
		}
	}

	if toDate := queryParams.Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			endDate := parsedDate.Add(24 * time.Hour)
			baseQuery += " AND t.timestamp <= $" + strconv.Itoa(argIndex)
			args = append(args, endDate)
			argIndex++
		}
	}

	var stats struct {
		TotalRecords  int     `db:"total_records"`
		TotalBytesIn  int64   `db:"total_bytes_in"`
		TotalBytesOut int64   `db:"total_bytes_out"`
		TotalTraffic  int64   `db:"total_traffic"`
		AvgTraffic    float64 `db:"avg_traffic"`
		MaxTraffic    int64   `db:"max_traffic"`
		MinTraffic    int64   `db:"min_traffic"`
		ActiveDays    int     `db:"active_days"`
	}

	err = h.DB.Get(&stats, baseQuery, args...)
	if err != nil {
		http.Error(w, "Failed to fetch connection stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем информацию о подключении
	connectionInfoQuery := `
		SELECT 
			c.id,
			c.address,
			c.ip_address,
			c.mask,
			c.connection_type,
			c.is_blocked,
			ct.id as contract_id,
			ct.number as contract_number,
			t.name as tariff_name,
			e.model as equipment_model,
			cl.id as client_id,
			COALESCE(cl.first_name || ' ' || cl.last_name, cl.full_name, cl.short_name, cl.email) as client_name,
			cl.email as client_email
		FROM connections c
		JOIN contracts ct ON c.contract_id = ct.id
		JOIN clients cl ON ct.client_id = cl.id
		LEFT JOIN tariffs t ON c.tariff_id = t.id
		LEFT JOIN equipment e ON c.equipment_id = e.id
		WHERE c.id = $1
	`

	var connectionInfo struct {
		ID             int     `db:"id"`
		Address        string  `db:"address"`
		IPAddress      string  `db:"ip_address"`
		Mask           int     `db:"mask"`
		ConnectionType string  `db:"connection_type"`
		IsBlocked      bool    `db:"is_blocked"`
		ContractID     int     `db:"contract_id"`
		ContractNumber string  `db:"contract_number"`
		TariffName     *string `db:"tariff_name"`
		EquipmentModel *string `db:"equipment_model"`
		ClientID       int     `db:"client_id"`
		ClientName     string  `db:"client_name"`
		ClientEmail    string  `db:"client_email"`
	}

	err = h.DB.Get(&connectionInfo, connectionInfoQuery, connectionID)
	if err != nil {
		http.Error(w, "Failed to fetch connection info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем топ дней по трафику для подключения
	topDaysQuery := `
		SELECT 
			DATE(t.timestamp) as date,
			COALESCE(SUM(t.bytes_in + t.bytes_out), 0) as daily_traffic
		FROM traffic t
		WHERE t.connection_id = $1
	`
	topDaysArgs := []interface{}{connectionID}
	argIdx := 2

	if fromDate := queryParams.Get("from"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			topDaysQuery += " AND t.timestamp >= $" + strconv.Itoa(argIdx)
			topDaysArgs = append(topDaysArgs, parsedDate)
			argIdx++
		}
	}

	if toDate := queryParams.Get("to"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			endDate := parsedDate.Add(24 * time.Hour)
			topDaysQuery += " AND t.timestamp <= $" + strconv.Itoa(argIdx)
			topDaysArgs = append(topDaysArgs, endDate)
			argIdx++
		}
	}

	topDaysQuery += `
		GROUP BY DATE(t.timestamp)
		ORDER BY daily_traffic DESC
		LIMIT 5
	`

	var topDays []struct {
		Date         time.Time `db:"date"`
		DailyTraffic int64     `db:"daily_traffic"`
	}

	err = h.DB.Select(&topDays, topDaysQuery, topDaysArgs...)
	if err != nil {
		topDays = []struct {
			Date         time.Time `db:"date"`
			DailyTraffic int64     `db:"daily_traffic"`
		}{}
	}

	tariffName := "Не указан"
	if connectionInfo.TariffName != nil {
		tariffName = *connectionInfo.TariffName
	}

	equipmentModel := "Не указано"
	if connectionInfo.EquipmentModel != nil {
		equipmentModel = *connectionInfo.EquipmentModel
	}

	response := map[string]interface{}{
		"connection": map[string]interface{}{
			"id":              connectionInfo.ID,
			"address":         connectionInfo.Address,
			"ip_address":      connectionInfo.IPAddress,
			"mask":            connectionInfo.Mask,
			"connection_type": connectionInfo.ConnectionType,
			"is_blocked":      connectionInfo.IsBlocked,
			"contract_id":     connectionInfo.ContractID,
			"contract_number": connectionInfo.ContractNumber,
			"tariff_name":     tariffName,
			"equipment_model": equipmentModel,
			"client_id":       connectionInfo.ClientID,
			"client_name":     connectionInfo.ClientName,
			"client_email":    connectionInfo.ClientEmail,
		},
		"traffic_stats": map[string]interface{}{
			"total_records":   stats.TotalRecords,
			"total_bytes_in":  stats.TotalBytesIn,
			"total_bytes_out": stats.TotalBytesOut,
			"total_traffic":   stats.TotalTraffic,
			"avg_traffic":     stats.AvgTraffic,
			"max_traffic":     stats.MaxTraffic,
			"min_traffic":     stats.MinTraffic,
			"active_days":     stats.ActiveDays,
		},
		"top_days": topDays,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//================================================================================
// CRUD: ISSUES (Доработки)
//================================================================================

// @Summary      Создать задачу
// @Description  Создает новую задачу в системе доработок
// @Tags         Issues
// @Accept       json
// @Produce      json
// @Param        issue body models.Issue true "Объект новой задачи"
// @Success      201  {object}  models.Issue
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues [post]
// @Security     BearerAuth
func (h *BillingHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var issue models.Issue
	if err := json.NewDecoder(r.Body).Decode(&issue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Если created_by не указан, используем значение по умолчанию (например, 1 - admin)
	if issue.CreatedBy == 0 {
		issue.CreatedBy = 1 // TODO: Get from auth context
	}

	query := `INSERT INTO issues (title, description, status, created_by) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := h.DB.QueryRowx(query, issue.Title, issue.Description, models.NewIssue, issue.CreatedBy).Scan(&issue.ID, &issue.CreatedAt)
	if err != nil {
		http.Error(w, "Could not create issue: "+err.Error(), http.StatusInternalServerError)
		return
	}

	issue.Status = models.NewIssue

	// Отправляем уведомление в Telegram
	if h.TelegramService != nil {
		go h.TelegramService.SendIssueCreated(&issue)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(issue)
}

// @Summary      Получить список задач
// @Description  Возвращает список всех задач с возможностью фильтрации по статусу
// @Tags         Issues
// @Produce      json
// @Param        status query string false "Статус задачи (new/resolved)"
// @Success      200  {array}  models.Issue
// @Failure      500  {object}  map[string]string
// @Router       /issues [get]
// @Security     BearerAuth
func (h *BillingHandler) GetIssues(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	baseQuery := "SELECT * FROM issues"
	var args []interface{}

	if status := queryParams.Get("status"); status != "" {
		baseQuery += " WHERE status = $1"
		args = append(args, status)
	}

	baseQuery += " ORDER BY created_at DESC"

	issues := []models.Issue{}
	if err := h.DB.Select(&issues, baseQuery, args...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues)
}

// @Summary      Получить задачу по ID
// @Description  Возвращает одну задачу по её ID
// @Tags         Issues
// @Produce      json
// @Param        id   path      int  true  "ID Задачи"
// @Success      200  {object}  models.Issue
// @Failure      404  {object}  map[string]string
// @Router       /issues/{id} [get]
// @Security     BearerAuth
func (h *BillingHandler) GetIssueByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	issue := models.Issue{}
	if err := h.DB.Get(&issue, "SELECT * FROM issues WHERE id=$1", id); err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issue)
}

// @Summary      Обновить задачу
// @Description  Обновляет данные задачи по ID
// @Tags         Issues
// @Accept       json
// @Param        id    path      int           true  "ID Задачи"
// @Param        issue body      models.Issue  true  "Обновленные данные"
// @Success      200   {string}  string "OK"
// @Failure      404   {object}  map[string]string
// @Router       /issues/{id} [put]
// @Security     BearerAuth
func (h *BillingHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var updatedIssue models.Issue
	if err := json.NewDecoder(r.Body).Decode(&updatedIssue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем текущие данные задачи для сравнения
	var currentIssue models.Issue
	err := h.DB.Get(&currentIssue, "SELECT * FROM issues WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Запрещаем редактирование решенных задач
	if currentIssue.Status == models.ResolvedIssue {
		http.Error(w, "Cannot edit resolved issues", http.StatusBadRequest)
		return
	}

	// Начинаем транзакцию для обновления задачи и записи истории
	tx, err := h.DB.Beginx()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Обновляем задачу
	updateQuery := `UPDATE issues SET title=$1, description=$2 WHERE id=$3`
	res, err := tx.Exec(updateQuery, updatedIssue.Title, updatedIssue.Description, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Issue not found or not updated", http.StatusNotFound)
		return
	}

	// Записываем историю изменений
	editedBy := 1 // TODO: Get from auth context

	// Проверяем изменения в заголовке
	if currentIssue.Title != updatedIssue.Title {
		_, err = tx.Exec(`INSERT INTO issue_history (issue_id, field_name, old_value, new_value, edited_by) VALUES ($1, $2, $3, $4, $5)`,
			id, "title", currentIssue.Title, updatedIssue.Title, editedBy)
		if err != nil {
			http.Error(w, "Failed to save title change history", http.StatusInternalServerError)
			return
		}
	}

	// Проверяем изменения в описании
	if currentIssue.Description != updatedIssue.Description {
		_, err = tx.Exec(`INSERT INTO issue_history (issue_id, field_name, old_value, new_value, edited_by) VALUES ($1, $2, $3, $4, $5)`,
			id, "description", currentIssue.Description, updatedIssue.Description, editedBy)
		if err != nil {
			http.Error(w, "Failed to save description change history", http.StatusInternalServerError)
			return
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Отправляем уведомление в Telegram о изменениях
	if h.TelegramService != nil {
		changes := []string{}
		if currentIssue.Title != updatedIssue.Title {
			changes = append(changes, fmt.Sprintf("Название: %s → %s", currentIssue.Title, updatedIssue.Title))
		}
		if currentIssue.Description != updatedIssue.Description {
			changes = append(changes, fmt.Sprintf("Описание изменено"))
		}
		if len(changes) > 0 {
			go h.TelegramService.SendIssueUpdated(&currentIssue, &updatedIssue, changes)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Решить задачу
// @Description  Отмечает задачу как решенную и устанавливает время решения
// @Tags         Issues
// @Accept       json
// @Param        id   path      int  true  "ID Задачи"
// @Param        resolved_by body object{resolved_by=int} true "ID пользователя, решившего задачу"
// @Success      200  {string}  string "OK"
// @Failure      404  {object}  map[string]string
// @Router       /issues/{id}/resolve [post]
// @Security     BearerAuth
func (h *BillingHandler) ResolveIssue(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var payload struct {
		ResolvedBy int `json:"resolved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Сначала получаем данные задачи для уведомления
	var issue models.Issue
	err := h.DB.Get(&issue, "SELECT * FROM issues WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	query := `UPDATE issues SET status=$1, resolved_at=NOW(), resolved_by=$2 WHERE id=$3 AND status='new'`
	res, err := h.DB.Exec(query, models.ResolvedIssue, payload.ResolvedBy, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Issue not found or already resolved", http.StatusNotFound)
		return
	}

	// Обновляем issue структуру для уведомления
	issue.Status = models.ResolvedIssue
	issue.ResolvedBy = &payload.ResolvedBy

	// Отправляем уведомление в Telegram
	if h.TelegramService != nil {
		go h.TelegramService.SendIssueResolved(&issue)
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Вернуть задачу в статус "новая"
// @Description  Возвращает решенную задачу обратно в статус "новая" с логированием
// @Tags         Issues
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID Задачи"
// @Param        payload body object{unresolve_reason:string,unresolve_by:int} true "Причина возврата и ID пользователя"
// @Success      200  {string}  string "OK"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /issues/{id}/unresolve [post]
// @Security     BearerAuth
func (h *BillingHandler) UnresolveIssue(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var payload struct {
		UnresolveReason string `json:"unresolve_reason"`
		UnresolveBy     int    `json:"unresolve_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем текущие данные задачи
	var issue models.Issue
	err := h.DB.Get(&issue, "SELECT * FROM issues WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Проверяем, что задача действительно решена
	if issue.Status != models.ResolvedIssue {
		http.Error(w, "Issue is not resolved", http.StatusBadRequest)
		return
	}

	// Начинаем транзакцию
	tx, err := h.DB.Beginx()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Обновляем статус задачи
	query := `UPDATE issues SET status=$1, resolved_at=NULL, resolved_by=NULL WHERE id=$2 AND status='resolved'`
	res, err := tx.Exec(query, models.NewIssue, id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Issue not found or not resolved", http.StatusNotFound)
		return
	}

	// Записываем в историю изменений
	historyQuery := `INSERT INTO issue_history (issue_id, field_name, old_value, new_value, edited_by) VALUES ($1, $2, $3, $4, $5)`
	reasonText := fmt.Sprintf("new (возврат в работу: %s)", payload.UnresolveReason)
	_, err = tx.Exec(historyQuery, id, "status", "resolved", reasonText, payload.UnresolveBy)
	if err != nil {
		http.Error(w, "Failed to log history", http.StatusInternalServerError)
		return
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Обновляем структуру для уведомления
	issue.Status = models.NewIssue
	issue.ResolvedBy = nil
	issue.ResolvedAt = nil

	// Отправляем уведомление в Telegram
	if h.TelegramService != nil {
		go h.TelegramService.SendIssueUnresolved(&issue, payload.UnresolveReason)
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Удалить задачу
// @Description  Удаляет задачу по ID
// @Tags         Issues
// @Param        id   path      int  true  "ID Задачи"
// @Success      204  {string}  string "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /issues/{id} [delete]
// @Security     BearerAuth
func (h *BillingHandler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Сначала получаем данные задачи для уведомления
	var issue models.Issue
	err := h.DB.Get(&issue, "SELECT * FROM issues WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	res, err := h.DB.Exec("DELETE FROM issues WHERE id=$1", id)
	if err != nil || mustRowsAffected(res) == 0 {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Отправляем уведомление в Telegram
	if h.TelegramService != nil {
		go h.TelegramService.SendIssueDeleted(&issue)
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Получить историю изменений задачи
// @Description  Возвращает историю всех изменений задачи по её ID
// @Tags         Issues
// @Produce      json
// @Param        id   path      int  true  "ID Задачи"
// @Success      200  {array}  models.IssueHistory
// @Failure      404  {object}  map[string]string
// @Router       /issues/{id}/history [get]
// @Security     BearerAuth
func (h *BillingHandler) GetIssueHistory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Проверяем, существует ли задача
	var issueExists bool
	err := h.DB.Get(&issueExists, "SELECT EXISTS(SELECT 1 FROM issues WHERE id=$1)", id)
	if err != nil || !issueExists {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Получаем историю изменений с информацией о пользователях
	query := `
		SELECT 
			ih.id, ih.issue_id, ih.field_name, ih.old_value, ih.new_value, 
			ih.edited_by, ih.edited_at
		FROM issue_history ih
		WHERE ih.issue_id = $1
		ORDER BY ih.edited_at DESC
	`

	var history []models.IssueHistory
	err = h.DB.Select(&history, query, id)
	if err != nil {
		http.Error(w, "Failed to fetch issue history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
