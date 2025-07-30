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
	// GetClusterId returns the unique identifier for the cluster.
	GetClusterId() (string, error)
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

// By default, Trento uses the cluster ID as the MD5 hash of the authkey file
// located at /etc/corosync/authkey. This is used to uniquely identify the cluster
// in the CRM operations. If the file does not exist or cannot be read, an error
// will be returned, and the operation will not proceed.
func (c *CrmClient) GetClusterId() (string, error) {
	id, err := md5sumFile("/etc/corosync/authkey")
	if err != nil {
		return "", fmt.Errorf("failed to read authkey file: %w", err)
	}
	return id, nil
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
