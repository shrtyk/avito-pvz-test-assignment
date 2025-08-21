package grpc

import (
	"context"

	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	pvz "github.com/shrtyk/avito-pvz-test-assignment/proto/pvz/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetPVZList(
	ctx context.Context,
	in *pvz.GetPVZListRequest,
) (*pvz.GetPVZListResponse, error) {
	pvzs, err := s.appService.GetAllPvzs(logger.ToCtx(ctx, s.logger))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pvz.GetPVZListResponse{
		Pvzs: toProtoFromDomainPvzs(pvzs),
	}, nil
}
