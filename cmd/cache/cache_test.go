package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

// mockCacheServiceClient implements CacheServiceClient for testing
type mockCacheServiceClient struct {
	mock.Mock
}

func (m *mockCacheServiceClient) Clear(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*connect.Response[emptypb.Empty]), args.Error(1)
}

func (m *mockCacheServiceClient) Cache(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*connect.Response[emptypb.Empty]), args.Error(1)
}

func TestInitDependencies(t *testing.T) {
	t.Run("initializes client when nil", func(t *testing.T) {
		o := &options{
			addr: "http://test-server:8080",
		}

		err := o.initDependencies()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if o.cacheServiceClient == nil {
			t.Error("cacheServiceClient should be initialized")
		}
	})

	t.Run("keeps existing client", func(t *testing.T) {
		mockClient := &mockCacheServiceClient{}
		o := &options{
			addr:               "http://test-server:8080",
			cacheServiceClient: mockClient,
		}

		err := o.initDependencies()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if o.cacheServiceClient != mockClient {
			t.Error("cacheServiceClient should not be replaced when already set")
		}
	})
}
func TestOptions_AddFlags(t *testing.T) {
	// Setup
	o := &options{}
	cmd := &cobra.Command{}

	// Execute
	o.AddFlags(cmd)

	// Assert
	flags := cmd.Flags()

	// Test clear flag
	clearFlag := flags.Lookup("clear")
	assert.NotNil(t, clearFlag)
	assert.Equal(t, "false", clearFlag.DefValue)
	assert.Equal(t, "Clear all cached graph data", clearFlag.Usage)

	// Test addr flag
	addrFlag := flags.Lookup("addr")
	assert.NotNil(t, addrFlag)
	assert.Equal(t, "http://localhost:8089", addrFlag.DefValue)
	assert.Equal(t, "Address of the minefield server", addrFlag.Usage)
}

func TestOptions_clearCache(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(*mockCacheServiceClient)
		wantErr bool
	}{
		{
			name: "successful clear",
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Clear", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, nil)
			},
			wantErr: false,
		},
		{
			name: "clear fails",
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Clear", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, errors.New("clear failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCacheServiceClient{}
			tt.mockFn(mockClient)

			o := &options{
				cacheServiceClient: mockClient,
			}

			err := o.clearCache(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestOptions_populateCache(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(*mockCacheServiceClient)
		wantErr bool
	}{
		{
			name: "successful populate",
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Cache", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, nil)
			},
			wantErr: false,
		},
		{
			name: "populate fails",
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Cache", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, errors.New("cache failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCacheServiceClient{}
			tt.mockFn(mockClient)

			o := &options{
				cacheServiceClient: mockClient,
			}

			err := o.populateCache(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestNew(t *testing.T) {
	cmd := New()

	// Test command properties
	assert.Equal(t, "cache", cmd.Use)
	assert.Equal(t, "Cache all nodes or remove existing cache", cmd.Short)
	assert.True(t, cmd.DisableAutoGenTag)

	// Test flags existence and defaults
	clearFlag := cmd.Flags().Lookup("clear")
	assert.NotNil(t, clearFlag)
	assert.Equal(t, "false", clearFlag.DefValue)

	addrFlag := cmd.Flags().Lookup("addr")
	assert.NotNil(t, addrFlag)
	assert.Equal(t, "http://localhost:8089", addrFlag.DefValue)
}
func TestOptions_Run(t *testing.T) {
	tests := []struct {
		name    string
		clear   bool
		mockFn  func(*mockCacheServiceClient)
		wantErr bool
	}{
		{
			name:  "successful cache populate",
			clear: false,
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Cache", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "successful cache clear",
			clear: true,
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Clear", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "cache populate error",
			clear: false,
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Cache", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, fmt.Errorf("cache error"))
			},
			wantErr: true,
		},
		{
			name:  "cache clear error",
			clear: true,
			mockFn: func(m *mockCacheServiceClient) {
				m.On("Clear", mock.Anything, mock.Anything).
					Return(&connect.Response[emptypb.Empty]{}, fmt.Errorf("clear error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCacheServiceClient{}
			tt.mockFn(mockClient)

			o := &options{
				clear:              tt.clear,
				cacheServiceClient: mockClient,
			}

			// Create a cobra command with context
			cmd := &cobra.Command{}
			cmd.SetContext(context.Background())

			err := o.Run(cmd, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
