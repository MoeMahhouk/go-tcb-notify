package evaluator

import (
	"context"
	"testing"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/mocks"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEvaluator_FetchAllQuotes(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.QuoteStore)
		expectedCount int
		expectedError bool
	}{
		{
			name: "successful fetch with quotes",
			setupMocks: func(qs *mocks.QuoteStore) {
				quotes := []*models.RegistryQuote{
					{
						ServiceAddress: "0x1234567890123456789012345678901234567890",
						BlockNumber:    100,
						QuoteBytes:     []byte("quote1"),
						QuoteHash:      "hash1",
					},
					{
						ServiceAddress: "0x9876543210987654321098765432109876543210",
						BlockNumber:    200,
						QuoteBytes:     []byte("quote2"),
						QuoteHash:      "hash2",
					},
				}
				qs.EXPECT().
					GetActiveQuotes(mock.Anything).
					Return(quotes, nil).
					Once()
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "empty result",
			setupMocks: func(qs *mocks.QuoteStore) {
				qs.EXPECT().
					GetActiveQuotes(mock.Anything).
					Return([]*models.RegistryQuote{}, nil).
					Once()
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "database error",
			setupMocks: func(qs *mocks.QuoteStore) {
				qs.EXPECT().
					GetActiveQuotes(mock.Anything).
					Return(nil, assert.AnError).
					Once()
			},
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockEvaluationStore := mocks.NewEvaluationStore(t)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockQuoteStore)
			}

			// Create evaluator
			evaluator := &Evaluator{
				quoteStore:      mockQuoteStore,
				evaluationStore: mockEvaluationStore,
				config:          &config.EvaluateQuotes{},
				logger:          logrus.WithField("test", true),
			}

			// Execute
			quotes, err := evaluator.fetchAllQuotes(context.Background())

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, quotes, tt.expectedCount)
			}

			// Verify mocks
			mockQuoteStore.AssertExpectations(t)
		})
	}
}

func TestEvaluator_EvaluateQuotes(t *testing.T) {
	tests := []struct {
		name          string
		quotes        []*models.RegistryQuote
		setupMocks    func(*mocks.EvaluationStore)
		expectedStats EvaluationStats
	}{
		{
			name: "successful evaluation of multiple quotes",
			quotes: []*models.RegistryQuote{
				{
					ServiceAddress: "0x1234567890123456789012345678901234567890",
					BlockNumber:    100,
					QuoteBytes:     []byte("valid-quote"),
					QuoteHash:      "hash1",
				},
				{
					ServiceAddress: "0x9876543210987654321098765432109876543210",
					BlockNumber:    200,
					QuoteBytes:     []byte("invalid-quote"),
					QuoteHash:      "hash2",
				},
			},
			setupMocks: func(es *mocks.EvaluationStore) {
				// Mock GetLastEvaluation for status change detection (no previous evaluation)
				es.EXPECT().
					GetLastEvaluation(mock.Anything, mock.Anything, mock.Anything).
					Return("", "", assert.AnError).
					Twice() // Called for each quote

				// Mock successful storage for both evaluations
				es.EXPECT().
					StoreEvaluation(mock.Anything, mock.Anything).
					Return(nil).
					Twice()
			},
			expectedStats: EvaluationStats{
				Total: 2,
				// Note: Valid/Invalid counts depend on the actual quote verification
				// which we can't easily mock without refactoring the verifier
			},
		},
		{
			name: "evaluation storage failure",
			quotes: []*models.RegistryQuote{
				{
					ServiceAddress: "0x1234567890123456789012345678901234567890",
					BlockNumber:    100,
					QuoteBytes:     []byte("quote"),
					QuoteHash:      "hash1",
				},
			},
			setupMocks: func(es *mocks.EvaluationStore) {
				// Mock GetLastEvaluation for status change detection (no previous evaluation)
				es.EXPECT().
					GetLastEvaluation(mock.Anything, mock.Anything, mock.Anything).
					Return("", "", assert.AnError).
					Once()

				// Mock storage failure
				es.EXPECT().
					StoreEvaluation(mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			expectedStats: EvaluationStats{
				Total: 1,
				// Storage failure should be logged but not affect stats
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockEvaluationStore := mocks.NewEvaluationStore(t)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockEvaluationStore)
			}

			// Create evaluator with real verifier (for simplicity)
			evaluator := &Evaluator{
				quoteStore:      mockQuoteStore,
				evaluationStore: mockEvaluationStore,
				verifier:        &tdx.QuoteVerifier{}, // Real verifier for now
				config:          &config.EvaluateQuotes{},
				logger:          logrus.WithField("test", true),
			}

			// Execute
			stats := evaluator.evaluateQuotes(context.Background(), tt.quotes)

			// Assert
			assert.Equal(t, tt.expectedStats.Total, stats.Total)

			// Verify mocks
			mockEvaluationStore.AssertExpectations(t)
		})
	}
}
