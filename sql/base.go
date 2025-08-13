package sql

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"index;not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"index;not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate hook to ensure ID is set
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// IsDeleted checks if the model is soft deleted
func (b BaseModel) IsDeleted() bool {
	return b.DeletedAt != nil
}

// ID represents a UUID identifier
type ID struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
}

// BeforeCreate hook to ensure ID is set
func (i *ID) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// Timestamp provides creation and update timestamps
type Timestamp struct {
	CreatedAt time.Time `gorm:"index;not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"index;not null" json:"updatedAt"`
}

// SoftDelete provides soft delete functionality
type SoftDelete struct {
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// IsDeleted checks if the model is soft deleted
func (s SoftDelete) IsDeleted() bool {
	return s.DeletedAt != nil
}

// Delete marks the model as soft deleted
func (s *SoftDelete) Delete() {
	now := time.Now()
	s.DeletedAt = &now
}

// Restore removes the soft delete mark
func (s *SoftDelete) Restore() {
	s.DeletedAt = nil
}

// NullableUUID provides a nullable UUID type
type NullableUUID struct {
	UUID  uuid.UUID
	Valid bool
}

// Scan implements the Scanner interface
func (nu *NullableUUID) Scan(value interface{}) error {
	if value == nil {
		nu.UUID, nu.Valid = uuid.Nil, false
		return nil
	}
	nu.Valid = true
	switch v := value.(type) {
	case string:
		return nu.UUID.UnmarshalText([]byte(v))
	case []byte:
		return nu.UUID.UnmarshalText(v)
	default:
		return nu.UUID.Scan(value)
	}
}

// Value implements the driver Valuer interface
func (nu NullableUUID) Value() (driver.Value, error) {
	if !nu.Valid {
		return nil, nil
	}
	return nu.UUID.String(), nil
}
