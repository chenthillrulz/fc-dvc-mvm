package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func getSocketPath(vmmID string) string {
	filename := strings.Join([]string{
		"firecracker.sock",
		strconv.Itoa(os.Getpid()),
		vmmID,
	},
		"-",
	)
	dir := os.TempDir()

	return filepath.Join(dir, filename)
}

func getFirecrackerConfig(vmmID string) (firecracker.Config, error) {
	socket := getSocketPath(vmmID)
	smt := false
	bootargs := "console=ttyS0 reboot=k panic=1 pci=off"
	kernelImagePath := "/home/parallels/fcdemo_resources/vmlinux"
	rootFsPath := "/home/parallels/fcdemo_resources/rootfs.ext4"
	connectorFsPath := "/home/parallels/fcdemo_resources/connectorfs.ext4"
	networkName := "alpine"
	ifName := "veth0"

	return firecracker.Config{
		SocketPath:      socket,
		KernelImagePath: kernelImagePath,
		KernelArgs:      bootargs,
		LogPath:         fmt.Sprintf("%s.log", socket),
		Drives: []models.Drive{{
			DriveID:      firecracker.String("1"),
			PathOnHost:   firecracker.String(rootFsPath),
			IsRootDevice: firecracker.Bool(true),
			IsReadOnly:   firecracker.Bool(false),
			RateLimiter: firecracker.NewRateLimiter(
				// bytes/s
				models.TokenBucket{
					OneTimeBurst: firecracker.Int64(1024 * 1024), // 1 MiB/s
					RefillTime:   firecracker.Int64(500),         // 0.5s
					Size:         firecracker.Int64(1024 * 1024),
				},
				// ops/s
				models.TokenBucket{
					OneTimeBurst: firecracker.Int64(100),  // 100 iops
					RefillTime:   firecracker.Int64(1000), // 1s
					Size:         firecracker.Int64(100),
				}),
		},
			{
				DriveID:      firecracker.String("2"),
				PathOnHost:   firecracker.String(connectorFsPath),
				IsRootDevice: firecracker.Bool(false),
				IsReadOnly:   firecracker.Bool(false),
				RateLimiter: firecracker.NewRateLimiter(
					// bytes/s
					models.TokenBucket{
						OneTimeBurst: firecracker.Int64(1024 * 1024), // 1 MiB/s
						RefillTime:   firecracker.Int64(500),         // 0.5s
						Size:         firecracker.Int64(1024 * 1024),
					},
					// ops/s
					models.TokenBucket{
						OneTimeBurst: firecracker.Int64(100),  // 100 iops
						RefillTime:   firecracker.Int64(1000), // 1s
						Size:         firecracker.Int64(100),
					}),
			},
		},

		NetworkInterfaces: []firecracker.NetworkInterface{{
			// Use CNI to get dynamic IP
			CNIConfiguration: &firecracker.CNIConfiguration{
				NetworkName: networkName,
				IfName:      ifName,
			},
		}},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(1),
			MemSizeMib: firecracker.Int64(256),
			Smt:        &smt,
		},
	}, nil
}
