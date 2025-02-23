package models

type OrderItem struct {
	OrderItemID int     `json:"order_item_id" gorm:"primaryKey"`
	OrderID     int     `json:"order_id" gorm:"not null;unique"`
	ProductID   int     `json:"product_id" gorm:"not null;unique"`
	Quantity    int     `json:"quantity" gorm:"not null"`
	Price       float64 `json:"price" gorm:"not null"`
	Order       Order   `json:"order" gorm:"foreignKey:OrderID"`
	Product     Product `json:"product" gorm:"foreignKey:ProductID"`
}
