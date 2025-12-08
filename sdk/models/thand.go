package models

import internal "github.com/thand-io/agent/internal/models"

// ThandConfig represents the configuration settings for the Thand agent.
type ThandConfig = internal.ThandConfig

// SynchronizeStartResponse is the response structure returned when initiating a synchronization process.
type SynchronizeStartResponse = internal.SynchronizeStartResponse

// SynchronizeChunkRequest represents a request to upload a chunk of data during synchronization.
type SynchronizeChunkRequest = internal.SynchronizeChunkRequest

// SynchronizeCommitRequest represents the request to finalize and commit a synchronization operation.
type SynchronizeCommitRequest = internal.SynchronizeCommitRequest

// SynchronizeStartRequest represents the initial request to start a synchronization session.
type SynchronizeStartRequest = internal.SynchronizeStartRequest
