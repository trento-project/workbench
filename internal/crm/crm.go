package crm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"

	"github.com/trento-project/workbench/internal/support"
)

type Crm interface {
	IsHostOnline(ctx context.Context) bool
	IsIdle(ctx context.Context) (bool, error)
	StartCluster(ctx context.Context) error
	StopCluster(ctx context.Context) error
}

type CrmClient struct {
	executor support.CmdExecutor
	logger   *slog.Logger
}

func NewDefaultCrmClient() Crm {
	return NewCrmClient(
		support.CliExecutor{},
		slog.Default(),
	)
}

func NewCrmClient(executor support.CmdExecutor, logger *slog.Logger) Crm {
	return &CrmClient{
		executor: executor,
		logger:   logger,
	}
}

func (c *CrmClient) IsHostOnline(ctx context.Context) bool {
	output, err := c.executor.Exec(ctx, "crm", "status", "simple")
	if err != nil {
		return false
	}

	c.logger.Debug("CRM status output", "output", string(output))

	return true
}

func (c *CrmClient) StartCluster(ctx context.Context) error {
	c.logger.Info("Starting CRM cluster")
	output, err := c.executor.Exec(ctx, "crm", "cluster", "start")
	if err != nil {
		return fmt.Errorf("failed to start CRM cluster: %w, output: %s", err, string(output))
	}

	c.logger.Info("CRM cluster started successfully")
	return nil
}

func (c *CrmClient) StopCluster(ctx context.Context) error {
	c.logger.Info("Stopping CRM cluster")
	output, err := c.executor.Exec(ctx, "crm", "cluster", "stop")
	if err != nil {
		return fmt.Errorf("failed to stop CRM cluster: %w, output: %s", err, string(output))
	}

	c.logger.Info("CRM cluster stopped successfully")
	return nil
}

func (c *CrmClient) IsIdle(ctx context.Context) (bool, error) {
	idleOutput, err := c.executor.Exec(ctx, "cs_clusterstate", "-i")
	if err != nil {
		return false, fmt.Errorf("error running cs_clusterstate: %w", err)
	}

	const clusterIdlePattern = `S_IDLE`
	clusterIdlePatternCompiled := regexp.MustCompile(clusterIdlePattern)

	if !clusterIdlePatternCompiled.Match(idleOutput) {
		return false, fmt.Errorf("cluster is not in S_IDLE state")
	}

	return true, nil
}

func md5sumFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New() //nolint:gosec
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
