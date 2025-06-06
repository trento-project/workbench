package hana

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/trento-project/workbench/internal/support"
	"github.com/trento-project/workbench/pkg/utils"
)

type SystemReplicationState struct {
	Online        bool
	Mode          string
	OperationMode string
}

func newSystemReplicationState(srData map[string]any) *SystemReplicationState {
	// @todo keys in srData might be missing
	return &SystemReplicationState{
		Online:        srData["online"] == "true",
		Mode:          srData["mode"].(string),
		OperationMode: srData["operation_mode"].(string),
	}
}

func (srState *SystemReplicationState) IsOnline() bool {
	// @todo check what is the bit of data that indicates the system is online
	return srState.Online
}

func (srState *SystemReplicationState) IsSecondary() bool {
	// @todo check what is the bit of data that indicates the system is a secondary
	return srState.Mode == "sync"
}

func (srState *SystemReplicationState) IsRegistered() bool {
	// @todo check what is the bit of data that indicates the system is a registered/unregistered secondary
	return srState.Mode != "none"
}

type Hdbnsutil interface {
	SystemReplicationState(ctx context.Context, sid string) (*SystemReplicationState, error)
	UnregisterHANASecondary(ctx context.Context, sid string) error
}

type hdbnsutil struct {
	executor support.CmdExecutor
	logger   *logrus.Entry
}

func NewHdbnsutil(
	executor support.CmdExecutor,
	logger *logrus.Entry,
) Hdbnsutil {
	return &hdbnsutil{
		executor: executor,
		logger:   logger,
	}
}

func (h *hdbnsutil) SystemReplicationState(ctx context.Context, sid string) (*SystemReplicationState, error) {
	srData, err := runHdbnsutilCommand(ctx, h.executor, h.logger, "-sr_state -sapcontrol=1", sid)

	if err != nil {
		h.logger.Errorf("could not execute hdbnsutil command: %v", err)
		return nil, err
	}

	systemReplicationStateMap := utils.FindMatches("(.+)=(.*)", srData)

	return newSystemReplicationState(systemReplicationStateMap), nil
}

func (h *hdbnsutil) UnregisterHANASecondary(ctx context.Context, sid string) error {
	_, err := runHdbnsutilCommand(ctx, h.executor, h.logger, "-sr_unregister", sid)
	if err != nil {
		h.logger.Errorf("could not unregister HANA secondary instance %s: %v", sid, err)
		return err
	}

	h.logger.Infof("HANA secondary instance %s unregistered successfully", sid)
	return nil
}

func runHdbnsutilCommand(
	ctx context.Context,
	executor support.CmdExecutor,
	logger *logrus.Entry,
	command string,
	sid string,
) ([]byte, error) {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmd := fmt.Sprintf("hdbnsutil %s", command)
	output, err := executor.Exec(ctx, "/usr/bin/su", "-lc", cmd, user)
	if err != nil {
		logger.Errorf("could not execute hdbnsutil command '%s': %v", cmd, err)
		return nil, err
	}
	return output, nil
}
