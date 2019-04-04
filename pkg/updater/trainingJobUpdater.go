package updater

import (
	"fmt"
	"reflect"
	"time"

	log "github.com/golang/glog"

	padv1 "github.com/paddlepaddle/paddlejob/pkg/apis/paddlepaddle/v1"
	paddleJobClient "github.com/paddlepaddle/paddlejob/pkg/client/clientset/versioned"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	retry                 = 5
	retryTime             = 5 * time.Second
	convertedTimerTicker  = 10 * time.Second
	confirmResourceTicker = 5 * time.Second
	eventChLength         = 1000
	factor                = 0.8
)

type paddleJobEventType string

const (
	paddleJobEventDelete paddleJobEventType = "Delete"
	paddleJobEventModify paddleJobEventType = "Modify"
)

type paddleJobEvent struct {
	// pet is the PaddleJobEventType of PaddleJob
	pet paddleJobEventType
	// The job transfer the information fo job
	job *padv1.PaddleJob
}

// PaddleJobUpdater is used to manage a specific PaddleJob
type PaddleJobUpdater struct {
	// Job is the job the PaddleJob manager.
	job *padv1.PaddleJob

	// kubeClient is standard kubernetes client.
	kubeClient kubernetes.Interface

	// PaddleJobClient is the client of PaddleJob.
	paddleJobClient paddleJobClient.Interface

	// Status is the status in memory, update when PaddleJob status changed and update the CRD resource status.
	status padv1.PaddleJobStatus

	// EventCh receives events from the controller, include Modify and Delete.
	// When paddleJobEvent is Delete it will delete all resources
	// The capacity is 1000.
	eventCh chan *paddleJobEvent
}

// NewUpdater creates a new PaddleJobUpdater and start a goroutine to control current job.
func NewUpdater(job *padv1.PaddleJob, kubeClient kubernetes.Interface, paddleJobClient paddleJobClient.Interface) (*PaddleJobUpdater,
	error) {
	log.Infof("NewJobber namespace=%v name=%v", job.Namespace, job.Name)
	updater := &PaddleJobUpdater{
		job:               job,
		kubeClient:        kubeClient,
		paddleJobClient: paddleJobClient,
		status:            job.Status,
		eventCh:           make(chan *paddleJobEvent, eventChLength),
	}
	go updater.start()
	return updater, nil
}

// Notify is used to receive event from controller. While controller receive a informer,
// it will notify updater to process the event. It send event to updater's eventCh.
func (updater *PaddleJobUpdater) notify(te *paddleJobEvent) {
	updater.eventCh <- te
	lene, cape := len(updater.eventCh), cap(updater.eventCh)
	if lene > int(float64(cape)*factor) {
		log.Warning("the len of updater eventCh ", updater.job.Name, " is near to full")
	}
}

// Delete send a delete event to updater, updater will kill the PaddleJob and clear all the resource of the
// PaddleJob.
func (updater *PaddleJobUpdater) Delete() {
	updater.notify(&paddleJobEvent{pet: paddleJobEventDelete})
}

// Modify send a modify event to updater. updater will processing according to the situation.
func (updater *PaddleJobUpdater) Modify(nj *padv1.PaddleJob) {
	updater.notify(&paddleJobEvent{pet: paddleJobEventModify, job: nj})
}

func (updater *PaddleJobUpdater) releaseResource(tp padv1.TrainingResourceType) error {
	resource := new(v1beta1.ReplicaSet)
	switch tp {
	case padv1.Pserver:
		resource = updater.job.Spec.Pserver.ReplicaSpec
	default:
		return fmt.Errorf("unknow resource")
	}
	var replica int32
	resource.Spec.Replicas = &replica
	_, err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Update(resource)
	if errors.IsNotFound(err) {
		return err
	}
	key := "paddle-job-" + tp

	labels := Labels(map[string]string{
		string(key): updater.job.Name,
	})

	selector, _ := labels.LabelsParser()
	options := v1.ListOptions{
		LabelSelector: selector,
	}

	for j := 0; j <= retry; j++ {
		time.Sleep(confirmResourceTicker)
		pl, err := updater.kubeClient.CoreV1().Pods(updater.job.Namespace).List(options)
		if err == nil && len(pl.Items) == 0 {
			return nil
		}
	}
	return updater.kubeClient.CoreV1().Pods(updater.job.Namespace).DeleteCollection(&v1.DeleteOptions{}, options)
}

