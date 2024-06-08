package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Структура на книгата
type Book struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ISBN        string `json:"isbn"`
	Author      string `json:"author"`
	PublishedAt int    `json:"published_at"`
}

var (
	books  = make(map[int]Book)
	nextID = 1
	mu     sync.Mutex
)

// Функция за обработка на заявките към /books
func handleBooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getBooks(w, r)
	case http.MethodPost:
		createBook(w, r)
	default:
		http.Error(w, "Методът не е позволен", http.StatusMethodNotAllowed)
	}
}

// Функция за обработка на заявките към /books/{id}
func handleBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/books/"):]
	id, _ := strconv.Atoi(idStr)

	switch r.Method {
	case http.MethodGet:
		getBook(w, r, id)
	case http.MethodPut:
		updateBook(w, r, id)
	case http.MethodDelete:
		deleteBook(w, r, id)
	default:
		http.Error(w, "Методът не е позволен", http.StatusMethodNotAllowed)
	}
}

// Функция за връщане на всички книги
func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		limitInt, _ := strconv.Atoi(limit)
		var limitedBooks []Book
		count := 0
		for _, book := range books {
			if count >= limitInt {
				break
			}
			limitedBooks = append(limitedBooks, book)
			count++
		}
		json.NewEncoder(w).Encode(limitedBooks)
		return
	}
	var allBooks []Book
	for _, book := range books {
		allBooks = append(allBooks, book)
	}
	json.NewEncoder(w).Encode(allBooks)
}

// Функция за връщане на конкретна книга по ID
func getBook(w http.ResponseWriter, r *http.Request, id int) {
	book, exists := books[id]
	if !exists {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// Функция за създаване на нова книга
func createBook(w http.ResponseWriter, r *http.Request) {
	var newBook Book
	json.NewDecoder(r.Body).Decode(&newBook)
	mu.Lock()
	newBook.ID = nextID
	nextID++
	books[newBook.ID] = newBook
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newBook)
}

// Функция за актуализиране на съществуваща книга
func updateBook(w http.ResponseWriter, r *http.Request, id int) {
	var updatedBook Book
	json.NewDecoder(r.Body).Decode(&updatedBook)
	mu.Lock()
	if _, exists := books[id]; !exists {
		mu.Unlock()
		http.NotFound(w, r)
		return
	}
	updatedBook.ID = id
	books[id] = updatedBook
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedBook)
}

// Функция за изтриване на книга по ID
func deleteBook(w http.ResponseWriter, r *http.Request, id int) {
	mu.Lock()
	if _, exists := books[id]; !exists {
		mu.Unlock()
		http.NotFound(w, r)
		return
	}
	delete(books, id)
	mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/books", handleBooks) // GET /books, POST /books
	mux.HandleFunc("/books/", handleBook) // GET /books/{id}, PUT /books/{id}, DELETE /books/{id}

	fmt.Println("Сървърът е инициализиран успешно на порт 8080...")
	http.ListenAndServe(":8080", mux)
}
