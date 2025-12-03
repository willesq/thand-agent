// Package models provides public SDK types for the thand agent.
package models

import internal "github.com/thand-io/agent/internal/models"

// LocalSession represents a session stored locally on the client.
// It contains session data that is persisted on the local machine.
type LocalSession = internal.LocalSession

// Session represents an active authentication session with the agent.
// It contains the session state, credentials, and metadata.
type Session = internal.Session

// ExportableSession is a session format suitable for export and sharing.
// It contains only the data that can be safely exported or transferred.
type ExportableSession = internal.ExportableSession
