package handlers

import (
 "database/sql"
 "encoding/json"
 "net/http"
 "strconv"
 "strings"
 "time"

 "github.com/google/uuid"
)

type UserHandler struct {
 DB *sql.DB
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 path := strings.TrimPrefix(r.URL.Path, "/")

 // Маршруты для ролей (/role, /role/{id})[span_1](start_span)[span_1](end_span)
 if path == "role" {
  switch r.Method {
  case http.MethodPost:
   h.CreateRole(w, r)
  case http.MethodGet:
   h.GetRoles(w, r)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 if strings.HasPrefix(path, "role/") {
  id := strings.TrimPrefix(path, "role/")
  switch r.Method {
  case http.MethodPut:
   h.UpdateRole(w, r, id)
  case http.MethodDelete:
   h.DeleteRole(w, r, id)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 // Маршруты для пользователей (/user, /user/{id}, /user/role/{id})[span_2](start_span)[span_2](end_span)
 if path == "user" {
  switch r.Method {
  case http.MethodPost:
   h.CreateUser(w, r)
  case http.MethodGet:
   h.GetUsers(w, r)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 if strings.HasPrefix(path, "user/role/") {
  roleID := strings.TrimPrefix(path, "user/role/")
  if r.Method == http.MethodGet {
   h.GetUsersByRole(w, r, roleID)
   return
  }
  http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  return
 }

 if strings.HasPrefix(path, "user/") {
  id := strings.TrimPrefix(path, "user/")
  switch r.Method {
  case http.MethodPut:
   h.UpdateUser(w, r, id)
  case http.MethodDelete:
   h.DeleteUser(w, r, id)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 http.NotFound(w, r)
}

// === РОЛИ (Role) ===

// POST /role - создать роль[span_3](start_span)[span_3](end_span)
func (h *UserHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
 var body struct {
  Name string `json:"name"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 id := uuid.New().String()
 _, err := h.DB.Exec("INSERT INTO roles (id, name) VALUES (?, ?)", id, body.Name)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusCreated)
 json.NewEncoder(w).Encode(map[string]string{"id": id, "name": body.Name})
}

// GET /role - получить список ролей[span_4](start_span)[span_4](end_span)
func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
 rows, err := h.DB.Query("SELECT id, name FROM roles")
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 defer rows.Close()

 var roles []map[string]string
 for rows.Next() {
  var id, name string
  if err := rows.Scan(&id, &name); err != nil {
   continue
  }
  roles = append(roles, map[string]string{"id": id, "name": name})
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(roles)
}

// PUT /role/{id} - обновить роль по id[span_5](start_span)[span_5](end_span)
func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request, id string) {
 var body struct {
  Name string `json:"name"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 _, err := h.DB.Exec("UPDATE roles SET name = ? WHERE id = ?", body.Name, id)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// DELETE /role/{id} - удалить роль[span_6](start_span)[span_6](end_span)
func (h *UserHandler) DeleteRole(w http.ResponseWriter, r *http.Request, id string) {
 _, err := h.DB.Exec("DELETE FROM roles WHERE id = ?", id)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// === ПОЛЬЗОВАТЕЛИ (User) ===

// POST /user - создать пользователя[span_7](start_span)[span_7](end_span)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
 var body struct {
  Name   string `json:"name"`
  Email  string `json:"email"`
  Phone  string `json:"phone"`
  RoleID string `json:"role_id"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 id := uuid.New().String()
 regDate := time.Now().Format("2006-01-02 15:04:05")

 _, err := h.DB.Exec(
  "INSERT INTO users (id, name, email, phone, role_id, registration_date) VALUES (?, ?, ?, ?, ?, ?)",
  id, body.Name, body.Email, body.Phone, body.RoleID, regDate,
 )
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusCreated)
 json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// GET /user - получить всех пользователей (с поддержкой пагинации)[span_8](start_span)[span_8](end_span)
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
 limit := 10
 if l := r.URL.Query().Get("limit"); l != "" {
  if parsed, err := strconv.Atoi(l); err == nil {
   limit = parsed
  }
 }
 offset := 0
 if o := r.URL.Query().Get("offset"); o != "" {
  if parsed, err := strconv.Atoi(o); err == nil {
   offset = parsed
  }
 }

rows, err := h.DB.Query(
	`SELECT u.id, u.name, u.email, u.phone, r.name as role_name, u.registration_date 
	FROM users u 
	LEFT JOIN roles r ON u.role_id = r.id 
	LIMIT ? OFFSET ?`, // <- Запрос обернут в `, запятая стоит снаружи
	limit, 
	offset,
)






 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 defer rows.Close()

 var users []map[string]interface{}
 for rows.Next() {
  var id, name, email, phone, roleName, regDate string
  if err := rows.Scan(&id, &name, &email, &phone, &roleName, &regDate); err != nil {
   continue
  }
  users = append(users, map[string]interface{}{
   "id":                 id,
   "name":               name,
   "email":              email,
   "phone":              phone,
   "role":               roleName,
   "registration_date":  regDate,
  })
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(users)
}

// GET /user/role/{id} - получить всех пользователей роли по id роли[span_9](start_span)[span_9](end_span)
func (h *UserHandler) GetUsersByRole(w http.ResponseWriter, r *http.Request, roleID string) {
 rows, err := h.DB.Query(
 		`SELECT id, name, email, phone, registration_date 
 		FROM users 
 		WHERE role_id = ?`, // <- Запрос обернут в `, запятая снаружи
 		roleID,
 	)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 defer rows.Close()

 var users []map[string]interface{}
 for rows.Next() {
  var id, name, email, phone, regDate string
  if err := rows.Scan(&id, &name, &email, &phone, &regDate); err != nil {
   continue
  }
  users = append(users, map[string]interface{}{
   "id":                id,
   "name":              name,
   "email":             email,
   "phone":             phone,
   "registration_date": regDate,
  })
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(users)
}

// PUT /user/{id} - обновить пользователя[span_10](start_span)[span_10](end_span)
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request, id string) {
 var body struct {
  Name   string `json:"name"`
  Email  string `json:"email"`
  Phone  string `json:"phone"`
  RoleID string `json:"role_id"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 _, err := h.DB.Exec(
  "UPDATE users SET name = ?, email = ?, phone = ?, role_id = ? WHERE id = ?",
  body.Name, body.Email, body.Phone, body.RoleID, id,
 )
if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
// DELETE /user/{id} - удалить пользователя[span_11](start_span)[span_11](end_span)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, id string) {
	_, err := h.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}