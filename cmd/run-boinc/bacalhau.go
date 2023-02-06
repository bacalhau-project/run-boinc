package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/bacalhau/pkg/model"
	"github.com/samber/lo"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const bacalhauLabel = "boinc"

func startJob(ctx context.Context, image string, projectUrl string, weakAccountKey string, timeout time.Duration, domains []string) (string, error) {
	domains = lo.Map[string, string](domains, func(item string, index int) string {
		return fmt.Sprintf("--domain=%s", item)
	})

	args := []string{
		"docker", "run",
		image,
		"--id-only",
		fmt.Sprintf("--labels=%s", bacalhauLabel),
		// TODO "--memory"
		// TODO "--cpu"
		"--timeout", strconv.FormatFloat(timeout.Seconds(), 'f', -1, 64),
		"--wait=false",
		"--network=http",
	}
	args = append(args, domains...)
	args = append(args,
		"--",
		"boinc",
		"--dir", "/outputs",
		"--attach_project", projectUrl, weakAccountKey,
		"--exit_after_finish",
		"--fetch_minimal_work",
	)

	cmd := exec.CommandContext(ctx, "bacalhau", args...)

	id, err := cmd.CombinedOutput()
	if err != nil {
		// TODO bacalhau doesn't output errors to stderr
		//if ee, ok := err.(*exec.ExitError); ok {
		//	return "",
		//}
		return "", fmt.Errorf("%s: %w", string(id), err)
	}

	return strings.TrimSpace(string(id)), nil
}

func waitUntilJobIsRunning(ctx context.Context, jobId string) error {
	t := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-t.C:
			fmt.Printf("Checking if job %s has been accepted\n", jobId)
			running, err := isJobAccepted(ctx, jobId)
			if err != nil {
				return err
			}
			if running {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isJobAccepted(ctx context.Context, jobId string) (bool, error) {
	job, err := getBacalhauJob(ctx, jobId)
	if err != nil {
		return false, err
	}

	for _, state := range job.Status.State.Nodes {
		for _, shard := range state.Shards {
			if shard.State.IsError() {
				return false, fmt.Errorf("job failed")
			}
			if shard.State.HasPassedBidAcceptedStage() {
				return true, nil
			}
		}
	}

	return false, nil
}

func isJobFinished(job model.Job) bool {
	for _, state := range job.Status.State.Nodes {
		for _, shard := range state.Shards {
			if !shard.State.IsTerminal() {
				return false
			}
		}
	}
	return true
}

func jobAlreadyRunning(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "bacalhau", "list", "--output=json")

	list, err := cmd.CombinedOutput()
	if err != nil {
		// TODO bacalhau doesn't output errors to stderr
		//if ee, ok := err.(*exec.ExitError); ok {
		//	return "",
		//}
		return "", fmt.Errorf("%s: %w", string(list), err)
	}

	var jobs []model.Job
	if err := json.Unmarshal(list, &jobs); err != nil {
		return "", err
	}

	for _, job := range jobs {
		for _, annotation := range job.Spec.Annotations {
			if annotation == bacalhauLabel && !isJobFinished(job) {
				return job.Metadata.ID, nil
			}
		}
	}

	return "", nil
}

func getBacalhauJob(ctx context.Context, jobId string) (model.Job, error) {
	cmd := exec.CommandContext(ctx, "bacalhau", "describe", jobId)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return model.Job{}, err
	}

	var job model.Job
	if err := model.YAMLUnmarshalWithMax(output, &job); err != nil {
		return model.Job{}, err
	}

	return job, nil
}
