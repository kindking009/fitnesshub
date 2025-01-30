package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"fitnesshub/models"
)

func UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
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
		http.Error(w, "Error updating user profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "User profile updated successfully"})
}

func ChangeUserPasswordHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var credentials struct {
		UserID      primitive.ObjectID `json:"user_id"`
		OldPassword string             `json:"old_password"`
		NewPassword string             `json:"new_password"`
	}
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	var user models.User
	err = collection.FindOne(context.TODO(), bson.M{"_id": credentials.UserID}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.OldPassword))
	if err != nil {
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing new password", http.StatusInternalServerError)
		return
	}

	filter := bson.M{"_id": credentials.UserID}
	update := bson.M{"$set": bson.M{"password": string(hashedPassword)}}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Password changed successfully"})
}
