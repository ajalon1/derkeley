// Copyright 2026 DataRobot, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tui

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
)

// KeyValue represents a labeled value for human-readable output.
type KeyValue struct {
	Label string
	Value string
	Style lipgloss.Style
}

// PrintKeyValues writes label-value pairs to w with tabwriter alignment.
// Format: "Label:   Value" with consistent spacing.
func PrintKeyValues(w io.Writer, pairs ...KeyValue) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	for _, pair := range pairs {
		styled := pair.Style.Render(pair.Value)
		if _, err := fmt.Fprintf(tw, "%s:\t%s\n", pair.Label, styled); err != nil {
			return err
		}
	}

	return tw.Flush()
}
