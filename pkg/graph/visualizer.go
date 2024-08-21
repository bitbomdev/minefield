package graph

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func RunGraphVisualizer(storage Storage, ids *roaring.Bitmap, query string, server *http.Server) (func(), error) {
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chart, err := graphQuery(storage, ids, query)
		if err != nil {
			http.Error(w, "Error generating graph: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = chart.Render(w)
		if err != nil {
			http.Error(w, "Error rendering graph: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("Graph rendered successfully")
	})

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("HTTP server ListenAndServe: %v", err)
		}
	}()

	fmt.Printf("Starting server on %s\n", server.Addr)

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Server forced to shutdown: %v\n", err)
		}
		fmt.Println("Server stopped")
	}

	return shutdown, nil
}

func graphQuery(storage Storage, ids *roaring.Bitmap, query string) (*charts.Graph, error) {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: query,
		}),
		charts.WithAnimation(false),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "9000px",
			Height: "5000px",
			Theme:  "dark",
		}),
	)

	var nodes []opts.GraphNode
	var links []opts.GraphLink

	alreadyCreatedNodes := roaring.New()
	alreadyCreatedLinks := make(map[uint32]*roaring.Bitmap)

	for _, id := range ids.ToArray() {
		node, err := storage.GetNode(id)
		if err != nil {
			return nil, err
		}
		connections := 0

		for _, dep := range append(node.Children.ToArray(), node.Parents.ToArray()...) {
			if ids.Contains(dep) {
				connections++
			}
		}
		if !alreadyCreatedNodes.Contains(id) {
			alreadyCreatedNodes.Add(id)
			symbolSize := calculateSymbolSize(connections)
			color := getColorForSize(symbolSize)
			nodes = append(nodes, opts.GraphNode{
				SymbolSize: symbolSize,
				Name:       node.Name,
				ItemStyle:  &opts.ItemStyle{Color: color},
				X:          float32(rand.Intn(100000)),
				Y:          float32(rand.Intn(100000)),
			})
		}

	}
	for _, id := range ids.ToArray() {
		node, err := storage.GetNode(id)
		if err != nil {
			return nil, err
		}

		for _, dep := range append(node.Children.ToArray(), node.Parents.ToArray()...) {

			depNode, err := storage.GetNode(dep)
			if err != nil {
				return nil, err
			}

			if alreadyCreatedNodes.Contains(dep) {
				if bitmap := alreadyCreatedLinks[id]; bitmap == nil || !bitmap.Contains(dep) {
					links = append(links, opts.GraphLink{Source: node.Name, Target: depNode.Name})
					if bitmap == nil {
						bitmap = roaring.New()
					}
					bitmap.Add(dep)
					alreadyCreatedLinks[id] = bitmap

					if oppBitmap := alreadyCreatedLinks[dep]; oppBitmap == nil {
						oppBitmap = roaring.New()
						alreadyCreatedLinks[dep] = oppBitmap
					}
					alreadyCreatedLinks[dep].Add(id)
				}
			}
		}
	}

	graph.AddSeries("graph", nodes, links).
		SetSeriesOptions(
			charts.WithGraphChartOpts(opts.GraphChart{
				Roam:               opts.Bool(true),
				FocusNodeAdjacency: opts.Bool(true),
				Force:              &opts.GraphForce{Repulsion: 80000000, InitLayout: "circular", EdgeLength: 20},
			}),
			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:     opts.Bool(true),
					Color:    "white",
					Position: "left",
				},
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Curveness: 0.1,
			}),
		)
	fmt.Println("Graph generated successfully")
	fmt.Printf("Number of nodes: %d\n", len(nodes))
	fmt.Printf("Number of links: %d\n", len(links))
	return graph, nil
}

func getColorForSize(size int) string {
	// Map size to a value between 0 and 1
	t := math.Max(0, math.Min(1, float64(size-10)/50)) // Clamp t between 0 and 1

	// Define color stops (muted versions)
	colors := []struct{ r, g, b uint8 }{
		{139, 0, 0},     // Dark red (smallest nodes)
		{165, 42, 42},   // Brown
		{178, 34, 34},   // Firebrick
		{205, 92, 92},   // Indian red
		{210, 105, 30},  // Chocolate
		{205, 133, 63},  // Peru
		{210, 105, 30},  // Muted orange (middle nodes)
		{188, 143, 143}, // Rosy brown
		{199, 21, 133},  // Medium violet red
		{186, 85, 211},  // Medium orchid (replacing Pale violet red)
		{255, 20, 147},  // Deep pink (largest nodes)
	}

	// Find the two colors to interpolate between
	i := int(t * float64(len(colors)-1))
	i = int(math.Min(float64(len(colors)-2), float64(i))) // Ensure i is within bounds

	c1, c2 := colors[i], colors[i+1]

	// Interpolate between the two colors
	f := t*float64(len(colors)-1) - float64(i)
	r := uint8(float64(c1.r)*(1-f) + float64(c2.r)*f)
	g := uint8(float64(c1.g)*(1-f) + float64(c2.g)*f)
	b := uint8(float64(c1.b)*(1-f) + float64(c2.b)*f)

	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

func calculateSymbolSize(connections int) int {
	if connections == 0 {
		return 5 // Minimum size for nodes with no connections
	}
	// Use logarithmic scale to compress the range
	logSize := math.Log1p(float64(connections)) // log(x+1) to handle 0 connections
	// Map the log value to a range between 8 and 80
	minSize := 8.0
	maxSize := 80.0
	maxLogConnections := math.Log1p(1000) // Adjust this based on your max expected connections
	scaledSize := minSize + math.Pow(logSize/maxLogConnections, 1.5)*(maxSize-minSize)
	return int(math.Round(scaledSize))
	// return max(20, connections)
}
