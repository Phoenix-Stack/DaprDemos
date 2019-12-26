package main

import (
	"context"
	"fmt"
	"os"

	pb "github.com/dapr/go-sdk/dapr"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"

	"daprdemos/golang/shoppingCartForJava/protos/daprexamples"
)

func main() {
	// Get the Dapr port and create a connection
	daprPort := os.Getenv("DAPR_GRPC_PORT")
	daprAddress := fmt.Sprintf("localhost:%s", daprPort)
	conn, err := grpc.Dial(daprAddress, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Create the client
	client := pb.NewDaprClient(conn)

	createOrderRequest := &daprexamples.CreateOrderRequest{
		ProductID:  "095d1f49-41c8-4716-81f0-35e05303faea",
		Amount:     20,
		CustomerID: "0d158a88-73de-42e5-87c7-fdbc00bdc5f9",
	}
	createOrderRequestData, err := ptypes.MarshalAny(createOrderRequest)
	if err != nil {
		fmt.Println(createOrderRequestData)
	} else {
		fmt.Println(createOrderRequestData)
	}

	// Invoke a method called MyMethod on another Dapr enabled service with id client
	response, err := client.InvokeService(context.Background(), &pb.InvokeServiceEnvelope{
		Id:     "OrderService",
		Data:   createOrderRequestData,
		Method: "createOrder",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	createOrderResponse := &daprexamples.CreateOrderResponse{}

	if err := proto.Unmarshal(response.Data.Value, createOrderResponse); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(createOrderResponse.Succeed)

	if !createOrderResponse.Succeed {
		//下单失败
		return
	}

	storageReduceData := &daprexamples.StorageReduceData{
		ProductID: createOrderRequest.ProductID,
		Amount:    createOrderRequest.Amount,
	}
	storageReduceDataData, err := ptypes.MarshalAny(storageReduceData)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = client.PublishEvent(context.Background(), &pb.PublishEventEnvelope{
		Topic: "Storage.Reduce",
		Data:  storageReduceDataData,
	})

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Published message!")
	}
}
