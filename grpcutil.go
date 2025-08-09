package grpcutil

import (
	"context"
	"io"
	"iter"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type StreamReciever[T any] interface {
	Recv() (T, error)
}

type StreamSender[T any] interface {
	Send(T) error
}

func FromStream[T any](stream StreamReciever[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for {
			item, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if !yield(item, err) {
				return
			}
		}
	}
}

type ResultError[T any] struct {
	item T
	err  error
}

func (re ResultError[T]) Item() T { return re.item }

func (re ResultError[T]) Err() error { return re.err }

func FromStreamChan[T any](stream StreamReciever[T]) <-chan ResultError[T] {
	ch := make(chan ResultError[T])
	go func() {
		defer close(ch)
		for {
			item, err := stream.Recv()
			if err == io.EOF {
				return
			}
			ch <- ResultError[T]{item: item, err: err}
			if err != nil {
				return
			}
		}
	}()
	return ch
}

func ToStream[T any](stream StreamSender[T], items iter.Seq[T]) error {
	for item := range items {
		if err := stream.Send(item); err != nil {
			return err
		}
	}
	return nil
}

type GrpcServerRegisterer interface {
	Register(*grpc.Server)
}

func BufconnClientConn(ctx context.Context, size int, registerer GrpcServerRegisterer) (*grpc.Server, grpc.ClientConnInterface) {
	lis := bufconn.Listen(size)
	s := grpc.NewServer()
	registerer.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	return s, conn
}
