package graph

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func RunGraphVisualizer(storage Storage, ids *roaring.Bitmap, query, addr string) (func(), error) {
	srv := &http.Server{Addr: ":" + addr}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		chart, err := graphQuery(storage, ids, query)
		if err != nil {
			http.Error(w, "Error generating graph: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Add debug information
		err = chart.Render(w)
		if err != nil {
			http.Error(w, "Error rendering graph: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("Graph rendered successfully")
	})

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("HTTP server ListenAndServe: %v", err)
		}
	}()

	fmt.Printf("Starting server on http://localhost:%s\n", addr)

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
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
			nodes = append(nodes, opts.GraphNode{SymbolSize: max(10, connections), Name: node.Name, ItemStyle: &opts.ItemStyle{Color: "#42b0f5"}, X: float32(rand.Intn(100000)), Y: float32(rand.Intn(100000))})
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
