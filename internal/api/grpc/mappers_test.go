package grpc

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestToProtoFromDomainPvzs(t *testing.T) {
	t.Parallel()

	testUUID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name       string
		domainPvzs []*domain.Pvz
		wantLen    int
	}{
		{
			name: "multiple pvzs",
			domainPvzs: []*domain.Pvz{
				{
					Id:               testUUID,
					RegistrationDate: testTime,
					City:             "Moscow",
				},
				{
					Id:               uuid.New(),
					RegistrationDate: time.Now().Add(time.Hour),
					City:             "Kazan",
				},
			},
			wantLen: 2,
		},
		{
			name:       "empty slice",
			domainPvzs: []*domain.Pvz{},
			wantLen:    0,
		},
		{
			name:       "nil slice",
			domainPvzs: nil,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := toProtoFromDomainPvzs(tt.domainPvzs)

			assert.Len(t, got, tt.wantLen)

			if tt.wantLen > 0 {
				assert.Equal(t, tt.domainPvzs[0].Id.String(), got[0].Id)
				assert.Equal(t, string(tt.domainPvzs[0].City), got[0].City)
				assert.Equal(t, tt.domainPvzs[0].RegistrationDate.Unix(), got[0].RegistrationDate.Seconds)
			}
		})
	}
}
