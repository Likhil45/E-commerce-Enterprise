package conshand

import (
	"context"
	"e-commerce/consumer-service/consmetrics"
	"e-commerce/logger"
	"e-commerce/protobuf/protobuf"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

func CallPaymentService(paymentReq *protobuf.PaymentRequest) (*protobuf.PaymentResponse, error) {
	// Connect to gRPC Payment Service
	logger.Logger.Infof("Connecting to Payment Service for OrderID=%d", paymentReq.OrderId)
	conn, err := grpc.Dial("payment-service:50080", grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Payment Service: %v", err)
		consmetrics.GrpcCallsTotal.WithLabelValues("PaymentService", "ProcessPayment", "failed").Inc()
		return nil, err
	}
	defer conn.Close()

	// Creating gRPC client
	client := protobuf.NewPaymentServiceClient(conn)

	// Calling ProcessPayment RPC
	logger.Logger.Infof("Calling ProcessPayment for OrderID=%d", paymentReq.OrderId)
	response, err := client.ProcessPayment(context.Background(), paymentReq)
	if err != nil {
		logger.Logger.Errorf("Error calling ProcessPayment: %v", err)
		consmetrics.GrpcCallsTotal.WithLabelValues("PaymentService", "ProcessPayment", "failed").Inc()
		return nil, err
	}

	// Connect to Redis Service
	logger.Logger.Infof("Connecting to Redis Service for UserID=%s", paymentReq.UserId)
	connR, err := grpc.Dial("redis-service:50010", grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Redis Service: %v", err)
		consmetrics.GrpcCallsTotal.WithLabelValues("RedisService", "SetData", "failed").Inc()
		return nil, err
	}
	defer connR.Close()
	clientR := protobuf.NewRedisServiceClient(connR)

	// Set data in Redis based on payment status
	var setReq *protobuf.SetRequest
	if response.Status == "SUCCESS" {
		setReq = &protobuf.SetRequest{Key: paymentReq.UserId, Value: "Your Order was created Successfully!!!"}
	} else {
		setReq = &protobuf.SetRequest{Key: paymentReq.UserId, Value: "Payment Failed!!, Update card details"}
	}
	logger.Logger.Infof("Setting data in Redis for UserID=%s, Value=%s", setReq.Key, setReq.Value)
	resp, err := clientR.SetData(context.Background(), setReq)
	if err != nil {
		logger.Logger.Errorf("Failed to set data in Redis: %v", err)
	}
	consmetrics.GrpcCallsTotal.WithLabelValues("RedisService", "SetData", "success").Inc()

	logger.Logger.Infof("ProcessPayment Response: %+v", response)
	logger.Logger.Infof("Redis SetData Response: %+v", resp)

	return response, nil
}

// CallNotificationService sends notifications via Redis and Notification Service
func CallNotificationService(not *protobuf.NotificationRequest) (*protobuf.NotificationResponse, error) {
	logger.Logger.Infof("Calling Notification Service for UserID=%s", not.UserId)

	// Connect to Redis Service
	connR, err := grpc.Dial("redis-service:50010", grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Redis Service: %v", err)
		consmetrics.GrpcCallsTotal.WithLabelValues("RedisService", "SetData", "failed").Inc()
		return nil, err
	}
	defer connR.Close()
	clientR := protobuf.NewRedisServiceClient(connR)

	// Set data in Redis
	logger.Logger.Infof("Setting data in Redis for UserID=%s, Value=%s", not.UserId, not.Message)
	check, err := clientR.SetData(context.Background(), &protobuf.SetRequest{Key: not.UserId, Value: not.Message})
	if err != nil {
		logger.Logger.Errorf("Failed to set data in Redis: %v", err)
		return nil, err
	}
	logger.Logger.Infof("SetData response from Redis: %+v", check)

	// Connect to Notification Service
	conn, err := grpc.Dial("notification-service:50020", grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Notification Service: %v", err)
		consmetrics.GrpcCallsTotal.WithLabelValues("NotificationService", "SendNotification", "failed").Inc()
		return nil, err
	}
	defer conn.Close()
	client := protobuf.NewNotificationServiceClient(conn)

	// Send notification
	logger.Logger.Infof("Sending notification to UserID=%s", not.UserId)
	response, err := client.SendNotification(context.Background(), not)
	if err != nil {
		logger.Logger.Errorf("Failed to send notification: %v", err)
		return nil, err
	}
	logger.Logger.Infof("Notification sent successfully: %+v", response)

	return response, nil
}

func CallInventoryService(req *protobuf.StockUpdateRequest) (*protobuf.StockResponse, error) {
	logger.Logger.Infof("Calling Inventory Service for ProductID=%d, Quantity=%d", req.ProductId, req.Quantity)

	// Connect to Inventory Service
	conn, err := grpc.Dial("inventory-service:50051", grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Inventory Service: %v", err)
		return nil, err
	}
	defer conn.Close()
	client := protobuf.NewInventoryServiceClient(conn)

	// Update stock
	resp, err := client.UpdateStock(context.Background(), req)
	if err != nil {
		logger.Logger.Errorf("Failed to update stock in Inventory Service: %v", err)
		return nil, err
	}
	logger.Logger.Infof("Stock updated successfully in Inventory Service: %+v", resp)

	return resp, nil
}
