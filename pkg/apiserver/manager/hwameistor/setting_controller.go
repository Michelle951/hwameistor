package hwameistor

import (
	"context"
	"fmt"
	hoapisv1alpha1 "github.com/hwameistor/hwameistor-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
)

const (
	drbdJobPrefix = "drbd-adapter"
	drbdVersion   = "drbd-version"
)

// SettingController
type SettingController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
}

// NewSettingController
func NewSettingController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *SettingController {
	return &SettingController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
	}
}

// EnableHighAvailability
func (settingController *SettingController) EnableHighAvailability() (*hwameistorapi.DrbdEnableSettingRspBody, error) {
	var RspBody = &hwameistorapi.DrbdEnableSettingRspBody{}
	clusterList := &hoapisv1alpha1.ClusterList{}
	if err := settingController.Client.List(context.TODO(), clusterList); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to list clusterList")
		} else {
			log.Info("Not found the clusterList")
		}
		return RspBody, err
	}

	// for _, cluster := range clusterList.Items {
	// 	if cluster.Name == OperatorClusterName {
	// 		drbdSpec := &hoapisv1alpha1.DRBDSpec{}
	// 		drbdSpec.Enable = true
	// 		cluster.Spec.DRBD = drbdSpec

	// 		if err := settingController.Client.Update(context.TODO(), &cluster); err != nil {
	// 			return RspBody, err
	// 		}
	// 		var drbdEnableSetting = &hwameistorapi.DrbdEnableSetting{}
	// 		drbdEnableSetting.Enable = true
	// 		drbdEnableSetting.State = hwameistorapi.DrbdModuleStatusEnabled
	// 		drbdEnableSetting.Version = "v0.0.1"
	// 		RspBody.DrbdEnableSetting = drbdEnableSetting
	// 		break
	// 	}
	// }

	return RspBody, nil
}

// GetDRBDSetting
func (settingController *SettingController) GetDRBDSetting() (*hwameistorapi.DrbdEnableSetting, error) {

	jobs, err := settingController.getDrbdJobListByNS()
	if err != nil {
		log.WithError(err).Error("Failed to getJobListByNS")
		return nil, err
	}
	var drbdSetting = &hwameistorapi.DrbdEnableSetting{}
	for _, job := range jobs {
		if label, exists := job.Labels[drbdVersion]; exists {
			fmt.Println("GetDRBDSetting label = %v", label)
			drbdSetting.Version = label
		}
		fmt.Println("job.Status.Succeeded = %v, job.Status.Active = %v", job.Status.Succeeded, job.Status.Active)
		if job.Status.Succeeded != 0 && (job.Status.Active == job.Status.Succeeded) {
			drbdSetting.State = hwameistorapi.DrbdModuleStatusEnabled
			drbdSetting.Enable = true
		} else {
			drbdSetting.State = hwameistorapi.DrbdModuleStatusDisabled
			drbdSetting.Enable = false
		}
	}

	return drbdSetting, nil
}

// getDrbdJobListByNS 获取当前namespace下同环境的job Item实例
func (settingController *SettingController) getDrbdJobListByNS() ([]v1.Job, error) {
	var jobList, err = settingController.clientset.BatchV1().Jobs("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// 过滤非同前缀的Job
	var items []v1.Job
	for _, v := range jobList.Items {
		fmt.Println("getDrbdJobListByNS v.Name = %v", v.Name)
		if strings.HasPrefix(v.Name, drbdJobPrefix) {
			items = append(items, v)
		}
	}

	return items, nil
}
