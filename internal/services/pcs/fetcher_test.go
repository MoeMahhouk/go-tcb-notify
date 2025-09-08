package pcs

import (
	"context"
	"testing"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/mocks"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetcher_CreateTCBUpdateAlert(t *testing.T) {
	tests := []struct {
		name          string
		fmspc         string
		oldEval       uint32
		newEval       uint32
		affectedCount int64
		setupMocks    func(*mocks.QuoteStore, *mocks.AlertStore)
		expectedError bool
	}{
		{
			name:          "successful alert creation",
			fmspc:         "00906EA10000",
			oldEval:       5,
			newEval:       10,
			affectedCount: 15,
			setupMocks: func(qs *mocks.QuoteStore, as *mocks.AlertStore) {
				qs.EXPECT().
					CountAffectedQuotes(mock.Anything, "00906EA10000").
					Return(int64(15), nil).
					Once()

				as.EXPECT().
					CreateAlert(mock.Anything, mock.MatchedBy(func(alert *models.TCBAlert) bool {
						return alert.FMSPC == "00906EA10000" &&
							alert.OldEvalNumber == 5 &&
							alert.NewEvalNumber == 10 &&
							alert.AffectedQuotesCount == 15
					})).
					Return(nil).
					Once()
			},
			expectedError: false,
		},
		{
			name:          "count quotes failure",
			fmspc:         "00906EA10000",
			oldEval:       5,
			newEval:       10,
			affectedCount: 0,
			setupMocks: func(qs *mocks.QuoteStore, as *mocks.AlertStore) {
				qs.EXPECT().
					CountAffectedQuotes(mock.Anything, "00906EA10000").
					Return(int64(0), assert.AnError).
					Once()

				as.EXPECT().
					CreateAlert(mock.Anything, mock.MatchedBy(func(alert *models.TCBAlert) bool {
						return alert.AffectedQuotesCount == 0 // Should use 0 when count fails
					})).
					Return(nil).
					Once()
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTCBStore := mocks.NewTCBStore(t)
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockAlertStore := mocks.NewAlertStore(t)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockQuoteStore, mockAlertStore)
			}

			// Create fetcher with mocked dependencies
			fetcher := &Fetcher{
				tcbStore:   mockTCBStore,
				quoteStore: mockQuoteStore,
				alertStore: mockAlertStore,
				config:     &config.PCS{},
				logger:     logrus.WithField("test", true),
			}

			// Execute
			err := fetcher.createTCBUpdateAlert(context.Background(), tt.fmspc, tt.oldEval, tt.newEval)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockQuoteStore.AssertExpectations(t)
			mockAlertStore.AssertExpectations(t)
		})
	}
}

func TestFetcher_StoreFMSPC(t *testing.T) {
	tests := []struct {
		name          string
		fmspc         models.FMSPCResponse
		setupMocks    func(*mocks.TCBStore)
		expectedError bool
	}{
		{
			name: "successful FMSPC storage",
			fmspc: models.FMSPCResponse{
				FMSPC:    "00906ea10000",
				Platform: "E5",
			},
			setupMocks: func(ts *mocks.TCBStore) {
				ts.EXPECT().
					StoreFMSPC(mock.Anything, "00906EA10000", "E5").
					Return(nil).
					Once()
			},
			expectedError: false,
		},
		{
			name: "FMSPC with empty platform",
			fmspc: models.FMSPCResponse{
				FMSPC:    "00906ea10000",
				Platform: "",
			},
			setupMocks: func(ts *mocks.TCBStore) {
				ts.EXPECT().
					StoreFMSPC(mock.Anything, "00906EA10000", "ALL").
					Return(nil).
					Once()
			},
			expectedError: false,
		},
		{
			name: "storage failure",
			fmspc: models.FMSPCResponse{
				FMSPC:    "00906ea10000",
				Platform: "E5",
			},
			setupMocks: func(ts *mocks.TCBStore) {
				ts.EXPECT().
					StoreFMSPC(mock.Anything, "00906EA10000", "E5").
					Return(assert.AnError).
					Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTCBStore := mocks.NewTCBStore(t)
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockAlertStore := mocks.NewAlertStore(t)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockTCBStore)
			}

			// Create fetcher
			fetcher := &Fetcher{
				tcbStore:   mockTCBStore,
				quoteStore: mockQuoteStore,
				alertStore: mockAlertStore,
				logger:     logrus.WithField("test", true),
			}

			// Execute
			err := fetcher.storeFMSPC(context.Background(), tt.fmspc)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockTCBStore.AssertExpectations(t)
		})
	}
}

func TestFetcher_GetCurrentEvalNumber(t *testing.T) {
	tests := []struct {
		name           string
		fmspc          string
		setupMocks     func(*mocks.TCBStore)
		expectedEval   uint32
		expectedError  bool
		expectNotFound bool
	}{
		{
			name:  "successful get",
			fmspc: "00906EA10000",
			setupMocks: func(ts *mocks.TCBStore) {
				ts.EXPECT().
					GetCurrentEvalNumber(mock.Anything, "00906EA10000").
					Return(uint32(10), nil).
					Once()
			},
			expectedEval:  10,
			expectedError: false,
		},
		{
			name:  "not found",
			fmspc: "00906EA10000",
			setupMocks: func(ts *mocks.TCBStore) {
				ts.EXPECT().
					GetCurrentEvalNumber(mock.Anything, "00906EA10000").
					Return(uint32(0), assert.AnError).
					Once()
			},
			expectedEval:   0,
			expectedError:  true,
			expectNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTCBStore := mocks.NewTCBStore(t)
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockAlertStore := mocks.NewAlertStore(t)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockTCBStore)
			}

			// Create fetcher
			fetcher := &Fetcher{
				tcbStore:   mockTCBStore,
				quoteStore: mockQuoteStore,
				alertStore: mockAlertStore,
			}

			// Execute
			evalNum, err := fetcher.getCurrentEvalNumber(context.Background(), tt.fmspc)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				// Note: We're not checking for specific ErrNotFound since the mock returns assert.AnError
				// In real implementation, the error conversion happens in the service layer
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedEval, evalNum)

			// Verify mocks
			mockTCBStore.AssertExpectations(t)
		})
	}
}
