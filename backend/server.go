package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.staticPath, r.URL.Path)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

type Topic struct {
	Id          *int    `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type Category struct {
	Id          *int         `json:"id"`
	Name        *string      `json:"name"`
	TopicId     *int         `json:"topicId"`
	Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
	Id         *int    `json:"id"`
	Name       *string `json:"name"`
	Content    *string `json:"content"`
	CategoryId *int    `json:"categoryId"`
}

type TopicResponse struct {
	Topic      Topic      `json:"topic"`
	Categories []Category `json:"categories"`
}

var db *sql.DB

func main() {
	// read environment variables
	dbPath := os.Getenv("SUGGESTION_DB_PATH")
	staticDir := os.Getenv("SUGGESTION_FRONTEND_DIST_PATH")

	// initialize db
	dtb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	db = dtb
	defer db.Close()

	pragStmt := `PRAGMA foreign_keys = ON;`
	_, err = db.Exec(pragStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, pragStmt)
		return
	}

	sqlStmt := `
	create table if not exists topics (id integer not null primary key autoincrement, name text, description text);
	create table if not exists categories  (id integer not null primary key autoincrement, name text, topic_id integer, FOREIGN KEY(topic_id) REFERENCES topics(id) ON DELETE CASCADE);
	create table if not exists suggestions (id integer not null primary key autoincrement, name text, content text, category_id integer, FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE CASCADE);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	// initialize router
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api").Subrouter()
	// fetch topic
	apiRouter.HandleFunc("/topic", getTopics).Methods("GET")
	apiRouter.HandleFunc("/topic/{topicID}", getSuggestionsForTopic).Methods("GET")

	// post routes
	apiRouter.HandleFunc("/topic", createTopic).Methods("POST")
	apiRouter.HandleFunc("/category", createCategory).Methods("POST")
	apiRouter.HandleFunc("/suggestion", createSuggestion).Methods("POST")

	// delete routes
	apiRouter.HandleFunc("/topic/{topicID}", deleteTopic).Methods("DELETE")
	apiRouter.HandleFunc("/category/{categoryID}", deleteCategory).Methods("DELETE")
	apiRouter.HandleFunc("/suggestion/{suggestionID}", deleteSuggestion).Methods("DELETE")

	// path for static files
	spa := spaHandler{staticPath: staticDir, indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func getTopics(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("GET /api/topic")
	var topics []Topic
	rows, err := db.Query("select * from topics")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var topic Topic
		err = rows.Scan(&topic.Id, &topic.Name, &topic.Description)
		if err != nil {
			log.Fatal(err)
		}
		topics = append(topics, topic)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(topics)
}

func getSuggestionsForTopic(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	log.Default().Println("GET /api/topic/" + params["topicID"])
	var topic Topic
	var categories []Category
	id := params["topicID"]
	// get topic
	row := db.QueryRow("select * from topics where id = ?", id)
	err := row.Scan(&topic.Id, &topic.Name, &topic.Description)
	if err != nil {
		log.Fatal(err)
	}

	// get categories and suggestions
	rows, err := db.Query("select * from categories left join suggestions on suggestions.category_id = categories.id where topic_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var category Category
		category.Suggestions = []Suggestion{}
		var suggestion Suggestion
		err = rows.Scan(&category.Id, &category.Name, &category.TopicId, &suggestion.Id, &suggestion.Name, &suggestion.Content, &suggestion.CategoryId)
		if err != nil {
			log.Fatal(err)
		}
		var found = false
		for i, c := range categories {
			if *c.Id == *category.Id {
				if suggestion.Content != nil {
					categories[i].Suggestions = append(categories[i].Suggestions, suggestion)
				}
				log.Default().Println("Found category: " + *category.Name)
				found = true
				break
			}
		}
		if !found {
			if suggestion.Content != nil {
				category.Suggestions = append(category.Suggestions, suggestion)
			}
			log.Default().Println("Adding category: " + *category.Name)
			categories = append(categories, category)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	resp := TopicResponse{Topic: topic, Categories: categories}
	json.NewEncoder(w).Encode(resp)
}

func createTopic(w http.ResponseWriter, r *http.Request) {
	var newTopic Topic
	_ = json.NewDecoder(r.Body).Decode(&newTopic)
	// sqlite insert with brought in data
	log.Default().Println("POST /api/topic")
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Began transaction")
	stmt, err := tx.Prepare("insert into topics(name, description) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newTopic.Name, newTopic.Description)
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Inserted topic")
	row := tx.QueryRow("SELECT last_insert_rowid()")
	err = row.Scan(&newTopic.Id)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Committed transaction")
	json.NewEncoder(w).Encode(newTopic)
}

func createCategory(w http.ResponseWriter, r *http.Request) {
	var newCategory Category
	_ = json.NewDecoder(r.Body).Decode(&newCategory)
	// sqlite insert with brought in data
	log.Default().Println("POST /api/category")
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Began transaction")
	stmt, err := tx.Prepare("insert into categories(name, topic_id) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newCategory.Name, newCategory.TopicId)
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Inserted topic")
	row := tx.QueryRow("SELECT last_insert_rowid()")
	err = row.Scan(&newCategory.Id)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Committed transaction")
	json.NewEncoder(w).Encode(newCategory)
}

func createSuggestion(w http.ResponseWriter, r *http.Request) {
	var newSuggestion Suggestion
	_ = json.NewDecoder(r.Body).Decode(&newSuggestion)
	// sqlite insert with brought in data
	log.Default().Println("POST /api/suggestion")
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Began transaction")
	stmt, err := tx.Prepare("insert into suggestions(name, content, category_id) values(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newSuggestion.Name, newSuggestion.Content, newSuggestion.CategoryId)
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Inserted topic")
	row := tx.QueryRow("SELECT last_insert_rowid()")
	err = row.Scan(&newSuggestion.Id)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Committed transaction")
	json.NewEncoder(w).Encode(newSuggestion)
}

func deleteTopic(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	topicID, err := strconv.Atoi(params["topicID"])
	if err != nil {
		log.Fatal(err)
		return
	}
	// sqlite insert with brought in data
	log.Default().Println("DELETE /api/topic/" + params["topicID"])
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Began transaction")
	stmt, err := tx.Prepare("delete from topics where id = (?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(topicID)
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("delete topic")
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Committed transaction")
	json.NewEncoder(w).Encode(topicID)
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	categoryID, err := strconv.Atoi(params["categoryID"])
	if err != nil {
		log.Fatal(err)
		return
	}
	// sqlite insert with brought in data
	log.Default().Println("DELETE /api/category/" + params["categoryID"])
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Began transaction")
	stmt, err := tx.Prepare("delete from categories where id = (?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(categoryID)
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("delete category")
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Committed transaction")
	json.NewEncoder(w).Encode(categoryID)
}

func deleteSuggestion(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	suggestionID, err := strconv.Atoi(params["suggestionID"])
	if err != nil {
		log.Fatal(err)
		return
	}
	// sqlite insert with brought in data
	log.Default().Println("DELETE /api/suggestion/" + params["suggestionID"])
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Began transaction")
	stmt, err := tx.Prepare("delete from suggestions where id = (?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(suggestionID)
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("delete suggestion")
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Println("Committed transaction")
	json.NewEncoder(w).Encode(suggestionID)
}
