package conshand

import (
	"context"
	"log"

	"e-commerce/protobuf/protobuf"

	"google.golang.org/grpc"
)

func CallPaymentService(paymentReq *protobuf.PaymentRequest) (*protobuf.PaymentResponse, error) {
	//  Connect to gRPC Payment Service
	conn, err := grpc.Dial("payment-service:50080", grpc.WithInsecure())
	if err != nil {
		log.Println("Failed to connect to Payment Service:", err)
		return nil, err
	}
	defer conn.Close()

	//  Creating gRPC client
	client := protobuf.NewPaymentServiceClient(conn)

	//  Calling ProcessPayment RPC
	response, err := client.ProcessPayment(context.Background(), paymentReq)
	if err != nil {
		log.Println("Error calling ProcessPayment:", err)
		return nil, err
	}

	connR, err := grpc.Dial("redis-service:50010", grpc.WithInsecure())
	if err != nil {
		log.Println("Failed to connect to Redis Service:", err)
		return nil, err
	}
	defer conn.Close()
	clientR := protobuf.NewRedisServiceClient(connR)
	// usrStr := strconv.Itoa(int(paymentReq.GetUserId()))
	var setReq *protobuf.SetRequest
	if response.Status == "SUCCESS" {
		setReq = &protobuf.SetRequest{Key: paymentReq.UserId, Value: "Your Order was created Successfully!!!"}
	} else {
		setReq = &protobuf.SetRequest{Key: paymentReq.UserId, Value: "Payment Failed!!, Update card details"}
	}
	resp, err := clientR.SetData(context.Background(), setReq)
	if err != nil {
		log.Println("Unable to set data to redis")
	}
	log.Println(resp)

	return response, nil
}

//Call Notification

func CallNotificationService(not *protobuf.NotificationRequest) (*protobuf.NotificationResponse, error) {
	connR, err := grpc.Dial("redis-service:50010", grpc.WithInsecure())
	if err != nil {
		log.Println("unable to dial to port 50010 - Redis service", err)
		return nil, err
	}
	defer connR.Close()
	clientR := protobuf.NewRedisServiceClient(connR)
	// usrId := strconv.Itoa(int(not.UserId))
	check, err1 := clientR.SetData(context.Background(), &protobuf.SetRequest{Key: not.UserId, Value: not.Message})
	if err1 != nil {
		log.Println("Unable to send data to notification service", err1)
		return nil, err1
	}
	log.Println("The response is ", check)

	conn, err := grpc.Dial("notification-service:50020", grpc.WithInsecure())
	if err != nil {
		log.Println("unable to dial to port 50020 - Notification service", err)
		return nil, err
	}
	defer conn.Close()
	client := protobuf.NewNotificationServiceClient(conn)

	response, err1 := client.SendNotification(context.Background(), not)
	if err1 != nil {
		log.Println("Unable to send data to notification service", err1)
		return nil, err1
	}
	return response, nil

}

func CallInventoryService(req *protobuf.StockUpdateRequest) (*protobuf.StockResponse, error) {
	conn, err := grpc.Dial("inventory-service:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("unable to dial to port 50051 - Inventory service", err)
		return nil, err
	}
	defer conn.Close()
	client := protobuf.NewInventoryServiceClient(conn)

	resp, err1 := client.UpdateStock(context.Background(), req)

	if err1 != nil {
		log.Println("Unable to send data to Inventory service", err1)
		return nil, err1
	}
	return resp, nil

}
