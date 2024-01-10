package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mime/multipart"
)

type Beer struct {
	ID        primitive.ObjectID    `bson:"_id,omitempty"`
	Name      string                `form:"name" binding:"required"`
	Category  string                `form:"category" binding:"required"`
	Detail    string                `form:"detail" binding:"required"`
	Image     *multipart.FileHeader `form:"image"`
	ImagePath string
	Deleted   bool `bson:"deleted,omitempty"` // New field indicating soft delete
}

type BeerPagingResult struct {
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
	PrevPage  int     `json:"prevPage"`
	NextPage  int     `json:"nextPage"`
	Count     int     `json:"count"`
	TotalPage int     `json:"totalPage"`
	Beer      []*Beer `json:"beer"`
}
