package internal

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twistedogic/grpcutil"
	"github.com/twistedogic/grpcutil/internal/examplepb"
)

func Test_FromStream(t *testing.T) {
	s, conn := grpcutil.BufconnClientConn(t.Context(), 1024*1024, &server{})
	defer s.Stop()
	client := examplepb.NewPositionServiceClient(conn)
	stream, err := client.TrackPosition(t.Context(), &examplepb.TrackPositionRequest{})
	require.NoError(t, err)
	for position, err := range grpcutil.FromStream[*examplepb.Position](stream) {
		require.NoError(t, err)
		require.NotNil(t, position)
	}
}

func Benchmark_FromStream(b *testing.B) {
	s, conn := grpcutil.BufconnClientConn(b.Context(), 1024*1024, &server{})
	defer s.Stop()
	client := examplepb.NewPositionServiceClient(conn)
	b.Run("FromStream", func(b *testing.B) {
		for b.Loop() {
			stream, err := client.TrackPosition(context.TODO(), &examplepb.TrackPositionRequest{})
			require.NoError(b, err)
			results := grpcutil.FromStream[*examplepb.Position](stream)
			for position, err := range results {
				require.NoError(b, err)
				require.NotNil(b, position)
			}
		}
	})
	b.Run("FromStreamChan", func(b *testing.B) {
		for b.Loop() {
			stream, err := client.TrackPosition(context.TODO(), &examplepb.TrackPositionRequest{})
			require.NoError(b, err)
			results := grpcutil.FromStreamChan[*examplepb.Position](stream)
			for result := range results {
				require.NoError(b, result.Err())
				require.NotNil(b, result.Item())
			}
		}
	})
	b.Run("Native", func(b *testing.B) {
		for b.Loop() {
			stream, err := client.TrackPosition(context.TODO(), &examplepb.TrackPositionRequest{})
			require.NoError(b, err)
			for {
				position, err := stream.Recv()
				if err == io.EOF {
					break
				}
				require.NoError(b, err)
				require.NotNil(b, position)
			}
		}
	})
}
