package internal

import (
	"context"
	"iter"
	"math/rand"

	"github.com/twistedogic/grpcutil"
	"github.com/twistedogic/grpcutil/internal/examplepb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	examplepb.UnimplementedPositionServiceServer
}

func (s *server) Register(srv *grpc.Server) {
	examplepb.RegisterPositionServiceServer(srv, s)
}

func (s *server) UpdatePosition(ctx context.Context, req *examplepb.UpdatePositionRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) generatePositions(size int) iter.Seq[*examplepb.Position] {
	return func(yield func(*examplepb.Position) bool) {
		for i := 0; i < size; i++ {
			pos := &examplepb.Position{X: rand.Int31(), Y: rand.Int31()}
			if !yield(pos) {
				return
			}
		}
	}
}

func (s *server) TrackPosition(req *examplepb.TrackPositionRequest, stream examplepb.PositionService_TrackPositionServer) error {
	return grpcutil.ToStream(stream, s.generatePositions(10))
}
