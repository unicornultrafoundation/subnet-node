package subnet

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/coreapi"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
)

var log = logrus.New().WithField("service", "subnet")

func Main(repoPath string, configPath string) {
	if err := run(repoPath, configPath); err != nil {
		log.Fatal(err)
	}
}

func run(repoPath string, configPath string) error {
	// let the user know we're going.
	fmt.Printf("Initializing subnetnode...\n")

	r, err := snrepo.OpenWithUserConfig(repoPath, configPath)
	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	node, err := core.NewNode(context.Background(), &core.BuildCfg{
		Repo:   r,
		Online: true,
	})

	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	defer node.Close()

	time.AfterFunc(1*time.Minute, func() {
		api, err := coreapi.NewCoreAPI(node)
		if err != nil {
			log.Errorf("failed to access CoreAPI: %v", err)
			return
		}
		peers, err := api.Swarm().Peers(context.Background())
		if err != nil {
			log.Errorf("failed to read swarm peers: %v", err)
			return
		}
		if len(peers) == 0 {
			log.Error("failed to bootstrap (no peers found): consider updating Bootstrap or Peering section of your config")
		}
	})

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)

	// Giả sử bạn có một công việc cần tiếp tục chạy
	for {
		select {
		case <-interrupts:
			// Nhận tín hiệu ngắt
			fmt.Println("Received interrupt signal, exiting...")
			return nil // Dừng chương trình khi nhận tín hiệu
		default:
			time.Sleep(1 * time.Second) // Nghỉ 1 giây để tránh lãng phí CPU
		}
	}
}
