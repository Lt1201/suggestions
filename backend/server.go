package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	Name        string `json:"name"`
	Description string `json:"description"`
	Id          int    `json:"id"`
}

type Suggestion struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Id      int    `json:"id"`
	TopicId int    `json:"topicId"`
}

var suggestions []Suggestion
var db *sql.DB

func main() {
	// initialize db
	dtb, err := sql.Open("sqlite3", "../build/suggestions.db")
	if err != nil {
		log.Fatal(err)
	}
	db = dtb
	defer db.Close()

	sqlStmt := `
	create table if not exists topics (id integer not null primary key autoincrement, name text, description text);
	create table if not exists categories  (id integer not null primary key autoincrement, name text, topic_id integer, FOREIGN KEY(topic_id) REFERENCES topics(id));
	create table if not exists suggestions (id integer not null primary key autoincrement, name text, content text, category_id integer, FOREIGN KEY(category_id) REFERENCES categories(id));`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	// initialize router
	router := mux.NewRouter()

	var newSuggestion Suggestion = Suggestion{
		Name:    "Test Name",
		Content: "Test Content",
		Id:      1,
		TopicId: 1,
	}
	suggestions = append(suggestions, newSuggestion)

	apiRouter := router.PathPrefix("/api").Subrouter()
	// topic related routes
	apiRouter.HandleFunc("/topic", getTopics).Methods("GET")
	apiRouter.HandleFunc("/topic/{topicName}", getSuggestionsForTopic).Methods("GET")
	apiRouter.HandleFunc("/topic", createTopic).Methods("POST")

	// path for static files
	spa := spaHandler{staticPath: "../frontend/dist/frontend/browser", indexPath: "index.html"}
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
	log.Default().Println("GET /api/topic/" + params["topicName"])
	json.NewEncoder(w).Encode(params["topicName"])
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
