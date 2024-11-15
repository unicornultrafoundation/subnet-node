package subnet

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
)

func Main(repoPath string, configPath string) {
	if err := run(repoPath, configPath); err != nil {
		log.Fatal(err)
	}
}

func run(repoPath string, configPath string) error {

	r, err := snrepo.OpenWithUserConfig(repoPath, configPath)
	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	node, err := core.NewNode(context.Background(), &core.BuildCfg{
		Repo: r,
	})

	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		return err
	}

	defer node.Close()
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
			// Xử lý công việc khác nếu không có tín hiệu
			// Ví dụ: in ra thông báo mỗi giây
			fmt.Println("Running...")
			time.Sleep(1 * time.Second) // Nghỉ 1 giây để tránh lãng phí CPU
		}
	}
}
