// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package support

import (
	"context"
	"os/exec"
)

type CmdExecutor interface {
	Exec(ctx context.Context, name string, arg ...string) ([]byte, error)
}

type CliExecutor struct{}

func (e CliExecutor) Exec(ctx context.Context, name string, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.CombinedOutput()
}
