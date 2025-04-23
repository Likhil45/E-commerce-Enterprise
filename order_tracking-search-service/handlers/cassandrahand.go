// // filepath: c:\GOLang\E-Commerce Platform\services\order_history.go
package cassandrahand

// import (
// 	"e-commerce/models"
// 	"e-commerce/order_tracking-search-service/cassandra"
// 	"log"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// func SaveOrderHistory(order models.Order) error {
// 	query := `INSERT INTO order_history (order_id, user_id, product_id, status, total_amount, created_at)
//               VALUES (?, ?, ?, ?, ?, ?)`

// 	for _, item := range order.OrderItems {
// 		err := cassandra.Session.Query(query,
// 			order.OrderID,
// 			order.UserID,
// 			item.ProductID,
// 			order.Status,
// 			order.TotalPrice,
// 			// order.CreatedAt,
// 		).Exec()
// 		if err != nil {
// 			log.Printf("Failed to save order history: %v", err)
// 			return err
// 		}
// 	}

// 	log.Println("Order history saved successfully")
// 	return nil
// }
// func SearchOrders(c *gin.Context) {
// 	userID := c.Query("user_id")
// 	productID := c.Query("product_id")
// 	status := c.Query("status")
// 	startDate := c.Query("start_date")
// 	endDate := c.Query("end_date")

// 	results, err := SearchOrderHistory(userID, productID, status, startDate, endDate)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search orders"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, results)
// }

// func SearchOrderHistory(userID, productID, status, startDate, endDate string) ([]map[string]interface{}, error) {
// 	query := `SELECT order_id, user_id, product_id, status, total_amount, created_at FROM order_history WHERE user_id = ?`
// 	params := []interface{}{userID}

// 	if startDate != "" && endDate != "" {
// 		query += " AND created_at >= ? AND created_at <= ?"
// 		params = append(params, startDate, endDate)
// 	}

// 	if productID != "" {
// 		query += " AND product_id = ?"
// 		params = append(params, productID)
// 	}

// 	if status != "" {
// 		query += " AND status = ?"
// 		params = append(params, status)
// 	}

// 	iter := cassandra.Session.Query(query, params...).Iter()
// 	var results []map[string]interface{}
// 	for {
// 		row := make(map[string]interface{})
// 		if !iter.MapScan(row) {
// 			break
// 		}
// 		results = append(results, row)
// 	}

// 	if err := iter.Close(); err != nil {
// 		return nil, err
// 	}

// 	return results, nil
// }

// func GetOrderStatus(c *gin.Context) {
//     orderID := c.Param("order_id")

//     query := `SELECT status FROM order_history WHERE order_id = ? LIMIT 1`
//     var status string
//     err := cassandra.Session.Query(query, orderID).Scan(&status)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order status"})
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{"order_id": orderID, "status": status})
// }
