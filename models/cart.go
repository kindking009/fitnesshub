package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Cart struct {
	UserID primitive.ObjectID `bson:"user_id"`
	Items  []CartItem         `bson:"items"`
}

type CartItem struct {
	ProductID primitive.ObjectID `bson:"product_id"`
	Quantity  int                `bson:"quantity"`
}
