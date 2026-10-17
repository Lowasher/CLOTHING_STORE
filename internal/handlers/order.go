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

type OrderHandler struct {
 DB *sql.DB
}

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 path := strings.TrimPrefix(r.URL.Path, "/")
 
 if path == "order" {
  switch r.Method {
  case http.MethodPost:
   h.CreateOrder(w, r)
  case http.MethodGet:
   h.GetOrders(w, r)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 if strings.HasPrefix(path, "order/") {
  id := strings.TrimPrefix(path, "order/")
  switch r.Method {
  case http.MethodGet:
   h.GetOrder(w, r, id)
  case http.MethodPut:
   h.UpdateOrder(w, r, id)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 if path == "order_item" {
  switch r.Method {
  case http.MethodPost:
   h.CreateOrderItem(w, r)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 if strings.HasPrefix(path, "order_item/") {
  orderID := strings.TrimPrefix(path, "order_item/")
  switch r.Method {
  case http.MethodGet:
   h.GetOrderItems(w, r, orderID)
  default:
   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
  return
 }

 http.NotFound(w, r)
}

// POST /order - добавить заказ[span_1](start_span)[span_1](end_span)
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
 var body struct {
  UserID string `json:"user_id"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 id := uuid.New().String()
 dateTime := time.Now().Format(time.RFC3339)
 status := "Принят" // Статус заказа из ТЗ по умолчанию

 _, err := h.DB.Exec(
  "INSERT INTO orders (id, user_id, datetime, status) VALUES (?, ?, ?, ?)",
  id, body.UserID, dateTime, status,
 )
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusCreated)
 json.NewEncoder(w).Encode(map[string]string{"id": id, "status": status})
}

// GET /order - получить список заказов (с пагинацией и фильтрацией по статусу и интервалу времени)[span_2](start_span)[span_2](end_span)
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
    query := "SELECT id, user_id, datetime, status FROM orders WHERE 1=1"
 var args []interface{}

 if status := r.URL.Query().Get("status"); status != "" {
  query += " AND status = ?"
  args = append(args, status)
 }

 if start := r.URL.Query().Get("start_date"); start != "" {
  query += " AND datetime >= ?"
  args = append(args, start)
 }

 if end := r.URL.Query().Get("end_date"); end != "" {
  query += " AND datetime <= ?"
  args = append(args, end)
 }

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

 var orders []map[string]interface{}
 for rows.Next() {
  var id, userID, dateTime, status string
  if err := rows.Scan(&id, &userID, &dateTime, &status); err != nil {
   continue
  }
  orders = append(orders, map[string]interface{}{
   "id":       id,
   "user_id":  userID,
   "datetime": dateTime,
   "status":   status,
  })
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(orders)
}

// GET /order/{id} - получить заказ[span_3](start_span)[span_3](end_span)
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request, id string) {
 var userID, dateTime, status string
 err := h.DB.QueryRow("SELECT user_id, datetime, status FROM orders WHERE id = ?", id).
  Scan(&userID, &dateTime, &status)
 if err == sql.ErrNoRows {
  http.Error(w, "Order not found", http.StatusNotFound)
  return
 } else if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(map[string]interface{}{
  "id":       id,
  "user_id":  userID,
  "datetime": dateTime,
  "status":   status,
 })
}

// PUT /order/{id} - обновить заказ (в т.ч. изменить статус)[span_4](start_span)[span_4](end_span)[span_5](start_span)[span_5](end_span)
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request, id string) {
 var body struct {
  Status string `json:"status"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 _, err := h.DB.Exec("UPDATE orders SET status = ? WHERE id = ?", body.Status, id)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// POST /order_item - добавить элемент заказа[span_6](start_span)[span_6](end_span)
func (h *OrderHandler) CreateOrderItem(w http.ResponseWriter, r *http.Request) {
 var body struct {
  OrderID   string  `json:"order_id"`
  ProductID string  `json:"product_id"`
  Price     float64 `json:"price"`
 }
 if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }

 id := uuid.New().String()
 _, err := h.DB.Exec(
  "INSERT INTO order_items (id, order_id, product_id, price) VALUES (?, ?, ?, ?)",
  id, body.OrderID, body.ProductID, body.Price,
 )
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.WriteHeader(http.StatusCreated)
 json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// GET /order_item/{order_id} - получить элементы заказа[span_7](start_span)[span_7](end_span)
func (h *OrderHandler) GetOrderItems(w http.ResponseWriter, r *http.Request, orderID string) {
    rows, err := h.DB.Query(`
		SELECT oi.id, oi.product_id, oi.price, p.name 
		FROM order_items oi 
		JOIN products p ON oi.product_id = p.id 
		WHERE oi.order_id = ?`, 
		orderID,
	)

 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 defer rows.Close()

 var items []map[string]interface{}
 for rows.Next() {
  var id, productID, name string
  var price float64
  if err := rows.Scan(&id, &productID, &price, &name); err != nil {
   continue
  }
  items = append(items, map[string]interface{}{
   "id":         id,
   "product_id": productID,
   "name":       name,
   "price":      price,
  })
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(items)
}