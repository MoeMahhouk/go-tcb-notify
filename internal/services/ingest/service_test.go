package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/internal/mocks"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIngester_ProcessEvent(t *testing.T) {
	tests := []struct {
		name          string
		event         Event
		setupMocks    func(*mocks.QuoteStore, *mocks.OffsetStore)
		expectedError bool
		validate      func(*testing.T, *mocks.QuoteStore)
	}{
		{
			name: "valid quote registration",
			event: Event{
				Type:        "TEEServiceRegistered",
				TeeAddress:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				RawQuote:    []byte{0x00, 0x01, 0x02}, // sample quote data
				BlockNumber: 100,
				LogIndex:    5,
				TxHash:      common.HexToHash("0xdeadbeef"),
				BlockTime:   time.Now(),
			},
			setupMocks: func(qs *mocks.QuoteStore, os *mocks.OffsetStore) {
				// Mock successful storage
				qs.EXPECT().
					StoreRawQuote(mock.Anything, mock.MatchedBy(func(q *models.RegistryQuote) bool {
						return q.ServiceAddress == "0x1234567890123456789012345678901234567890" &&
							q.BlockNumber == 100 &&
							len(q.QuoteBytes) > 0
					})).
					Return(nil).
					Once()
			},
			expectedError: false,
			validate: func(t *testing.T, qs *mocks.QuoteStore) {
				qs.AssertExpectations(t)
			},
		},
		{
			name: "invalidation event",
			event: Event{
				Type:        "TEEServiceInvalidated",
				TeeAddress:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				RawQuote:    nil, // No quote data for invalidation
				BlockNumber: 100,
				LogIndex:    5,
				TxHash:      common.HexToHash("0xdeadbeef"),
				BlockTime:   time.Now(),
			},
			setupMocks: func(qs *mocks.QuoteStore, os *mocks.OffsetStore) {
				// Mock successful invalidation storage
				qs.EXPECT().
					StoreInvalidation(mock.Anything, "0x1234567890123456789012345678901234567890", uint64(100), mock.Anything, "0x00000000000000000000000000000000000000000000000000000000deadbeef", uint32(5)).
					Return(nil).
					Once()
			},
			expectedError: false,
			validate: func(t *testing.T, qs *mocks.QuoteStore) {
				qs.AssertExpectations(t)
			},
		},
		{
			name: "storage failure",
			event: Event{
				Type:        "TEEServiceRegistered",
				TeeAddress:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				RawQuote:    []byte{0x00, 0x01, 0x02},
				BlockNumber: 100,
				LogIndex:    5,
				TxHash:      common.HexToHash("0xdeadbeef"),
				BlockTime:   time.Now(),
			},
			setupMocks: func(qs *mocks.QuoteStore, os *mocks.OffsetStore) {
				// Mock storage failure
				qs.EXPECT().
					StoreRawQuote(mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
			},
			expectedError: true,
			validate: func(t *testing.T, qs *mocks.QuoteStore) {
				qs.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockOffsetStore := mocks.NewOffsetStore(t)

			// Setup mocks BEFORE creating the ingester
			if tt.setupMocks != nil {
				tt.setupMocks(mockQuoteStore, mockOffsetStore)
			}

			// Create ingester with mocked dependencies
			ingester := &Ingester{
				quoteStore:  mockQuoteStore,
				offsetStore: mockOffsetStore,
				processor:   NewQuoteProcessor(),
			}

			// Process event
			err := ingester.processEvent(context.Background(), tt.event)

			// Check error expectation
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Additional validation
			if tt.validate != nil {
				tt.validate(t, mockQuoteStore)
			}
		})
	}
}

func TestIngester_SaveCheckpoint(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.OffsetStore)
		expectedError bool
	}{
		{
			name: "successful checkpoint save",
			setupMocks: func(os *mocks.OffsetStore) {
				os.EXPECT().
					SaveOffset(mock.Anything, ServiceName, uint64(100), uint32(5)).
					Return(nil).
					Once()
			},
			expectedError: false,
		},
		{
			name: "checkpoint save failure",
			setupMocks: func(os *mocks.OffsetStore) {
				os.EXPECT().
					SaveOffset(mock.Anything, ServiceName, uint64(100), uint32(5)).
					Return(assert.AnError).
					Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockOffsetStore := mocks.NewOffsetStore(t)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockOffsetStore)
			}

			// Create ingester with test state
			ingester := &Ingester{
				quoteStore:  mockQuoteStore,
				offsetStore: mockOffsetStore,
				lastBlock:   100,
				lastIdx:     5,
			}

			// Test checkpoint save
			err := ingester.saveCheckpoint(context.Background())

			// Check error expectation
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockOffsetStore.AssertExpectations(t)
		})
	}
}

func TestIngester_LoadCheckpoint(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.OffsetStore)
		expectedBlock uint64
		expectedIndex uint
		expectedError bool
	}{
		{
			name: "existing checkpoint found",
			setupMocks: func(os *mocks.OffsetStore) {
				os.EXPECT().
					LoadOffset(mock.Anything, ServiceName).
					Return(uint64(50), uint32(10), nil).
					Once()
			},
			expectedBlock: 50,
			expectedIndex: 10,
			expectedError: false,
		},
		{
			name: "no checkpoint found (first run)",
			setupMocks: func(os *mocks.OffsetStore) {
				os.EXPECT().
					LoadOffset(mock.Anything, ServiceName).
					Return(uint64(0), uint32(0), assert.AnError).
					Once()
			},
			expectedBlock: 0,
			expectedIndex: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockQuoteStore := mocks.NewQuoteStore(t)
			mockOffsetStore := mocks.NewOffsetStore(t)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockOffsetStore)
			}

			// Create ingester
			ingester := &Ingester{
				quoteStore:  mockQuoteStore,
				offsetStore: mockOffsetStore,
			}

			// Test load checkpoint
			err := ingester.loadCheckpoint(context.Background())

			// Check results
			if tt.expectedError {
				assert.Error(t, err)
				// On error, should default to 0,0
				assert.Equal(t, uint64(0), ingester.lastBlock)
				assert.Equal(t, uint(0), ingester.lastIdx)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBlock, ingester.lastBlock)
				assert.Equal(t, tt.expectedIndex, ingester.lastIdx)
			}

			// Verify mocks
			mockOffsetStore.AssertExpectations(t)
		})
	}
}

func TestQuoteProcessor_ProcessQuote(t *testing.T) {
	tests := []struct {
		name          string
		rawQuote      []byte
		expectedError bool
		validate      func(*testing.T, *models.RegistryQuote)
	}{
		{
			name:          "valid quote processing",
			rawQuote:      []byte{0x00, 0x01, 0x02, 0x03}, // sample quote
			expectedError: false,
			validate: func(t *testing.T, quote *models.RegistryQuote) {
				assert.NotEmpty(t, quote.QuoteHash)
				assert.Equal(t, uint32(4), quote.QuoteLength)
				assert.Equal(t, []byte{0x00, 0x01, 0x02, 0x03}, quote.QuoteBytes)
			},
		},
		{
			name:          "empty quote",
			rawQuote:      []byte{},
			expectedError: false,
			validate: func(t *testing.T, quote *models.RegistryQuote) {
				assert.NotEmpty(t, quote.QuoteHash) // Hash should still be calculated
				assert.Equal(t, uint32(0), quote.QuoteLength)
				assert.Empty(t, quote.QuoteBytes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewQuoteProcessor()

			quote, err := processor.ProcessQuote(tt.rawQuote)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, quote)
				if tt.validate != nil {
					tt.validate(t, quote)
				}
			}
		})
	}
}
