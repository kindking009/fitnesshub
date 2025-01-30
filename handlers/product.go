package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"fitnesshub/models"
)

func CreateProductHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var product models.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	_, err = collection.InsertOne(context.TODO(), product)
	if err != nil {
		http.Error(w, "Error adding product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Product added successfully"})
}

func GetProductByIDHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var product models.Product
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func UpdateProductByIDHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	var product models.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	if product.ID.IsZero() {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": product.ID}
	update := bson.M{"$set": product}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Error updating product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Product updated successfully"})
}

func DeleteProductByIDHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Product deleted successfully"})
}

func GetAllProductsHandler(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sort := r.URL.Query().Get("sort")
	filter := r.URL.Query().Get("filter")

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetSort(bson.D{{Key: sort, Value: 1}})

	var filterBson bson.M
	if filter != "" {
		filterBson = bson.M{"name": bson.M{"$regex": filter, "$options": "i"}}
	} else {
		filterBson = bson.M{}
	}

	var products []models.Product
	cursor, err := collection.Find(context.TODO(), filterBson, findOptions)
	if err != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var product models.Product
		cursor.Decode(&product)
		products = append(products, product)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func GetAllProducts(collection *mongo.Collection) []models.Product {
	var products []models.Product
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return products
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var product models.Product
		cursor.Decode(&product)
		products = append(products, product)
	}

	return products
}
