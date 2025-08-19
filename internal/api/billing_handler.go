package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"new-billing/internal/models"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type BillingHandler struct {
	DB *sqlx.DB
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
	query := `INSERT INTO clients (client_type, email, phone, first_name, last_name, patronymic, passport_number, passport_issued_by, passport_issue_date, registration_address, birth_date, inn, kpp, full_name, short_name, ogrn, ogrn_date, legal_address, actual_address, bank_details, ceo, accountant) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) RETURNING id`
	err := h.DB.QueryRow(query, client.ClientType, client.Email, client.Phone, client.FirstName, client.LastName, client.Patronymic, client.PassportNumber, client.PassportIssuedBy, client.PassportIssueDate, client.RegistrationAddress, client.BirthDate, client.INN, client.KPP, client.FullName, client.ShortName, client.OGRN, client.OGRNDate, client.LegalAddress, client.ActualAddress, client.BankDetails, client.CEO, client.Accountant).Scan(&client.ID)
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
	clients := []models.Client{}
	if err := h.DB.Select(&clients, "SELECT * FROM clients ORDER BY id"); err != nil {
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
	query := `UPDATE clients SET client_type=$1, email=$2, phone=$3, first_name=$4, last_name=$5, patronymic=$6, passport_number=$7, passport_issued_by=$8, passport_issue_date=$9, registration_address=$10, birth_date=$11, inn=$12, kpp=$13, full_name=$14, short_name=$15, ogrn=$16, ogrn_date=$17, legal_address=$18, actual_address=$19, bank_details=$20, ceo=$21, accountant=$22 WHERE id=$23`
	res, err := h.DB.Exec(query, client.ClientType, client.Email, client.Phone, client.FirstName, client.LastName, client.Patronymic, client.PassportNumber, client.PassportIssuedBy, client.PassportIssueDate, client.RegistrationAddress, client.BirthDate, client.INN, client.KPP, client.FullName, client.ShortName, client.OGRN, client.OGRNDate, client.LegalAddress, client.ActualAddress, client.BankDetails, client.CEO, client.Accountant, id)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contract)
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
	contracts := []models.Contract{}
	if err := h.DB.Select(&contracts, "SELECT * FROM contracts ORDER BY id"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contracts)
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
	query := `INSERT INTO connections (equipment_id, contract_id, address, connection_type, tariff_id, ip_address, mask) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := h.DB.QueryRow(query, conn.EquipmentID, conn.ContractID, conn.Address, conn.ConnectionType, conn.TariffID, conn.IPAddress, conn.Mask).Scan(&conn.ID)
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
	connections := []models.Connection{}
	if err := h.DB.Select(&connections, "SELECT * FROM connections ORDER BY id"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
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
	query := `UPDATE connections SET equipment_id=$1, contract_id=$2, address=$3, connection_type=$4, tariff_id=$5, ip_address=$6, mask=$7 WHERE id=$8`
	res, err := h.DB.Exec(query, conn.EquipmentID, conn.ContractID, conn.Address, conn.ConnectionType, conn.TariffID, conn.IPAddress, conn.Mask, id)
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
