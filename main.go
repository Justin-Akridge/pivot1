package main

import (
  "log"
  "html/template"
  "net/http"
  "os"
  "fmt"
  "github.com/joho/godotenv"
  _ "github.com/lib/pq"
  "database/sql"
  "github.com/gorilla/sessions"
  "github.com/gorilla/handlers"
  "github.com/gorilla/mux"
  "github.com/pivot/handler"
  "github.com/pivot/migrations"
)

var db *sql.DB
var templates *template.Template
var store *sessions.CookieStore

func main() {
  r := mux.NewRouter()

  err := godotenv.Load()
  if err != nil {
      log.Fatal("Error loading .env file")
  }

  sessionKey := os.Getenv("SESSION_KEY")
  if sessionKey == "" {
    log.Fatal("Session key is missing")
  }

  store = sessions.NewCookieStore([]byte(sessionKey))

  // apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
  // if apiKey == "" {
  //     log.Fatal("GOOGLE_MAPS_API_KEY environment variable is not set")
  // }

  // Load templates with API key
  // templates = template.Must(template.New("").Funcs(template.FuncMap{
  //     "apiKey": func() string { return apiKey },
  // }).ParseGlob(filepath.Join("static", "*.html")))

  r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

  // Database connection
  dbUser := os.Getenv("DB_USER")
  dbPassword := os.Getenv("DB_PASSWORD")
  dbName := os.Getenv("DB_NAME")
  dbHost := os.Getenv("DB_HOST")
  dbPort := os.Getenv("DB_PORT")
  connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
  db, err = sql.Open("postgres", connStr)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  err = db.Ping()
  if err != nil {
    log.Fatal("Cannot connect to the database:", err)
  }

  if err := migrations.RunMigrations(db); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
  }

  corsHandler := handlers.CORS(
    handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
    handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
    handlers.AllowedOrigins([]string{"*"}),
    handlers.AllowCredentials(),
  )



// unprotected routes
r.HandleFunc("/login", handler.HandleLogin(store, templates, db)).Methods("GET", "POST")
r.HandleFunc("/signup", handler.HandleSignup(store, templates, db)).Methods("GET", "POST")
r.HandleFunc("/logout", handler.HandleLogout(store)).Methods("GET")
r.HandleFunc("/contact", handler.HandleGetContactPage(templates)).Methods("GET")

// protected routes
// GET
r.Handle("/map", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleGetMap(templates)))))
r.Handle("/map/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleGetMapWithJob(templates, db))))).Methods("GET")
r.Handle("/createJob", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleCreateJob(store, db))))).Methods("POST")
r.Handle("/jobs", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleGetJobs(store, db, templates))))).Methods("GET")
r.Handle("/uploadLas/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleUploadLas(store, db))))).Methods("POST")
r.Handle("/replaceLas/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleReplaceLas(store, db))))).Methods("POST")
r.Handle("/poleLocations/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleGetPoleLocations(store, db))))).Methods("GET")
r.Handle("/savePaths/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handler.HandleSavePathsOfPoleLine(store, db))))).Methods("POST")

log.Println("Server started at http://localhost:8080")
log.Fatal(http.ListenAndServe(":8080", r))
  // corsHandler := handlers.CORS( handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	// 	handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
	// 	handlers.AllowedOrigins([]string{"*"}),
	// 	handlers.AllowCredentials(),
	// )

  // // unprotected routes
  // r.HandleFunc("/login", handlers.HandleLogin(store, templates, db)).Methods("GET", "POST")
  // r.HandleFunc("/signup", handlers.HandleSignup(store, templates, db)).Methods("GET", "POST")
  // r.HandleFunc("/logout", handlers.HandleLogout(store)).Methods("GET")
  // r.HandleFunc("/contact", handlers.HandleGetContactPage(templates)).Methods("GET")

  // // protected routes
  // // GET
  // r.Handle("/map", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetMap(templates))))).Methods("GET")
	// r.Handle("/map/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetMapWithJob(templates, db))))).Methods("GET")
	// r.Handle("/createJob", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleCreateJob(store, db))))).Methods("POST")
	// r.Handle("/jobs", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetJobs(store, db, templates))))).Methods("GET")
	// r.Handle("/uploadLas/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleUploadLas(store, db))))).Methods("POST")
	// r.Handle("/replaceLas/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleReplaceLas(store, db))))).Methods("POST")
	// r.Handle("/poleLocations/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetPoleLocations(store, db))))).Methods("GET")
	// r.Handle("/savePaths/{id}", corsHandler(authenticationMiddleware(store, http.HandlerFunc(handlers.HandleSavePathsOfPoleLine(store, db))))).Methods("POST")

  // // r.Handle("/map", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetMap(templates)))).Methods("GET")
  // // r.Handle("/map/{id}", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetMapWithJob(templates, db)))).Methods("GET")
  // // r.Handle("/createJob", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleCreateJob(store, db)))).Methods("POST")
  // // r.Handle("/jobs", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetJobs(store, db, templates)))).Methods("GET")
  // // r.Handle("/uploadLas/{id}", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleUploadLas(store, db)))).Methods("POST")
  // // r.Handle("/replaceLas/{id}", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleReplaceLas(store, db)))).Methods("POST")
  // // r.Handle("/poleLocations/{id}", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetPoleLocations(store, db)))).Methods("GET")
  // //r.Handle("/vegetationEncroachments/${id}", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleGetVegetationEncroachments(store,db)))).Methods("GET")

  // r.Handle("/savePaths/{id}", authenticationMiddleware(store, http.HandlerFunc(handlers.HandleSavePathsOfPoleLine(store, db)))).Methods("POST")


  // log.Println("Server started at http://localhost:8080/map")
  // log.Fatal(http.ListenAndServe(":8080", r))
}

func authenticationMiddleware(store *sessions.CookieStore, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // prevent caching of sensitive pages
    w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

    session, err := store.Get(r, "SESSION_KEY")
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    auth, ok := session.Values["authenticated"].(bool);
    if !ok || !auth {
      http.Redirect(w, r, "/login", http.StatusSeeOther)
      return
    }

    next.ServeHTTP(w, r)
  })
}

func createNewUser(email, password string) error {
  query := `INSERT INTO users (id, email, password, created_at)
            VALUES (gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP)`
  _, err := db.Exec(query)
	if err != nil {
		log.Printf("Error inserting user into database: %v", err)
    return err
	}
  return nil
}
