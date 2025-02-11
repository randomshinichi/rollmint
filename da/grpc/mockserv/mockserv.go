package mockserv

import (
	"context"
	"os"

	tmlog "github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	grpcda "github.com/celestiaorg/rollmint/da/grpc"
	"github.com/celestiaorg/rollmint/da/mock"
	"github.com/celestiaorg/rollmint/store"
	"github.com/celestiaorg/rollmint/types"
	"github.com/celestiaorg/rollmint/types/pb/dalc"
	"github.com/celestiaorg/rollmint/types/pb/rollmint"
)

// GetServer creates and returns gRPC server instance.
func GetServer(kv store.KVStore, conf grpcda.Config, mockConfig []byte) *grpc.Server {
	logger := tmlog.NewTMLogger(os.Stdout)

	srv := grpc.NewServer()
	mockImpl := &mockImpl{}
	err := mockImpl.mock.Init([8]byte{}, mockConfig, kv, logger)
	if err != nil {
		logger.Error("failed to initialize mock DALC", "error", err)
		panic(err)
	}
	err = mockImpl.mock.Start()
	if err != nil {
		logger.Error("failed to start mock DALC", "error", err)
		panic(err)
	}
	dalc.RegisterDALCServiceServer(srv, mockImpl)
	return srv
}

type mockImpl struct {
	mock mock.DataAvailabilityLayerClient
}

func (m *mockImpl) SubmitBlock(_ context.Context, request *dalc.SubmitBlockRequest) (*dalc.SubmitBlockResponse, error) {
	var b types.Block
	err := b.FromProto(request.Block)
	if err != nil {
		return nil, err
	}
	resp := m.mock.SubmitBlock(&b)
	return &dalc.SubmitBlockResponse{
		Result: &dalc.DAResponse{
			Code:     dalc.StatusCode(resp.Code),
			Message:  resp.Message,
			DAHeight: resp.DAHeight,
		},
	}, nil
}

func (m *mockImpl) CheckBlockAvailability(_ context.Context, request *dalc.CheckBlockAvailabilityRequest) (*dalc.CheckBlockAvailabilityResponse, error) {
	resp := m.mock.CheckBlockAvailability(request.DAHeight)
	return &dalc.CheckBlockAvailabilityResponse{
		Result: &dalc.DAResponse{
			Code:    dalc.StatusCode(resp.Code),
			Message: resp.Message,
		},
		DataAvailable: resp.DataAvailable,
	}, nil
}

func (m *mockImpl) RetrieveBlocks(context context.Context, request *dalc.RetrieveBlocksRequest) (*dalc.RetrieveBlocksResponse, error) {
	resp := m.mock.RetrieveBlocks(request.DAHeight)
	blocks := make([]*rollmint.Block, len(resp.Blocks))
	for i := range resp.Blocks {
		blocks[i] = resp.Blocks[i].ToProto()
	}
	return &dalc.RetrieveBlocksResponse{
		Result: &dalc.DAResponse{
			Code:    dalc.StatusCode(resp.Code),
			Message: resp.Message,
		},
		Blocks: blocks,
	}, nil
}
