package models

// User Service
type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	UserID   int32  `json:"userid"`
}

type AuthenticateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GetUserRequest struct {
	UserID int32 `json:"user_id"`
}

type UserResponse struct {
	UserID   int32  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// Product Service
type ProductRequest struct {
	ProductID   int32   `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
}

type ProductIDRequest struct {
	ProductID int32 `json:"product_id"`
}

type ProductResponse struct {
	ProductID   int32   `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
}

type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
}

type Empty struct{}

// Inventory Service
type StockRequest struct {
	ProductID int32 `json:"product_id"`
}

type StockUpdateRequest struct {
	ProductID int32 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type StockResponse struct {
	ProductID int32  `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Status    string `json:"status"`
}

// Order Service
type OrderRequest struct {
	OrderID     int32       `json:"order_id"`
	UserID      int32       `json:"user_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float32     `json:"total_amount"`
}

type OrderIDRequest struct {
	OrderID int32 `json:"order_id"`
}

type OrderResponse struct {
	OrderID     int32       `json:"order_id"`
	UserID      int32       `json:"user_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float32     `json:"total_amount"`
}

type OrderItem struct {
	ProductID int32   `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	Price     float32 `json:"price"`
}

type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
}

// Payment Service
type PaymentRequest struct {
	OrderID int32   `json:"order_id"`
	Amount  float32 `json:"amount"`
	Method  string  `json:"method"`
}

type PaymentResponse struct {
	PaymentID int32   `json:"payment_id"`
	OrderID   int32   `json:"order_id"`
	Amount    float32 `json:"amount"`
	Status    string  `json:"status"`
}

// Notification Service
type NotificationRequest struct {
	UserID  int32  `json:"user_id"`
	Message string `json:"message"`
}

type NotificationResponse struct {
	Status string `json:"status"`
}

// Audit & Logging Service
type LogRequest struct {
	Event       string `json:"event"`
	Description string `json:"description"`
}

type LogResponse struct {
	Status string `json:"status"`
}

type LogListResponse struct {
	Logs []LogRequest `json:"logs"`
}

// Order Tracking & Search Service
type OrderTrackingRequest struct {
	OrderID int32 `json:"order_id"`
}

type OrderTrackingResponse struct {
	OrderID int32        `json:"order_id"`
	Events  []OrderEvent `json:"events"`
}

type OrderEvent struct {
	Event string `json:"event"`
}

type OrderSearchRequest struct {
	UserID    int32  `json:"user_id"`
	ProductID int32  `json:"product_id"`
	Status    string `json:"status"`
}

type OrderSearchResponse struct {
	Orders []OrderTrackingResponse `json:"orders"`
}
