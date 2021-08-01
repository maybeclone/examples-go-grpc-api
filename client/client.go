package main

import (
	"bufio"
	"context"
	"google.golang.org/grpc"
	calculatorpub "grpc/exmple"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	cc, err := grpc.Dial("localhost:50001", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("dial err %v", err)
	}

	defer cc.Close()

	client := calculatorpub.NewCalculatorServiceClient(cc)
	_ = client
	//callSum(client)
	//countdownStreaming(client)
	//averageNumbers(client)
	//findMax(client)
	console(client)
}

func callSum(client calculatorpub.CalculatorServiceClient, a int32, b int32) {
	resp, _ := client.Sum(context.Background(), &calculatorpub.SumRequest{
		Num1: a,
		Num2: b,
	})

	log.Printf("Sum is %d", resp.GetResult())
}

func countdownStreaming(client calculatorpub.CalculatorServiceClient, num int32) {
	stream, err := client.Countdown(context.Background(), &calculatorpub.CountdownRequest{
		Number: num,
	})

	if err != nil {
		log.Fatalf("get stream failed %v", stream)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		log.Printf("Countdown %v", resp.GetResult())
	}

	log.Println("Countdown finished")
}

func averageNumbers(client calculatorpub.CalculatorServiceClient, scanner *bufio.Scanner) {
	stream, err := client.Average(context.Background())
	if err != nil {
		return
	}

	for {
		scanner.Scan()
		a := scanner.Text()
		if a == "" {
			break
		}
		num, _ := strconv.Atoi(a)
		err := stream.Send(&calculatorpub.AverageRequest{
			Number: int32(num),
		})
		if err != nil {
			break
		}
		log.Printf("send num %v", num)
		time.Sleep(1000 * time.Millisecond)
	}
	result, err := stream.CloseAndRecv()
	if err != nil {
		return
	}
	log.Printf("Average is %v", result.GetResult())
}

func findMax(client calculatorpub.CalculatorServiceClient, scanner *bufio.Scanner) {
	stream, err := client.FindMax(context.Background())
	if err != nil {
		return
	}

	wait := make(chan bool)

	go func() {
		for {
			scanner.Scan()
			a := scanner.Text()
			if a == "" {
				break
			}
			num, _ := strconv.Atoi(a)
			err := stream.Send(&calculatorpub.FindMaxRequest{
				Number: int32(num),
			})
			if err != nil {
				break
			}
		}
		err := stream.CloseSend()
		if err != nil {
			return
		}
	}()

	go func() {
		for {
			recv, err := stream.Recv()
			if err == io.EOF {
				close(wait)
				break
			}
			if err != nil {
				return
			}
			log.Printf("A max found %v", recv.GetResult())
		}
	}()

	<-wait
}

func menu() {
	println("======DEMO======")
	println("1. Unary (Sum num1 & num 2)")
	println("2. Server Streaming (Countdown num)")
	println("3. Client Streaming (Average)")
	println("4. Bidirectional Streaming (Find max)")
	println("5. Exit")
	println("======DEMO======")
	print("Your choice...")
}

func console(client calculatorpub.CalculatorServiceClient) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		menu()
		scanner.Scan()
		sel, _ := strconv.Atoi(scanner.Text())

		switch sel {
		case 1:
			{
				println("Num1: ")
				scanner.Scan()
				a := scanner.Text()
				println("Num2: ")
				scanner.Scan()
				b := scanner.Text()
				num1, _ := strconv.Atoi(a)
				num2, _ := strconv.Atoi(b)
				callSum(client, int32(num1), int32(num2))
				break
			}
		case 2:
			{
				println("Countdown number: ")
				scanner.Scan()
				a := scanner.Text()
				num, _ := strconv.Atoi(a)
				countdownStreaming(client, int32(num))
				scanner.Scan()
				break
			}
		case 3:
			{
				println("Input numbers: ")
				averageNumbers(client, scanner)
				break
			}
		case 4:
			{
				findMax(client, scanner)
			}
		case 5:
			{
				println("Exit")
				return
			}
		}
		scanner.Scan()
	}
}
