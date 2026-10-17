package handlers
import (
 "database/sql"
 "encoding/json"
 "net/http"
 "strconv"

 "github.com/google/uuid"
)

type ProductHandler struct {
 DB *sql.DB
}

func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 switch r.Method {
 case http.MethodGet:
  h.GetProducts(w, r)
 case http.MethodPost:
  h.CreateProduct(w, r)
 default:
  http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
 }
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
 query := "SELECT id, name, type, gender, size, color, description, brand, season, price FROM products WHERE 1=1"
 var args []interface{}

 // Фильтры
 if t := r.URL.Query().Get("type"); t != "" {
  query += " AND type = ?"
  args = append(args, t)
 }
 if g := r.URL.Query().Get("gender"); g != "" {
  query += " AND gender = ?"
  args = append(args, g)
 }
 if sz := r.URL.Query().Get("size"); sz != "" {
  query += " AND size = ?"
  args = append(args, sz)
 }
 if c := r.URL.Query().Get("color"); c != "" {
  query += " AND color = ?"
  args = append(args, c)
 }
 if b := r.URL.Query().Get("brand"); b != "" {
  query += " AND brand = ?"
  args = append(args, b)
 }
 if s := r.URL.Query().Get("season"); s != "" {
  query += " AND season = ?"
  args = append(args, s)
 }

 // Сортировка по цене
 if sort := r.URL.Query().Get("sort"); sort == "asc" {
  query += " ORDER BY price ASC"
 } else if sort == "desc" {
  query += " ORDER BY price DESC"
 }

 // Пагинация
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
 query += " LIMIT ? OFFSET ?"
 args = append(args, limit, offset)

 rows, err := h.DB.Query(query, args...)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 defer rows.Close()

 var products []map[string]interface{}
 for rows.Next() {
  var id, name, pType, gender, size, color, description, brand, season string
  var price float64
  if err := rows.Scan(&id, &name, &pType, &gender, &size, &color, &description, &brand, &season, &price); err != nil {
   continue
  }
  products = append(products, map[string]interface{}{
   "id":          id,
   "name":        name,
   "type":        pType,
   "gender":      gender,
   "size":        size,
   "color":       color,
   "description": description,
   "brand":       brand,
   "season":      season,
   "price":       price,
  })
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
 var p struct {
  Name        string  `json:"name"`
  Type        string  `json:"type"`
  Gender      string  `json:"gender"`
  Size        string  `json:"size"`
  Color       string  `json:"color"`
  Description string  `json:"description"`
  Brand       string  `json:"brand"`
  Season      string  `json:"season"`
  Price       float64 `json:"price"`
 }

 if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 id := uuid.New().String()
 _, err := h.DB.Exec(
  "INSERT INTO products (id, name, type, gender, size, color, description, brand, season, price) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
  id, p.Name, p.Type, p.Gender, p.Size, p.Color, p.Description, p.Brand, p.Season, p.Price,
 )
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusCreated)
 json.NewEncoder(w).Encode(map[string]string{"id": id})
}