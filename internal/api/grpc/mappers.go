package grpc

import (
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	pvz "github.com/shrtyk/avito-pvz-test-assignment/proto/pvz/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoFromDomainPvzs(domainPvzs []*domain.Pvz) []*pvz.PVZ {
	res := make([]*pvz.PVZ, len(domainPvzs))
	for i, p := range domainPvzs {
		res[i] = &pvz.PVZ{
			Id: p.Id.String(),
			RegistrationDate: &timestamppb.Timestamp{
				Seconds: p.RegistrationDate.Unix(),
			},
			City: string(p.City),
		}
	}

	return res
}
