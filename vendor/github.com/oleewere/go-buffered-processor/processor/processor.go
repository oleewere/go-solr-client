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

// ProcessData processes the buffer data and clean the buffer - processing itself should be implemented
func ProcessData(data interface{}, batchContext *BatchContext, processor Processor) error {
	dataArray := (*batchContext).BufferData
	*dataArray = append(*dataArray, data)
	if batchContext.MaxBufferSize <= len(*dataArray) {
		err := retryFunction(batchContext.MaxRetries, batchContext.RetryTimeInterval, func() error {
			return processor.Process(batchContext)
		})
		if err != nil {
			return err
		}
		actualTime := time.Now()
		*batchContext.LastChanged = actualTime
		*batchContext.BufferData = make([]interface{}, 0)
	}
	return nil
}

// StartTimeBasedProcessing starts a scheduled tasks (with time interval) to call Process on the Processor interface.
// use this as a go routine, that can be useful if the buffered data has not updated recently, but it is needed to process the data anyway after some time.
func StartTimeBasedProcessing(batchContext *BatchContext, processor Processor, waitIntervalSec time.Duration) {
	if batchContext.TimeBasedProcessing {
		for {
			lastChangeTime := *batchContext.LastChanged
			processTimeInterval := batchContext.ProcessTimeInterval
			diff := time.Now().Sub(lastChangeTime)
			if diff > processTimeInterval {
				err := retryFunction(batchContext.MaxRetries, batchContext.RetryTimeInterval, func() error {
					return processor.Process(batchContext)
				})
				if err != nil {
					processor.HandleError(batchContext, err)
				} else {
					actualTime := time.Now()
					*batchContext.LastChanged = actualTime
					*batchContext.BufferData = make([]interface{}, 0)
				}
			}
			time.Sleep(waitIntervalSec * time.Second)
		}
	}
}

// CreateDefaultBatchContext creates a default batch context (buffer size: 1000, lastChanged: actualTime, process time interval: 30 sec, retry time interval: 5)
func CreateDefaultBatchContext() *BatchContext {
	emptyData := make([]interface{}, 0)
	emptyExtraParams := make(map[string]interface{})
	actualTime := time.Now()
	return &BatchContext{
		BufferData: &emptyData, MaxBufferSize: 1000, LastChanged: &actualTime, TimeBasedProcessing: true,
		ProcessTimeInterval: 30 * time.Second, RetryTimeInterval: 5 * time.Second, ExtraParams: emptyExtraParams}
}

func retryFunction(maxAttempts int, sleep time.Duration, fn func() error) error {
	return retry(0, maxAttempts, sleep, fn)
}

func retry(attempts int, maxAttempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if maxAttempts == 0 {
			time.Sleep(sleep)
			return retry(attempts, maxAttempts, sleep, fn)
		} else if attempts++; attempts >= maxAttempts {
			time.Sleep(sleep)
			return retry(attempts, maxAttempts, sleep, fn)
		}
		return err
	}
	return nil
}
