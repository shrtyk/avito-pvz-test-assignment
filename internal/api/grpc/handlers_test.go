package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	mocks "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service/mocks"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	pvz "github.com/shrtyk/avito-pvz-test-assignment/proto/pvz/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetPVZList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log, _ := logger.NewTestLogger()

	sampleDomainPvzs := []*domain.Pvz{
		{
			Id:               uuid.New(),
			RegistrationDate: time.Now(),
			City:             "Test City",
		},
	}

	testCases := []struct {
		name          string
		setupMock     func(mockService *mocks.MockService)
		checkResponse func(t *testing.T, resp *pvz.GetPVZListResponse, err error)
	}{
		{
			name: "success",
			setupMock: func(mockService *mocks.MockService) {
				mockService.EXPECT().GetAllPvzs(mock.Anything).Return(sampleDomainPvzs, nil)
			},
			checkResponse: func(t *testing.T, resp *pvz.GetPVZListResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Len(t, resp.Pvzs, 1)
				assert.Equal(t, sampleDomainPvzs[0].Id.String(), resp.Pvzs[0].Id)
			},
		},
		{
			name: "success with empty list",
			setupMock: func(mockService *mocks.MockService) {
				mockService.EXPECT().GetAllPvzs(mock.Anything).Return([]*domain.Pvz{}, nil)
			},
			checkResponse: func(t *testing.T, resp *pvz.GetPVZListResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Len(t, resp.Pvzs, 0)
			},
		},
		{
			name: "error",
			setupMock: func(mockService *mocks.MockService) {
				mockService.EXPECT().GetAllPvzs(mock.Anything).Return(nil, errors.New("internal service error"))
			},
			checkResponse: func(t *testing.T, resp *pvz.GetPVZListResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, resp)

				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.Internal, st.Code())
				assert.Equal(t, "internal service error", st.Message())
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockService := mocks.NewMockService(t)
			tc.setupMock(mockService)

			s := Server{
				appService: mockService,
				logger:     log,
			}

			req := &pvz.GetPVZListRequest{}
			resp, err := s.GetPVZList(ctx, req)

			tc.checkResponse(t, resp, err)
		})
	}
}
