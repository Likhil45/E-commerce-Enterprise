// // filepath: c:\GOLang\E-Commerce Platform\cassandra\client.go
package cassandra

// import (
// 	"context"
// 	"e-commerce/logger"
// 	"e-commerce/protobuf/protobuf"
// 	"log"
// 	"time"

// 	"github.com/gocql/gocql"
// )

// var Session *gocql.Session

// func InitCassandra() {
// 	cluster := gocql.NewCluster("cassandra:9042")
// 	cluster.Keyspace = "ecommerce"
// 	cluster.Consistency = gocql.Quorum

// 	session, err := cluster.CreateSession()
// 	if err != nil {
// 		log.Fatalf("Failed to connect to Cassandra: %v", err)
// 	}

// 	log.Println("Connected to Cassandra successfully")
// 	Session = session
// }

// type OrderTrackingService struct {
// 	protobuf.UnimplementedOrderTrackingServiceServer
// }

// // SearchOrders searches for orders based on filters
// func (s *OrderTrackingService) SearchOrders(ctx context.Context, req *protobuf.SearchOrdersRequest) (*protobuf.SearchOrdersResponse, error) {
// 	logger.Logger.Infof("SearchOrders called with filters: %+v", req)

// 	return &protobuf.SearchOrdersResponse{Orders: orders}, nil
// }

// // GetOrderTracking provides real-time tracking updates for an order
// func (s *OrderTrackingService) GetOrderTracking(ctx context.Context, req *protobuf.OrderTrackingRequest) (*protobuf.OrderTrackingResponse, error) {
// 	logger.Logger.Infof("GetOrderTracking called for OrderID=%s", req.OrderId)

// 	// Simulate fetching tracking data
// 	trackingResponse := &protobuf.OrderTrackingResponse{
// 		OrderId:           req.OrderId,
// 		Status:            "In Transit",
// 		EstimatedDelivery: time.Now().Add(48 * time.Hour).Format(time.RFC3339),
// 	}

// 	return trackingResponse, nil
// }
