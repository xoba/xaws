package xaws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

// if an activity function returns this, that means it wants to handle the task token response itself.
var ErrDeferTaskTokenResponse = errors.New("task token response deferred")

// TaskContext contains the Step Functions task token.
type TaskContext struct {
	TaskToken string
}

// RunActivity polls for a task and executes activityFunc when one arrives.
func RunActivity[IN, OUT any](
	c context.Context,
	svc *sfn.Client,
	workerName string,
	arn string,
	heartbeatInterval time.Duration,
	activityFunc func(IN, TaskContext) (OUT, error),
) error {
	for {
		select {
		case <-c.Done():
			return c.Err()
		default:
		}
		if err := func() error {
			r, err := svc.GetActivityTask(c, &sfn.GetActivityTaskInput{
				ActivityArn: aws.String(arn),
				WorkerName:  aws.String(workerName),
			})
			if err != nil {
				return err
			}
			if r.TaskToken == nil {
				return nil
			}
			if r.Input == nil {
				return fmt.Errorf("nil input")
			}
			var input IN
			if err := json.Unmarshal([]byte(*r.Input), &input); err != nil {
				return err
			}
			ctx, cancel := context.WithCancel(c)
			defer cancel()
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(heartbeatInterval / 2):
						if _, err := svc.SendTaskHeartbeat(c, &sfn.SendTaskHeartbeatInput{
							TaskToken: r.TaskToken,
						}); err != nil {
							log.Printf("can't send task heartbeat: %v", err)
						}
					}
				}
			}()

			output, err := activityFunc(input, TaskContext{TaskToken: *r.TaskToken})
			switch {
			case err == nil:
				buf, err := json.Marshal(output)
				if err != nil {
					return err
				}
				if _, err := svc.SendTaskSuccess(c, &sfn.SendTaskSuccessInput{
					Output:    aws.String(string(buf)),
					TaskToken: r.TaskToken,
				}); err != nil {
					return fmt.Errorf("can't send task success: %w", err)
				}
			case err == ErrDeferTaskTokenResponse:
			default:
				log.Printf("worker %q task failed: %v", workerName, err)
				const max = 200
				myError := err.Error()
				if len(myError) > max {
					myError = myError[:max-3] + "..."
				}
				if _, err := svc.SendTaskFailure(c, &sfn.SendTaskFailureInput{
					TaskToken: r.TaskToken,
					Error:     aws.String(myError),
					Cause:     aws.String(err.Error()),
				}); err != nil {
					return fmt.Errorf("can't send task failure: %w", err)
				}
			}
			return nil
		}(); err != nil {
			log.Printf("error in long poll for %q: %v", workerName, err)
			time.Sleep(time.Second)
		}
	}
}
