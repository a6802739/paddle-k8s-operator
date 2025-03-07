// Copyright 2019 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource_test

import (
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	paddleresource "github.com/paddlepaddle/paddlejob/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestNeedGPU(t *testing.T) {
	var j paddleresource.PaddleJob
	assert.False(t, j.NeedGPU())

	q, err := resource.ParseQuantity("1")
	assert.Nil(t, err)

	j.Spec.Trainer.Resources.Limits = make(v1.ResourceList)
	j.Spec.Trainer.Resources.Limits[v1.ResourceNvidiaGPU] = q
	assert.True(t, j.NeedGPU())
}
