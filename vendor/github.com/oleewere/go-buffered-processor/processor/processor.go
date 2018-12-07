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

func CreateDefaultBatchContext() *BatchContext {
	emptyData := make([]interface{}, 0)
	actualTime := time.Now()
	return &BatchContext{
		BufferData: &emptyData, MaxBufferSize: 1000, LastChanged: &actualTime, TimeBasedProcessing: true,
		ProcessTimeInterval: 30 * time.Second, RetryTimeInterval: 5 * time.Second}
}

func ProcessData(data interface{}, processor Processor) error {
	batchContext := processor.GetBatchContext()
	dataArray := (*batchContext).BufferData
	*dataArray = append(*dataArray, data)
	if batchContext.MaxBufferSize <= len(*dataArray) {
		err := retryFunction(batchContext.MaxRetries, batchContext.RetryTimeInterval, func() error {
			return processor.Process()
		})
		if err != nil {
			return err
		} else {
			actualTime := time.Now()
			*batchContext.LastChanged = actualTime
			*batchContext.BufferData = make([]interface{}, 0)
		}
	}
	return nil
}

func StartTimeBasedProcessing(processor Processor, waitIntervalSec time.Duration) {
	batchContext := processor.GetBatchContext()
	if batchContext.TimeBasedProcessing {
		for {
			lastChangeTime := *batchContext.LastChanged
			processTimeInterval := batchContext.ProcessTimeInterval
			diff := time.Now().Sub(lastChangeTime)
			if diff > processTimeInterval {
				err := retryFunction(batchContext.MaxRetries, batchContext.RetryTimeInterval, func() error {
					return processor.Process()
				})
				if err != nil {
					processor.HandleError(err)
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
