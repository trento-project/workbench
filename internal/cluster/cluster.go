package cluster

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/trento-project/workbench/internal/support"
)

const resourceRefreshedMessage = "got reply (done)"

var clusterIdlePatternCompiled = regexp.MustCompile("S_IDLE")

type Cluster interface {
	IsHostOnline(ctx context.Context) bool
	IsIdle(ctx context.Context) (bool, error)
	ResourceRefresh(ctx context.Context, resourceID, nodeID string) error
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
	output, err := c.executor.Exec(ctx, "crm", "status")
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

// ResourceRefresh runs the `crm resource refresh [<rsc>] [<node>]` command.
// https://crmsh.github.io/man-5.0/#cmdhelp.resource.refresh
// The node argument requires the resource beforehand.
// If the given node is not found, the command does not return -1, so the
// std output must be compared to see if it returns a correct value.
func (c *Client) ResourceRefresh(ctx context.Context, resourceID, nodeID string) error {
	if nodeID != "" && resourceID == "" {
		return errors.New("nodeID cannot be provided without a resourceID")
	}

	args := []string{"resource", "refresh"}
	if resourceID != "" {
		args = append(args, resourceID)
		if nodeID != "" {
			args = append(args, nodeID)
		}
	}

	c.logger.Info("Refreshing cluster resource", "resourceID", resourceID, "nodeID", nodeID)
	output, err := c.executor.Exec(ctx, "crm", args...)
	if err != nil {
		return fmt.Errorf("failed to refresh resource: %w, output: %s", err, string(output))
	}

	if !strings.Contains(string(output), resourceRefreshedMessage) {
		return fmt.Errorf("failed to refresh resource, unexpected output: %s", string(output))
	}

	c.logger.Info("Cluster resource refreshed successfully")
	return nil
}
