package graph

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunGraphVisualizer(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		ids         *roaring.Bitmap
		query       string
		addr        string
		expectError bool
	}{
		{
			name: "Successful visualization",
			setupMock: func(ms *MockStorage) {
				node := &Node{ID: 1, Name: "TestNode", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node)
				node2 := &Node{ID: 2, Name: "TestNode2", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node2)
				node3 := &Node{ID: 3, Name: "TestNode3", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node3)
				node4 := &Node{ID: 4, Name: "TestNode4", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node4)
				node.SetDependency(ms, node2)
				node2.SetDependency(ms, node3)
				node3.SetDependency(ms, node4)
			},
			ids:         roaring.BitmapOf(1),
			query:       "test query",
			addr:        "8081",
			expectError: false,
		},
		{
			name:        "Empty bitmap",
			setupMock:   func(ms *MockStorage) {},
			ids:         roaring.New(),
			query:       "empty query",
			addr:        "8082",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultServeMux = http.NewServeMux()

			mockStorage := NewMockStorage()
			tt.setupMock(mockStorage)

			// Create a test server
			testServer := httptest.NewServer(nil)
			defer testServer.Close()

			server := &http.Server{
				Addr: testServer.Listener.Addr().String(),
			}

			shutdown, err := RunGraphVisualizer(mockStorage, tt.ids, tt.query, server)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, shutdown)

				// Test the shutdown function
				done := make(chan struct{})
				go func() {
					shutdown()
					close(done)
				}()

				select {
				case <-done:
					// Shutdown completed successfully
				case <-time.After(5 * time.Second):
					t.Fatal("Shutdown timed out")
				}
			}
		})
	}
}

func TestGraphQuery(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		ids         *roaring.Bitmap
		query       string
		expectError bool
		nodeCount   int
		linkCount   int
	}{
		{
			name: "Single node graph",
			setupMock: func(ms *MockStorage) {
				node := &Node{ID: 1, Name: "Node1", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node)
			},
			ids:         roaring.BitmapOf(1),
			query:       "single node",
			expectError: false,
			nodeCount:   1,
			linkCount:   0,
		},
		{
			name: "Two connected nodes",
			setupMock: func(ms *MockStorage) {
				node1 := &Node{ID: 1, Name: "Node1", Children: roaring.BitmapOf(2), Parents: roaring.New()}
				node2 := &Node{ID: 2, Name: "Node2", Children: roaring.New(), Parents: roaring.BitmapOf(1)}
				ms.SaveNode(node1)
				ms.SaveNode(node2)
			},
			ids:         roaring.BitmapOf(1, 2),
			query:       "two nodes",
			expectError: false,
			nodeCount:   2,
			linkCount:   1,
		},
		{
			name:        "Empty bitmap",
			setupMock:   func(ms *MockStorage) {},
			ids:         roaring.New(),
			query:       "empty graph",
			expectError: false,
			nodeCount:   0,
			linkCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultServeMux = http.NewServeMux()

			mockStorage := NewMockStorage()
			tt.setupMock(mockStorage)

			graph, err := graphQuery(mockStorage, tt.ids, tt.query)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, graph)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, graph)
				assert.Equal(t, tt.nodeCount, len(graph.MultiSeries[0].Data.([]opts.GraphNode)))
				assert.Equal(t, tt.linkCount, len(graph.MultiSeries[0].Links.([]opts.GraphLink)))
			}
		})
	}
}

func TestGraphVisualizerHTTPHandler(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*MockStorage)
		ids          *roaring.Bitmap
		query        string
		expectedCode int
	}{
		{
			name: "Successful render",
			setupMock: func(ms *MockStorage) {
				node := &Node{ID: 1, Name: "TestNode", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node)
				node2 := &Node{ID: 2, Name: "TestNode2", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node2)
				node3 := &Node{ID: 3, Name: "TestNode3", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node3)
				node4 := &Node{ID: 4, Name: "TestNode4", Children: roaring.New(), Parents: roaring.New()}
				ms.SaveNode(node4)
				node.SetDependency(ms, node2)
				node2.SetDependency(ms, node3)
				node3.SetDependency(ms, node4)
			},
			ids:          roaring.BitmapOf(1),
			query:        "test query",
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultServeMux = http.NewServeMux()

			mockStorage := NewMockStorage()
			tt.setupMock(mockStorage)

			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				chart, err := graphQuery(mockStorage, tt.ids, tt.query)
				if err != nil {
					http.Error(w, "Error generating graph: "+err.Error(), http.StatusInternalServerError)
					return
				}
				err = chart.Render(w)
				if err != nil {
					http.Error(w, "Error rendering graph: "+err.Error(), http.StatusInternalServerError)
					return
				}
			})

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			// You can add more assertions here to check the response body if needed
		})
	}
}
