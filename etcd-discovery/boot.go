package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/deis/deis/pkg/etcd/discovery"
)

func main() {

	cmd := exec.Command("etcd")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	go func() {
		err := cmd.Start()
		if err != nil {
			log.Printf("Failed to start etcd: %s", err)
			os.Exit(2)
		}
	}()

	// Give etcd time to start up.
	log.Print("Sleeping for 5 seconds...")
	time.Sleep(5 * time.Second)
	log.Print("I'm awake.")

	uuid, err := discovery.Token()
	if err != nil {
		log.Printf("Failed to read %s", discovery.TokenFile)
		os.Exit(404)
	}
	size := os.Getenv("DEIS_ETCD_CLUSTER_SIZE")
	if size == "" {
		size = "3"
	}

	key := fmt.Sprintf(discovery.ClusterSizeKey, uuid)
	cli := etcd.NewClient([]string{"http://localhost:2379"})
	if _, err := cli.Create(key, size, 0); err != nil {
		log.Printf("Failed to add key: %s", err)
	}

	log.Printf("The etcd-discovery service is now ready and waiting.")
	if err := cmd.Wait(); err != nil {
		log.Printf("Etcd stopped running: %s", err)
	}
}
