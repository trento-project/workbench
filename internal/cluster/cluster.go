package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/trento-project/workbench/internal/support"
)

var clusterIdlePatternCompiled = regexp.MustCompile("S_IDLE")

type Cluster interface {
	IsHostOnline(ctx context.Context) bool
	IsIdle(ctx context.Context) (bool, error)
	StartCluster(ctx context.Context) error
	StopCluster(ctx context.Context) error
}

type Client struct {
	executor support.CmdExecutor
	logger   *slog.Logger
}

func NewDefaultClusterClient() Cluster {
	return NewClusterClient(
		support.CliExecutor{},
		slog.Default(),
	)
}

func NewClusterClient(executor support.CmdExecutor, logger *slog.Logger) Cluster {
	return &Client{
		executor: executor,
		logger:   logger,
	}
}

func (c *Client) IsHostOnline(ctx context.Context) bool {
	output, err := c.executor.Exec(ctx, "crm", "status", "simple")
	if err != nil {
		return false
	}

	c.logger.Debug("CRM status output", "output", string(output))

	return true
}

func (c *Client) StartCluster(ctx context.Context) error {
	c.logger.Info("Starting CRM cluster")
	output, err := c.executor.Exec(ctx, "crm", "cluster", "start")
	if err != nil {
		return fmt.Errorf("failed to start CRM cluster: %w, output: %s", err, string(output))
	}

	c.logger.Info("CRM cluster started successfully")
	return nil
}

func (c *Client) StopCluster(ctx context.Context) error {
	c.logger.Info("Stopping CRM cluster")
	output, err := c.executor.Exec(ctx, "crm", "cluster", "stop")
	if err != nil {
		return fmt.Errorf("failed to stop CRM cluster: %w, output: %s", err, string(output))
	}

	c.logger.Info("CRM cluster stopped successfully")
	return nil
}

func (c *Client) IsIdle(ctx context.Context) (bool, error) {
	idleOutput, err := c.executor.Exec(ctx, "cs_clusterstate", "-i")
	if err != nil {
		return false, fmt.Errorf("error running cs_clusterstate: %w", err)
	}

	if !clusterIdlePatternCompiled.Match(idleOutput) {
		return false, nil
	}

	return true, nil
}
