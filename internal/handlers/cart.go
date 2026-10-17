package handlers


import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type CartHandler struct {
 DB *sql.DB
}

func (h *CartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 path := strings.TrimPrefix(r.URL.Path, "/cart")
 if path == "" || path == "/" {
  switch r.Method {
  case http.MethodPost:
   h.AddToCart(w, r)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 parts := strings.Split(strings.Trim(path, "/"), "/")
 if len(parts) == 1 {
  id := parts[0]
  switch r.Method {
  case http.MethodGet:
   // GET /cart/{user_id} - получение корзины пользователя[span_1](start_span)[span_1](end_span)
   h.GetCartByUser(w, r, id)
  case http.MethodPut:
   // PUT /cart/{id} - обновить продукт в корзине[span_2](start_span)[span_2](end_span)
   h.UpdateCartItem(w, r, id)
  case http.MethodDelete:
   // DELETE /cart/{id} - удалить продукт из корзины[span_3](start_span)[span_3](end_span)
   h.DeleteCartItem(w, r, id)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 http.NotFound(w, r)
}

// POST /cart - добавить продукт в корзину[span_4](start_span)[span_4](end_span)
func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
    var item struct {
        UserID    string `json:"user_id"`
        ProductID string `json:"product_id"`
        Count     int    `json:"count"`
    }

 if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 if item.Count <= 0 {
  item.Count = 1
 }

 id := uuid.New().String()
 _, err := h.DB.Exec(
  "INSERT INTO carts (id, user_id, product_id, count) VALUES (?, ?, ?, ?)",
  id, item.UserID, item.ProductID, item.Count,
 )
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusCreated)
 json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// GET /cart/{user_id} - получить корзину пользователя[span_5](start_span)[span_5](end_span)
func (h *CartHandler) GetCartByUser(w http.ResponseWriter, r *http.Request, userID string) {
    rows, err := h.DB.Query(
		`SELECT c.id, c.product_id, c.count, p.name, p.price 
		FROM carts c 
		JOIN products p ON c.product_id = p.id 
		WHERE c.user_id = ?`, // <- Апостроф перенесен сюда, запятая снаружи
		userID,
	)
 
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 defer rows.Close()

 var cartItems []map[string]interface{}
 for rows.Next() {
  var id, productID, name string
  var count int
  var price float64
  if err := rows.Scan(&id, &productID, &count, &name, &price); err != nil {
   continue
  }
  cartItems = append(cartItems, map[string]interface{}{
   "id":         id,
   "product_id": productID,
   "name":       name,
   "price":      price,
   "count":      count,
  })
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(cartItems)
}

// PUT /cart/{id} - обновить продукт в корзине[span_6](start_span)[span_6](end_span)
func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request, id string) {
 var body struct {
  Count int `json:"count"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 _, err := h.DB.Exec("UPDATE carts SET count = ? WHERE id = ?", body.Count, id)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// DELETE /cart/{id} - удалить продукт из корзины[span_7](start_span)[span_7](end_span)
func (h *CartHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request, id string) {
 _, err := h.DB.Exec("DELETE FROM carts WHERE id = ?", id)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}