package models

type Role struct {
	Name        string   `bson:"name"`
	Permissions []string `bson:"permissions"`
}
