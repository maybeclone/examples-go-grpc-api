package main

import (
	"context"
	"google.golang.org/grpc"
	calculatorpub "grpc/exmple"
	"io"
	"log"
	"net"
	"time"
)

type server struct{}

func (*server) Subtract(ctx context.Context, request *calculatorpub.SubtractRequest) (*calculatorpub.SubtractResponse, error) {
	log.Println("[Subtract] called")
	log.Printf("Subtract %v", request)
	resp := &calculatorpub.SubtractResponse{
		Result: request.GetNum1() - request.GetNum2(),
	}
	return resp, nil
}

func (*server) Sum(context context.Context, req *calculatorpub.SumRequest) (*calculatorpub.SumResponse, error) {
	log.Println("[Sum] called")
	log.Printf("Sum payload %v", req)
	resp := &calculatorpub.SumResponse{
		Result: req.GetNum1() + req.GetNum2(),
	}
	return resp, nil
}

func (*server) Countdown(req *calculatorpub.CountdownRequest, stream calculatorpub.CalculatorService_CountdownServer) error {
	log.Println("[Countdown] called")
	i := req.GetNumber()
	for i >= 0 {
		err := stream.Send(&calculatorpub.CountdownResponse{
			Result: i,
		})
		i--
		time.Sleep(1000 * time.Millisecond)
		if err != nil {
			log.Fatalf("countdown failed %v", err)
		}
	}
	return nil
}

func (*server) Average(stream calculatorpub.CalculatorService_AverageServer) error {
	log.Println("[Average] called")
	var count = int32(0)
	var total = int32(0)
	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			err := stream.SendAndClose(&calculatorpub.AverageResponse{
				Result: float32(total / count),
			})
			if err != nil {
				log.Fatalf("closing stream err %v", err)
				return err
			}
			break
		}
		if err != nil {
			return err
		}
		log.Printf("recv number %v", recv.GetNumber())
		total += recv.GetNumber()
		count++
	}
	return nil
}

func (*server) FindMax(stream calculatorpub.CalculatorService_FindMaxServer) error {
	log.Println("[FindMax] called")

	var max = int32(0)

	for {
		recv, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}
		log.Printf("recv a num %v", recv.GetNumber())
		if recv.GetNumber() > max {
			max = recv.GetNumber()
			log.Printf("finding a new max %v", max)
			err := stream.Send(&calculatorpub.FindMaxResponse{
				Result: max,
			})
			if err != nil {
				return err
			}
		}
	}
}

func main() {

	lis, err := net.Listen("tcp", "0.0.0.0:50001")

	if err != nil {
		log.Fatalf("create listen err %v", err)
	}

	s := grpc.NewServer()

	calculatorpub.RegisterCalculatorServiceServer(s, &server{})

	log.Println("Server in running...")
	err = s.Serve(lis)

	if err != nil {
		log.Fatalf("serve listen err %v", err)
	}

}
