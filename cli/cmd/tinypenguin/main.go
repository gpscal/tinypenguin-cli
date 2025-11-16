package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "example.com/tinypenguin/pkg/pb"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement tinypenguin.TaskService
type server struct {
	pb.UnimplementedTaskServiceServer
}

// ExecuteTask implements tinypenguin.TaskService.ExecuteTask
func (s *server) ExecuteTask(req *pb.ExecuteTaskRequest, stream pb.TaskService_ExecuteTaskServer) error {
	log.Printf("Received task request: %s", req.Query)
	
	// Create task started response
	taskStarted := &pb.TaskStarted{
		TaskId: "task-" + fmt.Sprintf("%d", os.Getpid()),
	}
	
	response := &pb.ExecuteTaskResponse{
		Response: &pb.ExecuteTaskResponse_TaskStarted{
			TaskStarted: taskStarted,
		},
	}
	
	if err := stream.Send(response); err != nil {
		return err
	}
	
	return nil
}

// CancelTask implements tinypenguin.TaskService.CancelTask
func (s *server) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	log.Printf("Received cancel request for task: %s", req.TaskId)
	
	return &pb.CancelTaskResponse{
		Success: true,
	}, nil
}

// ListTasks implements tinypenguin.TaskService.ListTasks
func (s *server) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	log.Printf("Received list tasks request")
	
	// Return empty task list for now
	return &pb.ListTasksResponse{
		Tasks:          []*pb.Task{},
		NextPageToken:  "",
	}, nil
}

func main() {
	flag.Parse()
	
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	
	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &server{})
	
	// Register reflection service on gRPC server.
	reflection.Register(s)
	
	log.Printf("tinypenguin server listening at %v", lis.Addr())
	
	// Start the server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}