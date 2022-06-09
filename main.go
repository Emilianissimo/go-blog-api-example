package main

import (
	"os"
	"fmt"
	"log"
	"strconv"
	"net/http"
	"io/ioutil"
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
	"github.com/gorilla/mux"
)

// Global variables
var DB *sql.DB

// Structs

type Endpoints struct {
	Posts 	   string `json:"posts"`
	Categories string `json:"categories"`

}

type Post struct {
	Id         int `json:"id"`
	Title 	   string `json:"title"`
	Body  	   string `json:"body"`
	CategoryId int    `json:"category_id"`
	CreatedAt  string `json:"created_at"`
}

type Posts []Post

type RESTfulPost struct {
	Uri     string   `json:"uri"`
	Methods []string `json:"methods"`
	Data    Posts    `json:"data"`
}

type Category struct {
	Id         int `json:"id"`
	Title 	   string `json:"title"`
	Posts      Posts  `json:"posts"`
	CreatedAt  string `json:"created_at"`
}

type Categories []Category

type RESTfulCategory struct {
	Uri     string     `json:"uri"`
	Methods []string   `json:"methods"`
	Data    Categories `json:"data"`
}

// Endpoints

func index(page http.ResponseWriter, request *http.Request) {
	json.NewEncoder(page).Encode(Endpoints{
		Posts:      request.Host + "/api/posts/",
		Categories: request.Host + "/api/categories/",
	})
}

// Posts

func getAllPosts(page http.ResponseWriter, request *http.Request) {
	var posts = []Post{}

	row, err := DB.Query(`
		SELECT
			id,
			title,
			body,
			created_at,
			category_id
		FROM posts 
		ORDER BY created_at DESC
	`)
	if err != nil {
		panic(err)
	}

	for row.Next() {
		var post Post
		err = row.Scan(
			&post.Id,
			&post.Title,
			&post.Body,
			&post.CreatedAt,
			&post.CategoryId,
		)
		if err != nil {
			panic(err)	
		}
		posts = append(posts, post)
	}

	defer row.Close()

	json.NewEncoder(page).Encode(RESTfulPost{
		Uri:     request.Host + "/api/posts/",
		Methods: []string{"GET", "POST", "PATCH", "DELETE"},
		Data:    posts,
	})
}

func getOnePost(page http.ResponseWriter, request *http.Request) {
	postID := mux.Vars(request)["id"]
	row := DB.QueryRow(`
		SELECT
			id,
			title,
			body,
			created_at,
			category_id
		FROM posts 
		WHERE id=$1 
		LIMIT 1
	`, postID)

	var post Post 
	err := row.Scan(
		&post.Id,
		&post.Title,
		&post.Body,
		&post.CreatedAt,
		&post.CategoryId,
	)
	if err != nil {
		clientError(http.StatusNotFound, page)
		return
	}

	json.NewEncoder(page).Encode(post)
}

