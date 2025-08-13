package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/moonstripe/workflow-dag/graph"
)

func test_graphs() {
	if err := os.RemoveAll("example_output"); err != nil {
		log.Fatalf("Could not clear example_output dir: %s", err)
	}

	if err := os.MkdirAll("example_output", 0755); err != nil {
		log.Fatalf("Could not create example_output dir: %s", err)
	}

	n := flag.Int("n", 6, "number of nodes")
	p := flag.Float64("p", 0.3, "probability of each allowed edge")
	seed := flag.Int64("seed", 0, "random seed (0 -> time-based)")
	printFlag := flag.Bool("print", false, "print graph edges")
	useNames := flag.Bool("names", true, "use names T0..Tn-1")
	flag.Parse()

	if *n <= 0 {
		log.Fatal("n must be > 0")
	}
	if *p < 0 || *p > 1 {
		log.Fatal("p must be in [0,1]")
	}

	var rnd *rand.Rand
	if *seed == 0 {
		rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		rnd = rand.New(rand.NewSource(*seed))
	}

	g := graph.GenerateRandomSDAG(*n, *p, rnd)

	if *printFlag {
		fmt.Println("Edges (initial DAG):")
		for _, u := range g.GetAllNodes() {
			for _, v := range g.GetAdjacencyFromNode(u) {
				if *useNames {
					fmt.Printf("  %s -> T%s\n", u.Label(), v)
				} else {
					fmt.Printf("  %s -> %s\n", u.Label(), v)
				}
			}
		}
	}

	// 1) initial
	if err := graph.SaveDAGDotAndPNG(*g, "example_output/dag_initial", *useNames); err != nil {
		log.Fatalf("saving initial DAG: %v", err)
	}

	order, ok := graph.TopologicalSortKahn(g)
	if !ok {
		fmt.Println("ERROR: Cycle detected (shouldn't happen with generator).")
		return
	}
	fmt.Println("\nTopological order:")
	for i, v := range order {
		if i > 0 {
			fmt.Print(" -> ")
		}
		if *useNames {
			fmt.Printf("T%s", v)
		} else {
			fmt.Printf("%s", v)
		}
	}
	fmt.Println()

	layers := graph.KahnLayers(g)
	fmt.Println("\nExecution layers:")
	for i, layer := range layers {
		fmt.Printf("  Layer %d:", i)
		for _, v := range layer {
			if *useNames {
				fmt.Printf(" T%s", v)
			} else {
				fmt.Printf(" %s", v)
			}
		}
		fmt.Println()
	}

	// 2) final (linear)
	gLinear := graph.BuildLinearChain(order)
	if err := graph.SaveDAGDotAndPNG(*gLinear, "example_output/dag_final_linear", *useNames); err != nil {
		log.Fatalf("saving linear DAG: %v", err)
	}

	// 3) final (parallel)
	gParallel := graph.BuildLayeredDag(layers)
	if err := graph.SaveDAGDotAndPNGWithRanks(*gParallel, "example_output/dag_final_parallel", *useNames, layers); err != nil {
		log.Fatalf("saving parallel DAG: %v", err)
	}
}

func main() {
	// Test Graph functionality
	test_graphs()
}
