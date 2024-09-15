package main

import (
	"context"
	"fmt"
	"log"
	"net"

	proto "github.com/logeshwarann-dev/task-service/proto" // Import generated proto

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

const port = ":50052"

type TaskServiceServer struct {
	proto.UnimplementedTaskServiceServer
	mongoClient *mongo.Client
}

type Task struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID string             `bson:"user_id"`
	Title  string             `bson:"title"`
	Done   bool               `bson:"done"`
}

func (s *TaskServiceServer) AddTask(ctx context.Context, req *proto.AddTaskRequest) (*proto.AddTaskResponse, error) {
	collection := s.mongoClient.Database("task_db").Collection("tasks")
	task := Task{UserID: req.UserId, Title: req.Title, Done: false}
	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to add task: %v", err)
	}
	return &proto.AddTaskResponse{
		Success: true,
		TaskId:  fmt.Sprintf("%d", task.ID), // Return the TaskID
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// MongoDB connection
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	proto.RegisterTaskServiceServer(grpcServer, &TaskServiceServer{mongoClient: client})

	log.Printf("Task service listening on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
