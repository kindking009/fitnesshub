package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"fitnesshub/db"
	"fitnesshub/handlers"
	"fitnesshub/middleware"
)

func main() {
	// Подключение к MongoDB
	client, err := db.ConnectToMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// Инициализация коллекций
	userCollection := client.Database("fitnesshub").Collection("users")
	productCollection := client.Database("fitnesshub").Collection("products")

	// Регистрация обработчиков для аутентификации
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.ServeFile(w, r, "templates/signup.html")
		} else if r.Method == "POST" {
			handlers.SignUpHandler(w, r, userCollection)
		}
	})

	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		handlers.VerifyEmailHandler(w, r, userCollection)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.ServeFile(w, r, "templates/login.html")
		} else if r.Method == "POST" {
			handlers.LoginHandler(w, r, userCollection)
		}
	})

	// Регистрация обработчиков для административной панели
	adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.ServeFile(w, r, "templates/admin.html")
		}
	})
	http.Handle("/admin", middleware.RoleBasedAccessControl(adminHandler, "administrator"))

	adminUsersHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			users := handlers.GetAllUsers(userCollection)
			tmpl, err := template.ParseFiles("templates/admin_users.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, users)
		} else if r.Method == "POST" {
			handlers.AdminCreateUserHandler(w, r, userCollection)
		} else if r.Method == "DELETE" {
			handlers.AdminDeleteUserByIDHandler(w, r, userCollection)
		}
	})
	http.Handle("/admin/users", middleware.RoleBasedAccessControl(adminUsersHandler, "administrator"))

	adminProductsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			products := handlers.GetAllProducts(productCollection)
			tmpl, err := template.ParseFiles("templates/admin_products.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, products)
		} else if r.Method == "POST" {
			handlers.CreateProductHandler(w, r, productCollection)
		} else if r.Method == "DELETE" {
			handlers.DeleteProductByIDHandler(w, r, productCollection)
		}
	})
	http.Handle("/admin/products", middleware.RoleBasedAccessControl(adminProductsHandler, "administrator"))

	// Регистрация обработчиков для продуктов
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handlers.CreateProductHandler(w, r, productCollection)
		case "GET":
			if r.URL.Query().Get("id") != "" {
				handlers.GetProductByIDHandler(w, r, productCollection)
			} else {
				handlers.GetAllProductsHandler(w, r, productCollection)
			}
		case "PUT":
			handlers.UpdateProductByIDHandler(w, r, productCollection)
		case "DELETE":
			handlers.DeleteProductByIDHandler(w, r, productCollection)
		default:
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	})

	// Регистрация обработчиков для пользовательского профиля
	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			handlers.UpdateUserProfileHandler(w, r, userCollection)
		case "POST":
			handlers.ChangeUserPasswordHandler(w, r, userCollection)
		default:
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	})

	// Обслуживание статических файлов
	fs := http.FileServer(http.Dir("./templates"))
	http.Handle("/", fs)

	// Запуск сервера
	log.Println("Сервер запущен на порту 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))

}
