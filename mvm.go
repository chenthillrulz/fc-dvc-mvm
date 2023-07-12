package main

import (
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
)

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

type runningFirecracker struct {
	vmmCtx    context.Context
	vmmCancel context.CancelFunc
	vmmID     string
	machine   *firecracker.Machine
	ip        net.IP
}

func waitForVMToBoot(ctx context.Context, ip net.IP) error {
	// Wait until we receive a ping response
	for {
		select {
		case <-ctx.Done():
			// Timeout
			return ctx.Err()
		default:
			p := fastping.NewPinger()
			p.AddIP(ip.String())
			p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
				fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
			}
			p.OnIdle = func() {
				fmt.Printf("finish")
			}

			err := p.Run()
			if err != nil {
				log.WithError(err).Info("VM not ready yet")
				time.Sleep(time.Second)
				continue
			}
			log.WithField("ip", ip).Info("VM agent ready")
			return nil
		}
	}
}

func createVM(ctx context.Context) (*runningFirecracker, error) {
	vm, err := createAndStartVM(ctx)
	if err != nil {
		log.Errorf("Unable to start micro vm")
		return nil, err
	}

	// understand what it is
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	log.WithField("ip", vm.ip).Info("New VM created and started")

	err = waitForVMToBoot(ctx, vm.ip)
	if err != nil {
		log.WithError(err).Info("VM not ready yet")
		vm.vmmCancel()
		return nil, err
	}

	return vm, nil
}

// Create a VMM with a given set of options and start the VM
func createAndStartVM(ctx context.Context) (*runningFirecracker, error) {
	vmmID := xid.New().String()

	fcCfg, err := getFirecrackerConfig(vmmID)
	if err != nil {
		log.Errorf("Error: %s", err)
		return nil, err
	}
	logger := log.New()

	if false { // TODO
		log.SetLevel(log.DebugLevel)
		logger.SetLevel(log.DebugLevel)
	}

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.NewEntry(logger)),
	}

	firecrackerBinary, err := exec.LookPath("firecracker")
	if err != nil {
		log.Error("Unable to find firecracker")
		return nil, err
	}

	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("binary %q does not exist: %v", firecrackerBinary, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		return nil, fmt.Errorf("binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&0111 == 0 {
		return nil, fmt.Errorf("binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
	}

	// if the jailer is used, the final command will be built in NewMachine()
	if fcCfg.JailerCfg == nil {
		cmd := firecracker.VMCommandBuilder{}.
			WithBin(firecrackerBinary).
			WithSocketPath(fcCfg.SocketPath).
			// WithStdin(os.Stdin).
			// WithStdout(os.Stdout).
			WithStderr(os.Stderr).
			Build(ctx)

		machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))
	}

	vmmCtx, vmmCancel := context.WithCancel(ctx)

	m, err := firecracker.NewMachine(vmmCtx, fcCfg, machineOpts...)
	if err != nil {
		log.Error("Failed to create machine")
		vmmCancel()
		return nil, fmt.Errorf("failed creating machine: %s", err)
	}

	log.Info("created machine")

	if err := m.Start(vmmCtx); err != nil {
		log.Error("failed to start machine - ", err)
		vmmCancel()
		return nil, fmt.Errorf("failed to start machine: %v", err)
	}

	log.WithField("ip", m.Cfg.NetworkInterfaces[0].StaticConfiguration).Info("machine started")

	return &runningFirecracker{
		vmmCtx:    vmmCtx,
		vmmCancel: vmmCancel,
		vmmID:     vmmID,
		machine:   m,
		ip:        m.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP,
	}, nil
}
