# task graph

This is primarily a sandbox to play with workflow DAGs.

- `graph/graph.go` handles common graph operations, like sorting a DAG topologically and determining a DAG's flow for parallel execution. It exposes some interfaces for use elsewhere.
- `task/task.go` is a concrete instantiation of a task graph. WIP

## example output:

`example_output` shows the dot (graphviz) representation along with a png image.

| dot                                                              | png                                                                                      |
| ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| [dag_initial.dot](example_output/dag_initial.dot?raw=true)       | ![initial DAG](example_output/dag_initial.png?raw=true "initial DAG")                    |
| [dag_final_parallel.dot]{dag_final_parallel.dot?raw=true)        | ![transformed DAG](example_output/dag_final_parallel.png?raw=true "transformed DAG")     |
| [dag_final_linear.dot](example_output/dag_final_linear?raw=true) | ![linear ordered DAG](example_output/dag_final_linear.png?raw=true "linear ordered DAG") |

## project goals

- [x] extensible DAG interface with useful operations in workflows
- [ ] toy example with TaskGraph
- [ ] integrate K8s or Docker as sandboxes to complete Actions.
