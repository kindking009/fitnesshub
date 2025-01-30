package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mail.v2"

	"fitnesshub/models"
	"fitnesshub/utils"
)

func SignUpHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)
	user.Verified = false
	user.Role = "user" // По умолчанию обычный пользователь

	// Генерируем верификационный токен
	verificationToken, err := utils.GenerateVerificationToken()
	if err != nil {
		http.Error(w, "Error generating verification token", http.StatusInternalServerError)
		return
	}
	user.VerificationToken = verificationToken

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "Error adding user", http.StatusInternalServerError)
		return
	}

	// Ссылка на подтверждение
	verificationLink := "http://localhost:8081/verify?token=" + url.QueryEscape(verificationToken)

	// Отправляем email
	m := mail.NewMessage()
	m.SetHeader("From", "no-reply@fitnesshub.com")
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "Verify your email")
	m.SetBody("text/plain", "Please verify your email by clicking the following link: "+verificationLink)

	d := mail.NewDialer("smtp.mailtrap.io", 2525, "9c521257de733d", "ae506c9d02f243")
	err = d.DialAndSend(m)
	if err != nil {
		http.Error(w, "Error sending verification email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "User registered successfully. Check your email for verification."})
}

func VerifyEmailHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification token is required", http.StatusBadRequest)
		return
	}

	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"verification_token": token}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid or expired verification token", http.StatusBadRequest)
		return
	}

	filter := bson.M{"verification_token": token}
	update := bson.M{"$set": bson.M{"verified": true}}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Error verifying email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Email verified successfully"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	var user models.User
	err = collection.FindOne(context.TODO(), bson.M{"email": credentials.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if !user.Verified {
		http.Error(w, "Email not verified", http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "token": token})
}
