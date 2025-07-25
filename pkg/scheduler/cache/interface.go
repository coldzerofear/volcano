/*
Copyright 2017 The Kubernetes Authors.
Copyright 2017-2025 The Volcano Authors.

Modifications made by Volcano authors:
- Extended cache interface with enhanced binding, client access and metrics support
- Added new interfaces for pre-binding, batch binding and status updating

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	vcclient "volcano.sh/apis/pkg/client/clientset/versioned"
	"volcano.sh/volcano/pkg/scheduler/api"
)

// Cache collects pods/nodes/queues information
// and provides information snapshot
type Cache interface {
	// Run start informer
	Run(stopCh <-chan struct{})

	// Snapshot deep copy overall cache information into snapshot
	Snapshot() *api.ClusterInfo

	// WaitForCacheSync waits for all cache synced
	WaitForCacheSync(stopCh <-chan struct{})

	// AddBindTask binds Task to the target host.
	// TODO(jinzhej): clean up expire Tasks.
	AddBindTask(bindCtx *BindContext) error

	// BindPodGroup Pod/PodGroup to cluster
	BindPodGroup(job *api.JobInfo, cluster string) error

	// Evict evicts the task to release resources.
	Evict(task *api.TaskInfo, reason string) error

	// RecordJobStatusEvent records related events according to job status.
	// Deprecated: remove it after removed PDB support.
	RecordJobStatusEvent(job *api.JobInfo, updatePG bool)

	// UpdateJobStatus puts job in backlog for a while.
	UpdateJobStatus(job *api.JobInfo, updatePGStatus, updatePGAnnotations bool) (*api.JobInfo, error)

	// UpdateQueueStatus update queue status.
	UpdateQueueStatus(queue *api.QueueInfo) error

	// Client returns the kubernetes clientSet, which can be used by plugins
	Client() kubernetes.Interface

	// VCClient returns the volcano clientSet, which can be used by plugins
	VCClient() vcclient.Interface

	// ClientConfig returns the rest config
	ClientConfig() *rest.Config

	UpdateSchedulerNumaInfo(sets map[string]api.ResNumaSets) error

	// SharedInformerFactory return scheduler SharedInformerFactory
	SharedInformerFactory() informers.SharedInformerFactory

	// SetMetricsConf set the metrics server related configuration
	SetMetricsConf(conf map[string]string)

	// EventRecorder returns the event recorder
	EventRecorder() record.EventRecorder

	// RegisterBinder registers the passed binder to the cache's binderRegistry
	RegisterBinder(name string, binder interface{})

	// SharedDRAManager returns the shared DRAManager
	SharedDRAManager() framework.SharedDRAManager
}

// Binder interface for binding task and hostname
type Binder interface {
	Bind(kubeClient kubernetes.Interface, tasks []*api.TaskInfo) map[api.TaskID]string
}

// Evictor interface for evict pods
type Evictor interface {
	Evict(pod *v1.Pod, reason string) error
}

// StatusUpdater updates pod with given PodCondition
type StatusUpdater interface {
	UpdatePodStatus(pod *v1.Pod) (*v1.Pod, error)
	UpdatePodGroup(pg *api.PodGroup) (*api.PodGroup, error)
	UpdateQueueStatus(queue *api.QueueInfo) error
}

// BatchBinder updates podgroup or job information
type BatchBinder interface {
	Bind(job *api.JobInfo, cluster string) (*api.JobInfo, error)
}

type PreBinder interface {
	PreBind(ctx context.Context, bindCtx *BindContext) error

	// PreBindRollBack is called when the pre-bind or bind fails.
	PreBindRollBack(ctx context.Context, bindCtx *BindContext)
}
