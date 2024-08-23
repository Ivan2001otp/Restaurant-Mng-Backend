package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type User struct{
	ID				primitive.ObjectID 	 	`bson:"_id"`
	First_name		*string					`json:"first_name" validate:"required,min=2,max=100"`
	Last_name		*string					`json:"last_name" validate:"required,min=2,max=100"`
	Email			*string					`json:"password" validate:"required,min=6"`
	Password		*string					`json:"email" validate:"email,required"`
	Avatar			*string					`json:"avatar"`
	Token			*string					`json:"token"`
	Phone			*string					`json:"phone" validate:"required"`
	Refresh_token	*string					`json:"refresh_token"`
	Created_at		time.Time				`json:"created_at"`
	Updated_at		time.Time				`json:"updated_at"`
	User_id			string					`json:"user_id"`
}