package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"

	"daprdemos/golang/client/protos/productlist_v1"
	"daprdemos/golang/server/protos/shoppingCart"

	pbDapr "github.com/dapr/go-sdk/dapr"
	pb "github.com/dapr/go-sdk/daprclient"
	"google.golang.org/grpc"
)

// server is our user app
type server struct {
	productIDs []string
}

func main() {
	// create listiner
	lis, err := net.Listen("tcp", ":4001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	s := grpc.NewServer()
	pb.RegisterDaprClientServer(s, &server{})

	fmt.Println("Client starting...")

	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) AddProduct(addProductRequest *shoppingCart.AddProductRequest) *shoppingCart.AddProductResponse {
	s.productIDs = append(s.productIDs, addProductRequest.ProductID)
	return &shoppingCart.AddProductResponse{Succeed: true}
}

func (s *server) GetShoppingCart() (result shoppingCart.GetShoppingCartResponse) {
	result.ProductID = s.productIDs
	return
}

// This method gets invoked when a remote service has called the app through Dapr
// The payload carries a Method to identify the method, a set of metadata properties and an optional payload
func (s *server) OnInvoke(ctx context.Context, in *pb.InvokeEnvelope) (*any.Any, error) {
	fmt.Println(fmt.Sprintf("Got invoked by %s with: %s", in.Method, string(in.Data.Value)))

	switch in.Method {
	case "AddProduct":
		getAllProducts()
		addProductRequest := &shoppingCart.AddProductRequest{}
		if err := proto.Unmarshal(in.Data.Value, addProductRequest); err != nil {
			fmt.Println(err)
			return nil, err
		}
		addProductResponse := s.AddProduct(addProductRequest)
		any, err := ptypes.MarshalAny(addProductResponse)
		return any, err
	case "GetShoppingCart":
		getShoppingCartResponse := s.GetShoppingCart()
		any, err := ptypes.MarshalAny(&getShoppingCartResponse)
		return any, err
	}
	return &any.Any{}, nil
}

// Dapr will call this method to get the list of topics the app wants to subscribe to. In this example, we are telling Dapr
// To subscribe to a topic named TopicA
func (s *server) GetTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.GetTopicSubscriptionsEnvelope, error) {
	return &pb.GetTopicSubscriptionsEnvelope{
		Topics: []string{"TopicA"},
	}, nil
}

// Dapper will call this method to get the list of bindings the app will get invoked by. In this example, we are telling Dapr
// To invoke our app with a binding named storage
func (s *server) GetBindingsSubscriptions(ctx context.Context, in *empty.Empty) (*pb.GetBindingsSubscriptionsEnvelope, error) {
	return &pb.GetBindingsSubscriptionsEnvelope{
		Bindings: []string{"storage"},
	}, nil
}

// This method gets invoked every time a new event is fired from a registerd binding. The message carries the binding name, a payload and optional metadata
func (s *server) OnBindingEvent(ctx context.Context, in *pb.BindingEventEnvelope) (*pb.BindingResponseEnvelope, error) {
	fmt.Println("Invoked from binding")
	return &pb.BindingResponseEnvelope{}, nil
}

// This method is fired whenever a message has been published to a topic that has been subscribed. Dapr sends published messages in a CloudEvents 0.3 envelope.
func (s *server) OnTopicEvent(ctx context.Context, in *pb.CloudEventEnvelope) (*empty.Empty, error) {
	fmt.Println("Topic message arrived")
	return &empty.Empty{}, nil
}

func getAllProducts() {
	daprPort := os.Getenv("DAPR_GRPC_PORT")
	daprAddress := fmt.Sprintf("localhost:%s", daprPort)
	conn, err := grpc.Dial(daprAddress, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Create the client
	client := pbDapr.NewDaprClient(conn)

	//获取产品列表
	fmt.Println("获取产品列表")
	productListRequest := &productlist_v1.ProductListRequest{}
	data, err := ptypes.MarshalAny(productListRequest)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(data)
	}
	response, err := client.InvokeService(context.Background(), &pbDapr.InvokeServiceEnvelope{
		Id:     "productService",
		Data:   data,
		Method: "GetAllProducts",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	productList := &productlist_v1.ProductList{}
	if err := proto.Unmarshal(response.Data.Value, productList); err != nil {
		fmt.Println(err)
		return
	}
	for _, product := range productList.Results {
		fmt.Println(product.ID)
	}
}