func (updater *PaddleJobUpdater) releasePserver() error {
	return updater.releaseResource(padv1.Pserver)
}

func (updater *PaddleJobUpdater) releaseTrainer() error {
	labels := Labels(map[string]string{
		"paddle-job": updater.job.Name,
	})
	selector, _ := labels.LabelsParser()
	options := v1.ListOptions{
		LabelSelector: selector,
	}

	return updater.kubeClient.CoreV1().Pods(updater.job.Namespace).DeleteCollection(&v1.DeleteOptions{}, options)
}

func (updater *PaddleJobUpdater) deletePaddleJob() error {
	fault := false

	log.Infof("Start to delete PaddleJob namespace=%v name=%v", updater.job.Namespace, updater.job.Name)

	log.Infof("Release pserver, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
	if err := updater.releasePserver(); err != nil {
		log.Error(err.Error())
		fault = true
	}

	log.Infof("Deleting PaddleJob matadata, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Pserver.ReplicaSpec.Name)
	if err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Delete(updater.job.Spec.Pserver.ReplicaSpec.Name, &v1.DeleteOptions{}); err != nil {
		log.Error("delete pserver replicaset error: ", err.Error())
		fault = true
	}

	log.Infof("Deleting PaddleJob matadata, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
	if err := updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).Delete(updater.job.Spec.Trainer.ReplicaSpec.Name, &v1.DeleteOptions{}); err != nil {
		log.Error("delete trainer replicaset error: ", err.Error())
		fault = true
	}

	log.Infof("Release trainer, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
	if err := updater.releaseTrainer(); err != nil {
		log.Error("release trainer  error: ", err.Error())
		fault = true
	}

	log.Infof("End to delete PaddleJob namespace=%v name=%v", updater.job.Namespace, updater.job.Name)

	if fault {
		return fmt.Errorf("delete resource error namespace=%v name=%v", updater.job.Namespace, updater.job.Name)
	}
	return nil
}

func (updater *PaddleJobUpdater) createResource(tp padv1.TrainingResourceType) error {
	resource := new(v1beta1.ReplicaSet)
	switch tp {
	case padv1.Pserver:
		resource = updater.job.Spec.Pserver.ReplicaSpec
	default:
		return fmt.Errorf("unknown resource")
	}
	for {
		_, err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Get(resource.Name, v1.GetOptions{})
		if errors.IsNotFound(err) {
			log.Infof("Not found to create namespace=%v name=%v resourceName=%v", updater.job.Namespace, updater.job.Name, resource.Name)
			_, err = updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Create(resource)
			if err != nil {
				updater.status.Phase = padv1.PaddleJobPhaseFailed
				updater.status.Reason = "Internal error; create resource error:" + err.Error()
				return err
			}
		} else if err != nil {
			log.Errorf("Get resource error, namespace=%v name=%v resourceName=%v error=%v", updater.job.Namespace, updater.job.Name, resource.Name, err.Error())
			time.Sleep(retryTime)
			continue
		}
		ticker := time.NewTicker(confirmResourceTicker)
		defer ticker.Stop()
		for v := range ticker.C {
			rs, err := updater.kubeClient.ExtensionsV1beta1().ReplicaSets(updater.job.Namespace).Get(resource.Name, v1.GetOptions{})
			log.Infof("Current time %v runing pod is %v, resourceName=%v", v.String(), rs.Status.ReadyReplicas, resource.Name)
			if err != nil && !errors.IsServerTimeout(err) && !errors.IsTooManyRequests(err) {
				updater.status.Reason = "Internal error; create resource error:" + err.Error()
				return err
			}
			if errors.IsServerTimeout(err) || errors.IsTooManyRequests(err) {
				log.Warningf("Connect to kubernetes failed for reasons=%v, retry next ticker", err.Error())
				continue
			}
			if *resource.Spec.Replicas == 0 {
				return fmt.Errorf(" PaddleJob is deleting, namespace=%v name=%v ", updater.job.Namespace, updater.job.Name)

			}
			if rs.Status.ReadyReplicas == *resource.Spec.Replicas {
				log.Infof("Create resource done , namespace=%v name=%v resourceName=%v", updater.job.Namespace, updater.job.Name, resource.Name)
				return nil
			}
		}
	}
}

func (updater *PaddleJobUpdater) createTrainer() error {
	resource := updater.job.Spec.Trainer.ReplicaSpec
	for {
		_, err := updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).Get(resource.Name, v1.GetOptions{})
		if errors.IsNotFound(err) {
			log.Infof("not found to create trainer namespace=%v name=%v", updater.job.Namespace, updater.job.Name)
			_, err = updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).Create(resource)
			if err != nil {
				updater.status.Phase = padv1.PaddleJobPhaseFailed
				updater.status.Reason = "Internal error; create trainer error:" + err.Error()
				return err
			}
		} else if err != nil {
			log.Errorf("Get resource error, namespace=%v name=%v resourceName=%v error=%v", updater.job.Namespace, updater.job.Name, resource.Name, err.Error())
			time.Sleep(retryTime)
			continue
		}
		updater.status.Phase = padv1.PaddleJobPhaseRunning
		updater.status.Reason = ""
		return nil
	}
}

func (updater *PaddleJobUpdater) createPaddleJob() error {
	if err := updater.createResource(padv1.Pserver); err != nil {
		return err
	}
	return updater.createTrainer()
}

func (updater *PaddleJobUpdater) updateCRDStatus() error {
	if reflect.DeepEqual(updater.status, updater.job.Status) {
		return nil
	}
	newPaddleJob := updater.job
	newPaddleJob.Status = updater.status
	newPaddleJob, err := updater.paddleJobClient.PaddlepaddleV1().PaddleJobs(updater.job.Namespace).Update(newPaddleJob)
	if err != nil {
		return err
	}
	updater.job = newPaddleJob
	return nil
}

// parsePaddleJob validates the fields and parses the PaddleJob
func (updater *PaddleJobUpdater) parsePaddleJob() {
	if updater.job == nil {
		updater.status.Phase = padv1.PaddleJobPhaseFailed
		updater.status.Reason = "Internal error; Setup error; job is missing TainingJob"
		return
	}

	var parser DefaultJobParser
	var creatErr error
	updater.job, creatErr = parser.NewPaddleJob(updater.job)

	if creatErr != nil {
		updater.status.Phase = padv1.PaddleJobPhaseFailed
		updater.status.Reason = creatErr.Error()
	} else {
		updater.status.Phase = padv1.PaddleJobPhaseCreating
		updater.status.Reason = ""
	}
}

func (updater *PaddleJobUpdater) getTrainerReplicaStatuses() ([]*padv1.TrainingResourceStatus, error) {
	var replicaStatuses []*padv1.TrainingResourceStatus
	trs := padv1.TrainingResourceStatus{
		TrainingResourceType: padv1.Trainer,
		State:                padv1.ResourceStateNone,
		ResourceStates:       make(map[padv1.ResourceState]int),
	}
	// TODO(ZhengQi): get detail status in future
	replicaStatuses = append(replicaStatuses, &trs)
	return replicaStatuses, nil
}

// GetStatus get PaddleJob status from trainers.
func (updater *PaddleJobUpdater) GetStatus() (*padv1.PaddleJobStatus, error) {

	status := updater.status

	j, err := updater.kubeClient.BatchV1().Jobs(updater.job.Namespace).
		Get(updater.job.Spec.Trainer.ReplicaSpec.Name, v1.GetOptions{})
	if err != nil {
		log.Error("get trainer error:", err.Error())
		return &status, err
	}

	status.ReplicaStatuses, err = updater.getTrainerReplicaStatuses()
	if err != nil {
		log.Error("get trainer replica status error:", err.Error())
	}
	if j.Status.Failed != 0 {
		status.Phase = padv1.PaddleJobPhaseFailed
		status.Reason = "at least one trainer failed!"
	} else {
		if j.Status.Succeeded == *updater.job.Spec.Trainer.ReplicaSpec.Spec.Parallelism && j.Status.Active == 0 {
			status.Phase = padv1.PaddleJobPhaseSucceeded
			status.Reason = "all trainer have succeeded!"
		}
	}

	return &status, nil
}

// Convert is main process to convert PaddleJob to desire status.
func (updater *PaddleJobUpdater) Convert() {
	log.Infof("convert status, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)

	if updater.status.Phase == padv1.PaddleJobPhaseRunning {
		status, err := updater.GetStatus()
		if err != nil {
			log.Error("get current status of trainer from k8s error:", err.Error())
			return
		}
		updater.status = *status.DeepCopy()
		log.Infof("Current status namespace=%v name=%v status=%v : ", updater.job.Namespace, updater.job.Name, status)
		err = updater.updateCRDStatus()
		if err != nil {
			log.Warning("get current status to update PaddleJob status error: ", err.Error())
		}
		if updater.status.Phase == padv1.PaddleJobPhaseSucceeded || updater.status.Phase == padv1.PaddleJobPhaseFailed {
			log.Infof("Release Resource namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
			log.Infof("Release pserver, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Pserver.ReplicaSpec.Name)
			if err := updater.releasePserver(); err != nil {
				log.Error(err.Error())
			}
			log.Infof("Release trainer, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
			if err := updater.releaseTrainer(); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

// InitResource is used to parse PaddleJob and create PaddleJob resources.
func (updater *PaddleJobUpdater) InitResource() {
	if updater.status.Phase == padv1.PaddleJobPhaseNone {
		log.Infof("set up PaddleJob namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
		updater.parsePaddleJob()
		err := updater.updateCRDStatus()
		if err != nil {
			log.Warning("set up PaddleJob to update PaddleJob status error: ", err.Error())
		}
	}

	if updater.status.Phase == padv1.PaddleJobPhaseCreating {
		log.Infof("create PaddleJob namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
		_ = updater.createPaddleJob()
		err := updater.updateCRDStatus()
		if err != nil {
			log.Warning("create PaddleJob to update PaddleJob status error: ", err.Error())
		}
		if updater.status.Phase == padv1.PaddleJobPhaseFailed {
			log.Infof("Release Resource for failed namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
			log.Infof("Release pserver, namespace=%v name=%v", updater.job.Namespace, updater.job.Spec.Trainer.ReplicaSpec.Name)
			if err := updater.releasePserver(); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

// Start is the main process of life cycle of a PaddleJob, including create resources, event process handle and
// status convert.
func (updater *PaddleJobUpdater) start() {
	log.Infof("start updater, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
	go updater.InitResource()

	ticker := time.NewTicker(convertedTimerTicker)
	defer ticker.Stop()
	log.Infof("start ticker, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
	for {
		select {
		case ev := <-updater.eventCh:
			switch ev.pet {
			case paddleJobEventDelete:
				log.Infof("Delete updater, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
				if err := updater.deletePaddleJob(); err != nil {
					log.Errorf(err.Error())
				}
				return
			}
		case <-ticker.C:
			updater.Convert()
			if updater.status.Phase == padv1.PaddleJobPhaseSucceeded || updater.status.Phase == padv1.PaddleJobPhaseFailed {
				if ticker != nil {
					log.Infof("stop ticker for job has done, namespace=%v name=%v: ", updater.job.Namespace, updater.job.Name)
					ticker.Stop()
				}
			}
		}
	}
}
