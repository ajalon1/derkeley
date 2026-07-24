//go:build windows

package install

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const policyWarning = `  ⚠  PowerShell execution policy is set to %s.
     The completion profile was written but will not load on the next
     PowerShell launch.

     Run this command to fix it:
       Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
`

func warnIfExecutionPolicyRestricted(shellExe string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, shellExe, "-Command", "Get-ExecutionPolicy").Output()
	if err != nil {
		return
	}

	policy := strings.TrimSpace(string(out))

	if policy == "Restricted" {
		fmt.Fprintf(os.Stderr, policyWarning, policy)
	}
}
