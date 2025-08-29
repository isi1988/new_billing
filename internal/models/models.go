package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// CustomDate handles dates in YYYY-MM-DD format from JSON
type CustomDate struct {
	time.Time
}

// UnmarshalJSON handles JSON unmarshaling for date-only format
func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	// Remove quotes from JSON string
	dateStr := string(data[1 : len(data)-1])
	if dateStr == "" {
		return nil
	}

	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err
	}

	cd.Time = parsedTime
	return nil
}

// MarshalJSON handles JSON marshaling for date-only format
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, cd.Time.Format("2006-01-02"))), nil
}

// Value implements driver.Valuer interface for database storage
func (cd CustomDate) Value() (driver.Value, error) {
	if cd.Time.IsZero() {
		return nil, nil
	}
	return cd.Time, nil
}

// Scan implements sql.Scanner interface for database retrieval
func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		cd.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		cd.Time = v
		return nil
	default:
		return fmt.Errorf("cannot scan %T into CustomDate", value)
	}
}

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
	ID                  int         `json:"id" db:"id"`
	ClientType          ClientType  `json:"client_type" db:"client_type"`
	Email               string      `json:"email" db:"email"`
	Phone               string      `json:"phone" db:"phone"`
	IsBlocked           bool        `json:"is_blocked" db:"is_blocked"`
	FirstName           *string     `json:"first_name,omitempty" db:"first_name"`
	LastName            *string     `json:"last_name,omitempty" db:"last_name"`
	Patronymic          *string     `json:"patronymic,omitempty" db:"patronymic"`
	PassportNumber      *string     `json:"passport_number,omitempty" db:"passport_number"`
	PassportIssuedBy    *string     `json:"passport_issued_by,omitempty" db:"passport_issued_by"`
	PassportIssueDate   *CustomDate `json:"passport_issue_date,omitempty" db:"passport_issue_date"`
	RegistrationAddress *string     `json:"registration_address,omitempty" db:"registration_address"`
	BirthDate           *CustomDate `json:"birth_date,omitempty" db:"birth_date"`
	INN                 *string     `json:"inn,omitempty" db:"inn"`
	KPP                 *string     `json:"kpp,omitempty" db:"kpp"`
	FullName            *string     `json:"full_name,omitempty" db:"full_name"`
	ShortName           *string     `json:"short_name,omitempty" db:"short_name"`
	OGRN                *string     `json:"ogrn,omitempty" db:"ogrn"`
	OGRNDate            *CustomDate `json:"ogrn_date,omitempty" db:"ogrn_date"`
	LegalAddress        *string     `json:"legal_address,omitempty" db:"legal_address"`
	ActualAddress       *string     `json:"actual_address,omitempty" db:"actual_address"`
	BankName            *string     `json:"bank_name,omitempty" db:"bank_name"`
	BankAccount         *string     `json:"bank_account,omitempty" db:"bank_account"`
	BankBIK             *string     `json:"bank_bik,omitempty" db:"bank_bik"`
	BankCorrespondent   *string     `json:"bank_correspondent,omitempty" db:"bank_correspondent"`
	CEO                 *string     `json:"ceo,omitempty" db:"ceo"`
	Accountant          *string     `json:"accountant,omitempty" db:"accountant"`
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
	ID        int        `json:"id" db:"id"`
	ClientID  int        `json:"client_id" db:"client_id"`
	Number    string     `json:"number" db:"number"`
	SignDate  CustomDate `json:"sign_date" db:"sign_date"`
	IsBlocked bool       `json:"is_blocked" db:"is_blocked"`
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
	IsBlocked      bool   `json:"is_blocked" db:"is_blocked"`
}

type Traffic struct {
	ID           int       `json:"id" db:"id"`
	ConnectionID int       `json:"connection_id" db:"connection_id"`
	ClientID     int       `json:"client_id" db:"client_id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	BytesIn      int64     `json:"bytes_in" db:"bytes_in"`
	BytesOut     int64     `json:"bytes_out" db:"bytes_out"`
	PacketsIn    int64     `json:"packets_in" db:"packets_in"`
	PacketsOut   int64     `json:"packets_out" db:"packets_out"`
}

type TrafficResponse struct {
	ID           int       `json:"id" db:"id"`
	ConnectionID int       `json:"connection_id" db:"connection_id"`
	ClientID     int       `json:"client_id" db:"client_id"`
	ClientName   string    `json:"client_name" db:"client_name"`
	ClientEmail  string    `json:"client_email" db:"client_email"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	BytesIn      int64     `json:"bytes_in" db:"bytes_in"`
	BytesOut     int64     `json:"bytes_out" db:"bytes_out"`
	PacketsIn    int64     `json:"packets_in" db:"packets_in"`
	PacketsOut   int64     `json:"packets_out" db:"packets_out"`
	TotalTraffic int64     `json:"total_traffic" db:"total_traffic"`
}

// --- Issues (Доработки) ---
type IssueStatus string

const (
	NewIssue      IssueStatus = "new"
	ResolvedIssue IssueStatus = "resolved"
)

type Issue struct {
	ID          int         `json:"id" db:"id"`
	Title       string      `json:"title" db:"title"`
	Description string      `json:"description" db:"description"`
	Status      IssueStatus `json:"status" db:"status"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedBy   int         `json:"created_by" db:"created_by"`
	ResolvedBy  *int        `json:"resolved_by,omitempty" db:"resolved_by"`
}

type IssueHistory struct {
	ID        int       `json:"id" db:"id"`
	IssueID   int       `json:"issue_id" db:"issue_id"`
	FieldName string    `json:"field_name" db:"field_name"`
	OldValue  *string   `json:"old_value" db:"old_value"`
	NewValue  *string   `json:"new_value" db:"new_value"`
	EditedBy  int       `json:"edited_by" db:"edited_by"`
	EditedAt  time.Time `json:"edited_at" db:"edited_at"`
}

type IssueWithHistory struct {
	Issue
	History []IssueHistory `json:"history,omitempty"`
}
