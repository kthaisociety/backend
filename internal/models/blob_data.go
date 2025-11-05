package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlobData struct {
	gorm.Model
	UserId uuid.UUID `gorm:"uniqueIndex" json:"user_id"`
	Uri    string    `json:"uri"`
}

func (bd BlobData) getData() ([]byte, error) {
	///fetch bd.uri from bucket
	//return
	return []byte{}, nil
}

func NewBlobData(Blob []byte, db gorm.DB) (BlobData, error) {
	// push data to R2
	// create struct
	id, _ := uuid.NewUUID()
	bd := BlobData{
		UserId: id,
		Uri:    "asdfasdf",
	}
	// push to postgresql
	if err := db.Create(&bd).Error; err != nil {
		return bd, err
	}
	return bd, nil
}
