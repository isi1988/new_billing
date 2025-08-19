package models

import "time"

type Role string

const (
	AdminRole   Role = "admin"
	ManagerRole Role = "manager"
	ClientRole  Role = "client" // Для будущего личного кабинета
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         Role      `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// --- Clients ---
type ClientType string

const (
	IndividualType  ClientType = "individual"
	LegalEntityType ClientType = "legal_entity"
)

type Client struct {
	ID                  int        `json:"id" db:"id"`
	ClientType          ClientType `json:"client_type" db:"client_type"`
	Email               string     `json:"email" db:"email"`
	Phone               string     `json:"phone" db:"phone"`
	FirstName           *string    `json:"first_name,omitempty" db:"first_name"`
	LastName            *string    `json:"last_name,omitempty" db:"last_name"`
	Patronymic          *string    `json:"patronymic,omitempty" db:"patronymic"`
	PassportNumber      *string    `json:"passport_number,omitempty" db:"passport_number"`
	PassportIssuedBy    *string    `json:"passport_issued_by,omitempty" db:"passport_issued_by"`
	PassportIssueDate   *time.Time `json:"passport_issue_date,omitempty" db:"passport_issue_date"`
	RegistrationAddress *string    `json:"registration_address,omitempty" db:"registration_address"`
	BirthDate           *time.Time `json:"birth_date,omitempty" db:"birth_date"`
	INN                 *string    `json:"inn,omitempty" db:"inn"`
	KPP                 *string    `json:"kpp,omitempty" db:"kpp"`
	FullName            *string    `json:"full_name,omitempty" db:"full_name"`
	ShortName           *string    `json:"short_name,omitempty" db:"short_name"`
	OGRN                *string    `json:"ogrn,omitempty" db:"ogrn"`
	OGRNDate            *time.Time `json:"ogrn_date,omitempty" db:"ogrn_date"`
	LegalAddress        *string    `json:"legal_address,omitempty" db:"legal_address"`
	ActualAddress       *string    `json:"actual_address,omitempty" db:"actual_address"`
	BankDetails         *string    `json:"bank_details,omitempty" db:"bank_details"`
	CEO                 *string    `json:"ceo,omitempty" db:"ceo"`
	Accountant          *string    `json:"accountant,omitempty" db:"accountant"`
}

// --- Other Entities ---
type Equipment struct {
	ID          int    `json:"id" db:"id"`
	Model       string `json:"model" db:"model"`
	Description string `json:"description" db:"description"`
	MACAddress  string `json:"mac_address" db:"mac_address"`
}
type PaymentType string

const (
	Postpaid PaymentType = "postpaid"
	Prepaid  PaymentType = "prepaid"
)

type Tariff struct {
	ID               int         `json:"id" db:"id"`
	Name             string      `json:"name" db:"name"`
	IsArchived       bool        `json:"is_archived" db:"is_archived"`
	PaymentType      PaymentType `json:"payment_type" db:"payment_type"`
	IsForIndividuals bool        `json:"is_for_individuals" db:"is_for_individuals"`
	MaxSpeedIn       int         `json:"max_speed_in" db:"max_speed_in"`
	MaxSpeedOut      int         `json:"max_speed_out" db:"max_speed_out"`
	MaxTrafficIn     int64       `json:"max_traffic_in" db:"max_traffic_in"`
	MaxTrafficOut    int64       `json:"max_traffic_out" db:"max_traffic_out"`
}
type Contract struct {
	ID       int       `json:"id" db:"id"`
	ClientID int       `json:"client_id" db:"client_id"`
	Number   string    `json:"number" db:"number"`
	SignDate time.Time `json:"sign_date" db:"sign_date"`
}
type Connection struct {
	ID             int    `json:"id" db:"id"`
	EquipmentID    int    `json:"equipment_id" db:"equipment_id"`
	ContractID     int    `json:"contract_id" db:"contract_id"`
	Address        string `json:"address" db:"address"`
	ConnectionType string `json:"connection_type" db:"connection_type"`
	TariffID       int    `json:"tariff_id" db:"tariff_id"`
	IPAddress      string `json:"ip_address" db:"ip_address"`
	Mask           int    `json:"mask" db:"mask"`
}
