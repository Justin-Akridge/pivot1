package handler

import (
  "os/signal"
	"syscall"
  "bytes"
	"io/ioutil"
	"database/sql"
  "encoding/json"
  "html/template"
	"fmt"
  "os"
  "os/exec"
  "io"
  "path/filepath"
	"net/http"
  "time"
  "log"
	"github.com/gorilla/sessions"
  "github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
  _ "github.com/lib/pq"
)

type Job struct {
	ID          string    `json:"id"`
	JobName     string    `json:"job_name"`
	CompanyName string    `json:"company_name"`
	UserID      string    `json:"user_id"`
	CompanyID   string    `json:"company_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func HandleSignup(store *sessions.CookieStore, templates *template.Template, db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
      if store == nil || templates == nil || db == nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
      }

    if r.Method == "GET" {
      err := templates.ExecuteTemplate(w, "signup.html", nil)
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    } else if r.Method == "POST" {
      err := r.ParseForm()
      if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
      }
      
      email := r.FormValue("email")
      password := r.FormValue("password")
      confirmPassword := r.FormValue("confirm_password")
      
      if confirmPassword != password {
        renderSignupTemplateWithError(w, "Passwords do not match", templates)
      }

      hashedPassword, err := hashPassword(password)
      if err != nil {
        http.Error(w, "Unable to hash password", http.StatusInternalServerError)
        return
      }

      query := `INSERT INTO admins (id, email, password, created_at)
                VALUES (gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP)
                RETURNING id
               `
      var id string
      err = db.QueryRow(query, email, hashedPassword).Scan(&id)


      if err != nil {
        renderSignupTemplateWithError(w, "Error creating account. Please try again", templates)
        return
      }

      // create a new session
      session, err := store.Get(r, "SESSION_KEY")
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }

      //set session values
      session.Values["authenticated"] = true
      session.Values["user_email"] = email
      session.Values["user_id"] = id 
      session.Values["admin"] = true

      err = session.Save(r, w)
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }

      http.Redirect(w, r, "/map", http.StatusSeeOther)
    }
  }
}


func HandleLogin(store *sessions.CookieStore, templates *template.Template, db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if store == nil || templates == nil || db == nil {
      http.Error(w, "Internal server error", http.StatusInternalServerError)
      return
    }

    // Check if user session exists
    session, err := store.Get(r, "SESSION_KEY")
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    if auth, ok := session.Values["authenticated"].(bool); ok && auth {
      http.Redirect(w, r, "/map", http.StatusSeeOther)
      return
    }
    
    if r.Method == "GET" {
      err := templates.ExecuteTemplate(w, "login.html", nil)
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    } else if r.Method == "POST" {
      err := r.ParseForm()
      if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
      }

      email := r.FormValue("email")
      password := r.FormValue("password")

      var userId, storedPassword string
      var isAdmin bool
      // check if the user is a admin or not
      err = db.QueryRow("SELECT id, password FROM users WHERE email = $1", email).Scan(&userId, &storedPassword)
			if err != nil {
				// If user not found in users table, check admins table
				err = db.QueryRow("SELECT id, password FROM admins WHERE email = $1", email).Scan(&userId, &storedPassword)
				if err != nil {
					if err == sql.ErrNoRows {
						renderLoginTemplateWithError(w, "Incorrect email or password", templates)
					} else {
						http.Error(w, "Database error", http.StatusInternalServerError)
					}
					return
				}
				// User is found in admins table, set isAdmin to true
				isAdmin = true
			}

      if !checkPasswordHash(password, storedPassword) {
        renderLoginTemplateWithError(w, "Error creating account. Please try again", templates)
        return
      }
      
      // create a new session
      session, err := store.Get(r, "SESSION_KEY")
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }

      // check cookies match
      //set session values
      session.Values["authenticated"] = true
      session.Values["user_email"] = email
      session.Values["user_id"] = userId
      if isAdmin {
        session.Values["admin"] = true
      } else {
        session.Values["admin"] = false
      }

      err = session.Save(r, w)
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }

      http.Redirect(w, r, "/map", http.StatusSeeOther)
    }
  }
}

type loginSuccess struct {
	Success bool
	Error string
}

func renderLoginTemplateWithError(w http.ResponseWriter, errorMessage string, templates *template.Template) {
  data := loginSuccess{
    Success: false,
    Error:   errorMessage,
  }
  err := templates.ExecuteTemplate(w, "login.html", data)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func renderSignupTemplateWithError(w http.ResponseWriter, errorMessage string, templates *template.Template) {
  data := loginSuccess{
    Success: false,
    Error:   errorMessage,
  }
  err := templates.ExecuteTemplate(w, "signup.html", data)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func HandleGetJobs(store *sessions.CookieStore, db *sql.DB, templates *template.Template) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    // Check if user session exists
    fmt.Println("HERE")
    session, err := store.Get(r, "SESSION_KEY")
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    userId, ok := session.Values["user_id"].(string)
    if !ok {
        http.Error(w, "User ID not found in context", http.StatusUnauthorized)
        return
    }

    isAdmin, ok := session.Values["admin"].(bool)
    if !ok {
        http.Error(w, "Admin status not found in context", http.StatusUnauthorized)
        return
    }

    var adminId string

    if isAdmin {
      adminId = userId
    } else {
      err := db.QueryRow(`SELECT admin_id FROM users WHERE id = $1`, userId).Scan(&adminId)
      if err != nil {
        http.Error(w, "Error fetching admin id from database", http.StatusInternalServerError)
        return
      }
    }

    query := `SELECT id, job_name, company_name, created_at FROM jobs WHERE admin_id = $1`
    rows, err := db.Query(query, adminId)
    if err != nil {
      http.Error(w, "Database error", http.StatusInternalServerError)
      return
    }
    defer rows.Close()

    var jobs []Job
    for rows.Next() {
      var job Job
      err := rows.Scan(&job.ID, &job.JobName, &job.CompanyName, &job.CreatedAt)
      if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
      }
      jobs = append(jobs, job)
    }

    if len(jobs) == 0 {
      jobs = []Job{}
    }
    json.NewEncoder(w).Encode(jobs)
  }
}

func isLidarUploaded(db *sql.DB, jobId string) (bool, error) {
  var lidarUploaded bool
  query := `SELECT lidar_uploaded FROM jobs WHERE id = $1`
  err := db.QueryRow(query, jobId).Scan(&lidarUploaded)
  if err != nil {
    return false, err
  }
  return lidarUploaded, nil
}

func HandleGetMapWithJob(templates *template.Template, db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    jobId := vars["id"]

    lidarUploaded, err := isLidarUploaded(db, jobId)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    fmt.Println(lidarUploaded)
    data := struct {
      ShowTools bool
    }{
      ShowTools: true,
    }
    err = templates.ExecuteTemplate(w, "index.html", data)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
}

func HandleGetMap(templates *template.Template) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    data := struct {
      ShowTools bool
    }{
      ShowTools: false,
    }
    err := templates.ExecuteTemplate(w, "index.html", data)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
}

func HandleLogout(store *sessions.CookieStore) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "SESSION_KEY")
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    // Revoke users authentication
    session.Values["authenticated"] = false
    session.Save(r, w)

    http.Redirect(w, r, "/login", http.StatusSeeOther)
  }
}


func HandleGetContactPage(templates *template.Template) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    err := templates.ExecuteTemplate(w, "contact.html", nil)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
}

func HandleCreateJob(store *sessions.CookieStore, db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "SESSION_KEY")
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    userId, ok := session.Values["user_id"].(string)
    if !ok {
        http.Error(w, "User ID not found in context", http.StatusUnauthorized)
        return
    }

    isAdmin, ok := session.Values["admin"].(bool)
    if !ok {
        http.Error(w, "Admin status not found in context", http.StatusUnauthorized)
        return
    }

    jobName := r.FormValue("job-name")
    companyName := r.FormValue("company-name")

    var adminId string

    if isAdmin {
      adminId = userId
    } else {
      err := db.QueryRow(`SELECT admin_id FROM users WHERE id = $1`, userId).Scan(&adminId)
      if err != nil {
        http.Error(w, "Error fetching admin id from database", http.StatusInternalServerError)
        return
      }
    }

    // this is false since no lidar has been uploaded on new job
    lidarUploaded := false
    id, err := insertJobIntoDb(jobName, companyName, adminId, lidarUploaded, db)
    if err != nil {
    	http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    
    http.Redirect(w, r, fmt.Sprintf("/map/%s", id), http.StatusSeeOther)
  }
}

func insertJobIntoDb(jobName string, companyName string, adminId string, lidarUploaded bool, db *sql.DB) (string, error) {
  query := `INSERT INTO jobs (id, job_name, company_name, admin_id, lidar_uploaded, created_at)
            VALUES (gen_random_uuid(), $1, $2, $3, $4, CURRENT_TIMESTAMP)
            RETURNING id
           `
  var id string
  err := db.QueryRow(query, jobName, companyName, adminId, lidarUploaded).Scan(&id)
  if err != nil {
    log.Printf("Error inserting job into database: %v", err)
    return "", err
  }

  return id, nil
}


// UPLOAD LAS FILES //

func getJobIdFromMux(r *http.Request) (string, error) {
  vars := mux.Vars(r)
  jobId, exists := vars["id"]
  if !exists {
    return "", fmt.Errorf("Error getting job id from api endpoint")
  }
  return jobId, nil
}

func getAdminIdFromJobsTable(store *sessions.CookieStore, db *sql.DB, jobId string) (string, error) {
  var adminId string
  query := `SELECT admin_id FROM jobs WHERE id = $1`
  err := db.QueryRow(query, jobId).Scan(&adminId)
  if err != nil {
    return "", fmt.Errorf("Error getting admin id from jobs table")
  }
  return adminId, nil
}

func getFilePath(w http.ResponseWriter, adminId, jobId string) (string, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    log.Fatalf("Error getting user's home directory: %v", err)
  }

  baseDir := filepath.Join(homeDir, "pivot/uploads")
  folderPath := filepath.Join(baseDir, adminId)
  if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
    return "", fmt.Errorf("Error creating file path on server") 
  }

  filePath := filepath.Join(folderPath, fmt.Sprintf("%s.las", jobId))
  return filePath, nil
} 

func checkIfFileExists(filePath string) error {
  _, err := os.Stat(filePath)
  if err == nil {
    return fmt.Errorf("File already exists. Do you want to replace it?")
  } else if !os.IsNotExist(err) {
    return fmt.Errorf("Error checking file existence: %v", err)
  }
  return nil
}
// TODO refactor handleupload and handlereplace
func HandleUploadLas(store *sessions.CookieStore, db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(40 << 20)

    file, _, err := r.FormFile("file")
    if err != nil {
      fmt.Println("error recieving file from form")
      http.Error(w, "Error recieving file from form", http.StatusInternalServerError)
      return
    }

    defer file.Close()

    vars := mux.Vars(r)
    jobId := vars["id"]

    var adminId string
    query := `SELECT admin_id FROM jobs WHERE id = $1`
    err = db.QueryRow(query, jobId).Scan(&adminId)
    if err != nil {
      fmt.Println("failed to get jobs")
      http.Error(w, "Error getting admin id from jobs", http.StatusInternalServerError)
      return
    }

    homeDir, err := os.UserHomeDir()
    if err != nil {
      log.Fatalf("Error getting user's home directory: %v", err)
    }

    baseDir := filepath.Join(homeDir, "pivot/uploads")
    folderPath := filepath.Join(baseDir, adminId)
    if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
      http.Error(w, "Error creating uploads folder", http.StatusInternalServerError)
      return
    }

    filePath := filepath.Join(folderPath, fmt.Sprintf("%s.las", jobId))

    _, err = os.Stat(filePath)
    if err == nil {
      http.Error(w, "File already exists. Do you want to replace it?", http.StatusConflict)
      return
    } else if !os.IsNotExist(err) {
      http.Error(w, "Error checking file existence", http.StatusInternalServerError)
      return
    }

    outFile, err := os.Create(filePath)
    if err != nil {
      http.Error(w, "Unable to create file on server", http.StatusInternalServerError)
      return
    }
    defer outFile.Close()

    if _, err := io.Copy(outFile, file); err != nil {
      http.Error(w, "Error copying file contents to file", http.StatusInternalServerError)
      return
    }

    //err = convertToOctree(folderPath, filePath, jobId)
    //if err != nil {
    //  http.Error(w, err.Error(), http.StatusInternalServerError)
    //  return
    //}

    query = `UPDATE jobs SET lidar_uploaded = true WHERE id = $1`
    _, err = db.Exec(query, jobId)
    if err != nil {
      http.Error(w, "error updating job lidar_uploaded status: %v", http.StatusInternalServerError)
      return 
    }


    response := map[string]interface{}{
      "message": "File uploaded successfully",
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
  }
}

func convertToOctree(folderPath, lasFilePath, jobId string) error {
  // Invoke PotreeConverter process
  potreeConverterPath := "/home/ja/pivot/static/PotreeConverter"

  outputDir := filepath.Join(folderPath, jobId)
  cmd := exec.Command(potreeConverterPath, lasFilePath, "-o", outputDir)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  err := cmd.Start()
  if err != nil {
    return fmt.Errorf("Error starting PotreeConverter process")
  }


	// Channel to listen for interrupt signal
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Handle interrupt signal to terminate the process gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
		// Terminate the process
		err := cmd.Process.Signal(os.Interrupt)
		if err != nil {
			log.Printf("Error sending interrupt signal to PotreeConverter process: %v\n", err)
		}
		log.Println("PotreeConverter process terminated by interrupt signal.")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("Error waiting for PotreeConverter process to finish: %v", err)
		}
		log.Println("PotreeConverter process completed successfully.")
	}

	return nil
  //err = cmd.Wait()
  //if err != nil {
  //  return fmt.Errorf("Error waiting for PotreeConverter process to finish")
  //}

  //return nil
}

func HandleReplaceLas(store *sessions.CookieStore, db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(40 << 20)

    file, _, err := r.FormFile("file")
    if err != nil {
      http.Error(w, "Error receiving file from form", http.StatusInternalServerError)
      return
    }
    defer file.Close()

    vars := mux.Vars(r)
    jobId := vars["id"]

    var adminId string
    query := `SELECT admin_id FROM jobs WHERE id = $1`
    err = db.QueryRow(query, jobId).Scan(&adminId)
    if err != nil {
      http.Error(w, "Error getting admin id from jobs", http.StatusInternalServerError)
      return
    }

    homeDir, err := os.UserHomeDir()
    if err != nil {
      log.Fatalf("Error getting user's home directory: %v", err)
    }

    baseDir := filepath.Join(homeDir, "pivot/uploads")
    folderPath := filepath.Join(baseDir, adminId)
    if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
      http.Error(w, "Error creating uploads folder", http.StatusInternalServerError)
      return
    }

    filePath := filepath.Join(folderPath, fmt.Sprintf("%s.las", jobId))
    fmt.Println(filePath)
    if _, err := os.Stat(filePath); err == nil {
      if err := os.Remove(filePath); err != nil {
        http.Error(w, "Error deleting existing file", http.StatusInternalServerError)
        return
      }
    }

    outFile, err := os.Create(filePath)
    if err != nil {
      fmt.Println("error copying file contents to file")
      http.Error(w, "Unable to create file on server", http.StatusInternalServerError)
      return
    }
    defer outFile.Close()

    if _, err := io.Copy(outFile, file); err != nil {
      fmt.Println("error copying file contents to file")
      http.Error(w, "Error copying file contents to file", http.StatusInternalServerError)
      return
    }

    //err = convertToOctree(folderPath, filePath, jobId)
    //if err != nil {
    //  http.Error(w, err.Error(), http.StatusInternalServerError)
    //  return
    //}

    response := map[string]interface{}{
      "message": "File uploaded successfully",
    }

    w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
  }
}

func HandleGetPoleLocations(store *sessions.CookieStore, db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    jobId := vars["id"]

    var lidarIsUploaded bool
    err := db.QueryRow(`SELECT lidar_uploaded FROM jobs WHERE id = $1`, jobId).Scan(&lidarIsUploaded)
    if err != nil {
      http.Error(w, "Error fetching lidar uploaded status", http.StatusInternalServerError)
      return
    }

    if lidarIsUploaded {
      var adminId string
      query := `SELECT admin_id FROM jobs WHERE id = $1`
      err = db.QueryRow(query, jobId).Scan(&adminId)
      if err != nil {
        fmt.Println("failed to get jobs")
        http.Error(w, "Error getting admin id from jobs", http.StatusInternalServerError)
        return
      }

      homeDir, err := os.UserHomeDir()
      if err != nil {
        log.Fatalf("Error getting user's home directory: %v", err)
      }

      filePath := filepath.Join(homeDir, "pivot/uploads", adminId, fmt.Sprintf("%s.las", jobId))
      fmt.Println(filePath)

      fileContent, err := ioutil.ReadFile(filePath)
      if err != nil {
        http.Error(w, "Error reading LAS file", http.StatusInternalServerError)
        return
      }

      pythonScriptPath := "/home/ja/pivot/static/scripts/get-pole-locations.py"

      cmd := exec.Command("python3", pythonScriptPath)

      cmd.Stdin = bytes.NewReader(fileContent)

      var out bytes.Buffer
      cmd.Stdout = &out

      err = cmd.Run()
      if err != nil {
          http.Error(w, "Error executing Python script: "+err.Error(), http.StatusInternalServerError)
          return
      }

      poleData := []byte(out.String())
      err = savePolesToDatabase(db, jobId, poleData)
      if err != nil {
        http.Error(w, "Error saving pole json file to database: "+err.Error(), http.StatusInternalServerError)
        return
      }

      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(http.StatusOK)
      _, err = w.Write(poleData)

      if err != nil {
          http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
          return
      }
    } else {
        http.Error(w, "Las file has not been uploaded. Upload las file to get pole locations", http.StatusForbidden)
        return
    }
  }
}

func savePolesToDatabase(db *sql.DB, jobId string, poleData []byte) error {
  query := `UPDATE jobs SET poles = $1 WHERE id = $2`

	_, err := db.Exec(query, poleData, jobId)
	if err != nil {
		return err
	}

  fmt.Println("no error saving to database")
	return nil
}

func HandleSavePathsOfPoleLine(store *sessions.CookieStore, db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    jobId := vars["id"]

		midspanData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading JSON data: "+err.Error(), http.StatusBadRequest)
			return
		}

    err = saveMidspansToDatabase(db, jobId, midspanData)
    if err != nil {
      http.Error(w, "Error saving midspans json file to database: "+err.Error(), http.StatusInternalServerError)
      return
    }

		response := map[string]interface{}{
			"message": "JSON data received and saved successfully",
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
  }
}

func saveMidspansToDatabase(db *sql.DB, jobId string, midspanData []byte) error {
  query := `UPDATE jobs SET midspans = $1 WHERE id = $2`

	_, err := db.Exec(query, midspanData, jobId)
	if err != nil {
    return fmt.Errorf("error updating markers for job %s: %v", jobId, err)
	}

  fmt.Println("no error saving to database")
	return nil
}


// PRE-CONDITIONS 
// lidar data, midspans collected from user, pole locations
//func HandleGetVegetationEncroachments(store *sessions.CookieStore, db *sql.DB) http.HandlerFunc {
//  return func(w http.ResponseWriter, r *http.Request) {
//    vars := mux.Vars(r)
//    jobId := vars["id"]
//
//    // FIRST CHECK: need midspans from user
//    var midspans sql.NullString
//    query = `SELECT midspans FROM jobs WHERE id = $1`
//    err := db.QueryRow(query, jobId).Scan(&midspans)
//    if err != nil {
//      http.Error(w, "Error fetching vegetation check from database", http.StatusInternalServerError)
//      return
//    }
//
//    if !midspans.Valid {
//      http.Error(w, "Midspans must first be collected before getting vegetation encroachments", http.StatusBadRequest)
//      return
//    }
//
//    midspansJSON, err := json.Marshal(midspans.String)
//    if err != nil {
//      http.Error(w, "Error encoding JSON (midspans) response: "+err.Error(), http.StatusInternalServerError)
//      return
//    }
//
//    var vegetation sql.NullString
//    query = `SELECT vegetation FROM jobs WHERE id = $1`
//    err := db.QueryRow(query, jobId).Scan(&vegetation)
//    if err != nil {
//      http.Error(w, "Error fetching vegetation check from database", http.StatusInternalServerError)
//      return
//    }
//
//    if vegetation.Valid {
//      jsonResponse, err := json.Marshal(vegetation.String)
//      if err != nil {
//        http.Error(w, "Error encoding JSON (vegetation) response: "+err.Error(), http.StatusInternalServerError)
//        return
//      }
//
//      w.Header().Set("Content-Type", "application/json")
//      w.WriteHeader(http.StatusOK)
//      w.Write(jsonResponse)
//
//    } else {
//      var adminId string
//      query := `SELECT admin_id FROM jobs WHERE id = $1`
//      err := db.QueryRow(query, jobId).Scan(&adminId)
//
//      if err != nil {
//        fmt.Println("failed to get jobs")
//        http.Error(w, "Error getting admin id from jobs", http.StatusInternalServerError)
//        return
//      }
//
//      homeDir, err := os.UserHomeDir()
//      if err != nil {
//        log.Fatalf("Error getting user's home directory: %v", err)
//      }
//
//      filePath := filepath.Join(homeDir, "pivot/uploads", adminId, fmt.Sprintf("%s.las", jobId))
//
//      fileContent, err := ioutil.ReadFile(filePath)
//      if err != nil {
//        http.Error(w, "Error reading LAS file", http.StatusInternalServerError)
//        return
//      }
//
//      pythonScriptPath := "/home/ja/pivot/static/scripts/get-vegetation-encroachments.py"
//
//      cmd := exec.Command("python3", pythonScriptPath, string(midspansJSON))
//
//      cmd.Stdin = bytes.NewReader(fileContent)
//
//      var out bytes.Buffer
//      cmd.Stdout = &out
//
//      err = cmd.Run()
//      if err != nil {
//          http.Error(w, "Error executing Python script: "+err.Error(), http.StatusInternalServerError)
//          return
//      }
//
//      vegetationData := []byte(out.String())
//      err = saveVegetationToDatabase(db, jobId, vegetationData)
//      if err != nil {
//        http.Error(w, "Error saving pole json file to database: "+err.Error(), http.StatusInternalServerError)
//        return
//      }
//
//      w.Header().Set("Content-Type", "application/json")
//      w.WriteHeader(http.StatusOK)
//      _, err = w.Write(vegetationData)
//
//      if err != nil {
//          http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
//          return
//      }
//    }
//  }
//}
//
//func saveVegetationToDatabase(db *sql.DB, jobId string, vegetationData []byte) error {
//  query := `UPDATE jobs SET vegetation = $1 WHERE id = $2`
//	_, err := db.Exec(query, vegetationData, jobId)
//	if err != nil {
//		return err
//	}
//	return nil
//}
