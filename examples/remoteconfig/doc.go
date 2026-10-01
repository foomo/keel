// Remoteconfig is an example service demonstrating remote configuration from etcd
// and watching value changes.
//
// Start etcd on localhost:2379, run `go run ./examples/remoteconfig` and change foo
// in the cluster.yaml key; `curl localhost:8080` prints the current value to stdout.
package main
