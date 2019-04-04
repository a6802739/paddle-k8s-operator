package v1

import (
	"encoding/json"
	"fmt"
)

// GPU convert Resource Limit Quantity to int
func (s *PaddleJob) GPU() int {
	q := s.Spec.Trainer.Resources.Limits.NvidiaGPU()
	gpu, ok := q.AsInt64()
	if !ok {
		// FIXME: treat errors
		gpu = 0
	}
	return int(gpu)
}

// NeedGPU returns true if the job need GPU resource to run.
func (s *PaddleJob) NeedGPU() bool {
	return s.GPU() > 0
}

func (s *PaddleJob) String() string {
	b, _ := json.MarshalIndent(s, "", "   ")
	return fmt.Sprintf("%s", b)
}
