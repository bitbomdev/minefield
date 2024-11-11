package v1

import (
	"context"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/RoaringBitmap/roaring"
	service "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/bitbomdev/minefield/pkg/tools/ingest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func setupService() *Service {
	storage := graph.NewMockStorage()
	return NewService(storage, 1)
}

func TestGetNode(t *testing.T) {
	s := setupService()
	node, err := graph.AddNode(s.storage, "type1", "metadata1", "name1")
	require.NoError(t, err)
	req := connect.NewRequest(&service.GetNodeRequest{Id: node.ID})
	resp, err := s.GetNode(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp.Msg.Node)
	assert.Equal(t, node.ID, resp.Msg.Node.Id)
}

func TestGetNodeByName(t *testing.T) {
	s := setupService()
	node, err := graph.AddNode(s.storage, "type1", "metadata1", "name1")
	require.NoError(t, err)
	req := connect.NewRequest(&service.GetNodeByNameRequest{Name: node.Name})
	resp, err := s.GetNodeByName(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp.Msg.Node)
	assert.Equal(t, node.Name, resp.Msg.Node.Name)
}

func TestGetNodesByGlob(t *testing.T) {
	s := setupService()
	// Add test nodes
	_, err := graph.AddNode(s.storage, "type1", "metadata1", "test_node1")
	require.NoError(t, err)
	_, err = graph.AddNode(s.storage, "type1", "metadata1", "test_node2")
	require.NoError(t, err)
	_, err = graph.AddNode(s.storage, "type1", "metadata1", "other_node")
	require.NoError(t, err)

	// Test GetNodesByGlob with pattern "test_*"
	req := connect.NewRequest(&service.GetNodesByGlobRequest{Pattern: "test_*"})
	resp, err := s.GetNodesByGlob(context.Background(), req)
	require.NoError(t, err)
	assert.Len(t, resp.Msg.Nodes, 2)

	// Verify the nodes returned match the expected nodes
	nodeNames := []string{resp.Msg.Nodes[0].Name, resp.Msg.Nodes[1].Name}
	assert.Contains(t, nodeNames, "test_node1")
	assert.Contains(t, nodeNames, "test_node2")

	// Test with a pattern that matches no nodes
	req = connect.NewRequest(&service.GetNodesByGlobRequest{Pattern: "nonexistent_*"})
	resp, err = s.GetNodesByGlob(context.Background(), req)
	require.NoError(t, err)
	assert.Len(t, resp.Msg.Nodes, 0)
}

func TestQueriesAndCache(t *testing.T) {
	s := setupService()
	_, err := ingest.SBOM("../../testdata/osv-sboms/google_agi.sbom.json", s.storage, nil)
	require.NoError(t, err)

	content, err := os.ReadFile("../../testdata/osv-vulns/GHSA-cx63-2mw6-8hw5.json")
	require.NoError(t, err)
	err = ingest.LoadVulnerabilities(s.storage, content)
	require.NoError(t, err)
	err = ingest.Vulnerabilities(s.storage, nil)
	require.NoError(t, err)
	// Check if the node with name "pkg:pypi/astroid@2.11.7" exists
	req := connect.NewRequest(&service.GetNodeByNameRequest{Name: "pkg:pypi/astroid@2.11.7"})
	resp, err := s.GetNodeByName(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp.Msg.Node)

	// Get the dependencies of the node
	deps := roaring.New()
	_, err = deps.FromBase64(resp.Msg.Node.Dependencies)
	require.NoError(t, err)

	cacheReq := connect.NewRequest(&emptypb.Empty{})
	_, err = s.Cache(context.Background(), cacheReq)
	require.NoError(t, err)

	customLeaderboardReq := connect.NewRequest(&service.CustomLeaderboardRequest{Script: "dependencies vuln"})
	customLeaderboardResp, err := s.CustomLeaderboard(context.Background(), customLeaderboardReq)
	require.NoError(t, err)

	if len(customLeaderboardResp.Msg.Queries) > 0 {
		if len(customLeaderboardResp.Msg.Queries[0].Output) == 0 {
			t.Fatalf("Leaderboard top should have a vuln but got %v with %v vulns", customLeaderboardResp.Msg.Queries[0].Node, len(customLeaderboardResp.Msg.Queries[0].Output))
		}
	} else {
		t.Fatalf("No queries found")
	}

	queryReq := connect.NewRequest(&service.QueryRequest{Script: "dependencies vuln pkg:github.com/google/agi@"})
	queryResp, err := s.Query(context.Background(), queryReq)
	require.NoError(t, err)
	assert.Len(t, queryResp.Msg.Nodes, 1)
	assert.Equal(t, queryResp.Msg.Nodes[0].Name, "GHSA-cx63-2mw6-8hw5")

	allKeysReq := connect.NewRequest(&emptypb.Empty{})
	allKeysResp, err := s.AllKeys(context.Background(), allKeysReq)
	require.NoError(t, err)
	assert.Len(t, allKeysResp.Msg.Nodes, 24)

	clearReq := connect.NewRequest(&emptypb.Empty{})
	_, err = s.Clear(context.Background(), clearReq)
	require.NoError(t, err)
}

func TestQuery(t *testing.T) {
	s := setupService()

	// Add test nodes
	_, err := graph.AddNode(s.storage, "type1", "metadata1", "node1")
	require.NoError(t, err)
	_, err = graph.AddNode(s.storage, "type2", "metadata2", "node2")
	require.NoError(t, err)

	// Test query with no results
	req := connect.NewRequest(&service.QueryRequest{Script: "nonexistent"})
	_, err = s.Query(context.Background(), req)
	require.Error(t, err)

	// Test with empty script
	req = connect.NewRequest(&service.QueryRequest{Script: ""})
	_, err = s.Query(context.Background(), req)
	assert.Error(t, err)

	// Test with nil request
	_, err = s.Query(context.Background(), nil)
	assert.Error(t, err)
}
func TestHealthCheck(t *testing.T) {
	s := setupService()
	req := connect.NewRequest(&emptypb.Empty{})
	resp, err := s.Check(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Msg.Status)
}
