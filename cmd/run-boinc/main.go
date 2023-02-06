package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/filecoin-project/bacalhau/pkg/logger"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	var image string
	var projectUrl string
	var weakAccountKey string
	var domains stringSlice
	var timeout time.Duration

	flag.StringVar(&image, "image", "boinc/client:base-ubuntu", "Docker image to use")
	flag.StringVar(&projectUrl, "project-url", "", "URL of the BOINC project")
	flag.StringVar(&weakAccountKey, "weak-account-key", "", "*Weak* account key to connect to the project")
	flag.Var(&domains, "domain", "List of domains to allow network traffic to")
	flag.DurationVar(&timeout, "timeout", time.Hour*24*7, "How long jobs should run for before they get stopped by Bacalhau")

	flag.Parse()

	if projectUrl == "" {
		panic("missing project-url flag")
	}
	if weakAccountKey == "" {
		panic("missing weak-account-key flag")
	}
	if len(domains.value) == 0 {
		panic("missing domain flag")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			panic(ctx.Err())
		case <-t.C:
			if err := run(ctx, image, projectUrl, weakAccountKey, timeout, domains.value); err != nil {
				panic(err)
			}
		}
	}
}

func run(ctx context.Context, image string, projectUrl string, weakAccountKey string, timeout time.Duration, domains []string) error {
	if alreadyRunning, err := jobAlreadyRunning(ctx); err != nil {
		return err
	} else if alreadyRunning != "" {
		fmt.Printf("Job %s is already running\n", alreadyRunning)
		return nil
	}

	job, err := startJob(ctx, image, projectUrl, weakAccountKey, timeout, domains)
	if err != nil {
		return err
	}

	fmt.Printf("Work processing under Bacalhau job %s\n", job)

	if err := waitUntilJobIsRunning(ctx, job); err != nil {
		return err
	}

	return nil
}

type stringSlice struct {
	value []string
}

func (s *stringSlice) String() string {
	return strings.Join(s.value, ",")
}

func (s *stringSlice) Set(v string) error {
	s.value = append(s.value, v)
	return nil
}
