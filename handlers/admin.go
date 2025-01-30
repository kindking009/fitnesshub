package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"fitnesshub/models"
)

func AdminGetAllUsersHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var users []models.User
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var user models.User
		cursor.Decode(&user)
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func AdminDeleteUserByIDHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "User deleted successfully"})
}

func AdminCreateUserHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "Error adding user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "User added successfully"})
}

func AdminUpdateUserByIDHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	if user.ID.IsZero() {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "User updated successfully"})
}

func GetAllUsers(collection *mongo.Collection) []models.User {
	var users []models.User
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return users
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var user models.User
		cursor.Decode(&user)
		users = append(users, user)
	}

	return users
}
