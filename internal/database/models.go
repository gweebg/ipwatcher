package database

import (
	"gorm.io/gorm"
)

type AddressEntry struct {
	gorm.Model

	// ID of the record, auto incremented uint64 value
	ID uint64 `gorm:"primaryKey;autoIncrement:true"`

	// Address is the newly fetched address that differs from the previous registered
	Address string `gorm:"index" json:"address"`
	// PreviousAddress is the previous address, before the update
	PreviousAddress string `json:"previous_address"`

	// Source is the URL of the API that provided the updated address
	Source string `json:"source"`
	// Version specifies the version of the address this record refers to
	Version string `json:"version"`
	// CreatedAt is the UNIX time when the address update was detected
	CreatedAt uint64 `gorm:"autoCreateTime" json:"at"`
}

// Create is the function that creates a new AddressEntry record onto the database
func (e AddressEntry) Create(address string) (*AddressEntry, error) {

	database := GetDatabase()
	entry := AddressEntry{Address: address}

	if err := database.Create(&entry).Error; err != nil {
		return nil, err
	}

	return &entry, nil
}

// First returns the latest added record for a specific address version (addressVersion)
func (e AddressEntry) First(addressVersion string) (*AddressEntry, error) {

	database := GetDatabase()

	var entry AddressEntry

	query := database.
		Where("version = ?", addressVersion).
		Order("id desc").
		First(&entry)

	if query.Error != nil {
		return nil, query.Error
	}

	return &entry, nil

}
