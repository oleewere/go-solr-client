// Copyright 2018 Oliver Szabo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processor

import (
	"time"
)

// Processor interface that can be implemented to process buffered data (that is hold by the batch context)
type Processor interface {
	// Process handles/processes the data that is hold by the batch context
	Process(*BatchContext) error
	// HandleError handles the error that happens during time based data processing
	HandleError(*BatchContext, error)
}

// BatchContext holds data buffer related information/configuration
type BatchContext struct {
	// BufferData storing data that needs to be processed
	BufferData *[]interface{}
	// MaxBufferSize maximum size of the buffer
	MaxBufferSize int
	// LastChanged time that is updated after the buffer has cleaned
	LastChanged *time.Time
	// TimeBasedProcessing flag that can enable time based processing of the buffered data
	TimeBasedProcessing bool
	// ProcessTimeInterval time that used to wait between time based processing tasks
	ProcessTimeInterval time.Duration
	// RetryTimeInterval sleep between data processing calls that is retried because of an error
	RetryTimeInterval time.Duration
	// MaxRetries maximum number of retries on fail, 0 means the data processing will retry forever
	MaxRetries int
	// ExtraParams holds any extra context data
	ExtraParams map[string]interface{}
}