func createPost(page http.ResponseWriter, request *http.Request) {
	var newPost Post
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(requestBody, &newPost)

	if newPost.Title == "" || newPost.Body == "" || newPost.CategoryId == 0 {
		clientError(http.StatusUnprocessableEntity, page)
		return
	}

	SQL := `INSERT INTO posts(title, body, category_id) VALUES (?, ?, ?)`
	statement, err := DB.Prepare(SQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(newPost.Title, newPost.Body, newPost.CategoryId)
	if err != nil {
		log.Fatalln(err.Error())
	}

	page.WriteHeader(http.StatusCreated)

	json.NewEncoder(page)
}

func updatePost(page http.ResponseWriter, request *http.Request) {
	postID := mux.Vars(request)["id"]
	var newPost Post
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(requestBody, &newPost)

	if newPost.Title == "" && newPost.Body == "" && newPost.CategoryId == 0 {
		clientError(http.StatusUnprocessableEntity, page)
		return
	}

	if newPost.Title != "" {
		SQL := `UPDATE posts SET title=? WHERE id = ?`
		statement, err := DB.Prepare(SQL) // Prepare statement.
		// This is good to avoid SQL injections
		if err != nil {
			log.Fatalln(err.Error())
		}
		_, err = statement.Exec(newPost.Title, postID)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
	if newPost.Body != "" {
		SQL := `UPDATE posts SET body=? WHERE id = ?`
		statement, err := DB.Prepare(SQL) // Prepare statement.
		// This is good to avoid SQL injections
		if err != nil {
			log.Fatalln(err.Error())
		}
		_, err = statement.Exec(newPost.Body, postID)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
	if newPost.CategoryId != 0 {
		SQL := `UPDATE posts SET category_id=? WHERE id = ?`
		statement, err := DB.Prepare(SQL) // Prepare statement.
		// This is good to avoid SQL injections
		if err != nil {
			log.Fatalln(err.Error())
		}
		_, err = statement.Exec(newPost.CategoryId, postID)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
	

	page.WriteHeader(http.StatusNoContent)

	json.NewEncoder(page)
}

func deletePost(page http.ResponseWriter, request *http.Request) {
	postID := mux.Vars(request)["id"]
	SQL := `DELETE FROM posts WHERE id = ?`
	statement, err := DB.Prepare(SQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(postID)
	if err != nil {
		log.Fatalln(err.Error())
	}

	page.WriteHeader(http.StatusNoContent)

	json.NewEncoder(page)
}

// Categories

func getAllCategories(page http.ResponseWriter, request *http.Request) {
	var categories = []Category{}
	row, err := DB.Query(`
		SELECT
			categories.id as id,
			categories.title as title,
			categories.created_at as created_at
		FROM categories 
		ORDER BY created_at DESC
	`)
	if err != nil {
		panic(err)
	}

	for row.Next() {
		var category Category
		err = row.Scan(
			&category.Id,
			&category.Title,
			&category.CreatedAt,
		)
		post_row, err := DB.Query(`
			SELECT
				posts.id,
				posts.title,
				posts.body,
				posts.created_at,
				posts.category_id
			FROM posts
			WHERE posts.category_id = `+ strconv.Itoa(category.Id) +`
			ORDER BY posts.created_at DESC
		`)

		if err != nil {
			panic(err)
		}

		var category_posts = []Post{}

		for post_row.Next() {
			var post Post
			err = post_row.Scan(
				&post.Id,
				&post.Title,
				&post.Body,
				&post.CreatedAt,
				&post.CategoryId,
			)
			if err != nil {
				panic(err)	
			}
			category_posts = append(category_posts, post)
		}

		category.Posts = category_posts
		if err != nil {
			panic(err)	
		}
		categories = append(categories, category)
	}

	defer row.Close()

	json.NewEncoder(page).Encode(RESTfulCategory{
		Uri:     request.Host + "/api/categories/",
		Methods: []string{"GET", "POST", "PATCH", "DELETE"},
		Data:    categories,
	})
}

func getOneCategory(page http.ResponseWriter, request *http.Request) {
	categoryID := mux.Vars(request)["id"]
	row := DB.QueryRow(`
		SELECT
			categories.id as id,
			categories.title as title,
			categories.created_at as created_at
		FROM categories 
		WHERE categories.id = $1
		LIMIT 1
	`, categoryID)

	var category Category
	if err := row.Scan(
		&category.Id,
		&category.Title,
		&category.CreatedAt,
	);err != nil {
		clientError(http.StatusNotFound, page)
		return
	}

	post_row, err := DB.Query(`
		SELECT
			posts.id,
			posts.title,
			posts.body,
			posts.created_at,
			posts.category_id
		FROM posts
		WHERE posts.category_id = `+ strconv.Itoa(category.Id) +`
		ORDER BY posts.created_at DESC
	`)

	if err != nil {
		panic(err)
	}

	var category_posts = []Post{}

	for post_row.Next() {
		var post Post
		err = post_row.Scan(
			&post.Id,
			&post.Title,
			&post.Body,
			&post.CreatedAt,
			&post.CategoryId,
		)
		if err != nil {
			panic(err)	
		}
		category_posts = append(category_posts, post)
	}

	category.Posts = category_posts
	if err != nil {
		panic(err)	
	}

	json.NewEncoder(page).Encode(category)
}

func createCategory(page http.ResponseWriter, request *http.Request) {
	var newCategory Category
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(requestBody, &newCategory)

	if newCategory.Title == "" {
		clientError(http.StatusUnprocessableEntity, page)
		return
	}

	SQL := `INSERT INTO categories(title) VALUES ($1)`
	statement, err := DB.Prepare(SQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(newCategory.Title)
	if err != nil {
		log.Fatalln(err.Error())
	}

	page.WriteHeader(http.StatusCreated)

	json.NewEncoder(page)
}

func updateCategory(page http.ResponseWriter, request *http.Request) {
	categoryID := mux.Vars(request)["id"]
	var newCategory Category
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(requestBody, &newCategory)

	if newCategory.Title == "" {
		clientError(http.StatusUnprocessableEntity, page)
		return
	}

	SQL := `UPDATE categories SET title=? WHERE id = ?`
	statement, err := DB.Prepare(SQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(newCategory.Title, categoryID)
	if err != nil {
		log.Fatalln(err.Error())
	}

	page.WriteHeader(http.StatusNoContent)

	json.NewEncoder(page)
	
}

func deleteCategory(page http.ResponseWriter, request *http.Request) {
	categoryID := mux.Vars(request)["id"]
	SQL := `DELETE FROM categories WHERE id = ?`
	statement, err := DB.Prepare(SQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(categoryID)
	if err != nil {
		log.Fatalln(err.Error())
	}

	page.WriteHeader(http.StatusNoContent)

	json.NewEncoder(page)
}

// System functions

func clientError(status int, page http.ResponseWriter) {
	switch code := status; code {
	case http.StatusUnprocessableEntity:
		page.WriteHeader(http.StatusUnprocessableEntity)
		page.Header().Set("Content-Type", "application/json")
		response := make(map[string]string)
		response["message"] = "Unprocessable Entity"
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		page.Write(jsonResp)
		return
	case http.StatusNotFound:
		page.WriteHeader(http.StatusNotFound)
		page.Header().Set("Content-Type", "application/json")
		response := make(map[string]string)
		response["message"] = "Not found"
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		page.Write(jsonResp)
		return
	}
}

var router = mux.NewRouter().StrictSlash(true)

func dbConnect() {
	db, err := sql.Open("sqlite3", "./main.db")
	if err != nil {
		panic(err)
	}
	// Set one opened pool as global
	DB = db
	// defer db.Close()
}

func routes() {
	router.HandleFunc("/api/", index).Methods("GET")
	// Posts 
	router.HandleFunc("/api/posts/", getAllPosts).Methods("GET")
	router.HandleFunc("/api/posts/", createPost).Methods("POST")
	router.HandleFunc("/api/posts/{id}/", getOnePost).Methods("GET")
	router.HandleFunc("/api/posts/{id}/", updatePost).Methods("PATCH")
	router.HandleFunc("/api/posts/{id}/", deletePost).Methods("DELETE")
	// Categories
	router.HandleFunc("/api/categories/", getAllCategories).Methods("GET")
	router.HandleFunc("/api/categories/", createCategory).Methods("POST")
	router.HandleFunc("/api/categories/{id}/", getOneCategory).Methods("GET")
	router.HandleFunc("/api/categories/{id}/", updateCategory).Methods("PATCH")
	router.HandleFunc("/api/categories/{id}/", deleteCategory).Methods("DELETE")
}

func runServer() {
	fmt.Println("Server going on 127.0.0.1:8000")
	routes()
	dbConnect()
	log.Fatal(http.ListenAndServe(":8000", router))
}

func main() {
	if _, err := os.Stat("./main.db"); err == nil {
		runServer()
	} else {
		migrate()
		runServer()
	}
}


// Migrations
func migrate() {
	os.Remove("./main.db") // I delete the file to avoid duplicated records.
	// SQLite is a file based database.

	log.Println("Creating main.db...")
	file, err := os.Create("main.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("main.db created")

	sqliteDatabase, _ := sql.Open("sqlite3", "./main.db") // Open the created SQLite File
	defer sqliteDatabase.Close()                          // Defer Closing the database
	createTables(sqliteDatabase)                           // Create Database Tables
}

func createTables(db *sql.DB) {
	// Creating categories
	createCategories := `
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	` // SQL Statement for Create Table

	log.Println("Create categories table...")
	c_statement, c_err := db.Prepare(createCategories) // Prepare SQL Statement
	if c_err != nil {
		log.Fatal(c_err.Error())
	}
	c_statement.Exec() // Execute SQL Statements
	log.Println("Categories table created")

	// Creating posts
	createPosts := `
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
			title TEXT NOT NULL,
			body TEXT NOT NULL,
			category_id INTEGER NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE RESTRICT
		);
	` // SQL Statement for Create Table

	log.Println("Create licenses table...")
	p_statement, p_err := db.Prepare(createPosts) // Prepare SQL Statement
	if p_err != nil {
		log.Fatal(p_err.Error())
	}
	p_statement.Exec() // Execute SQL Statements
	log.Println("Posts table created")
}
