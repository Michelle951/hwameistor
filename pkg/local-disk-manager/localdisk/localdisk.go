package localdisk

import (
	"context"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crmanager "sigs.k8s.io/controller-runtime/pkg/manager"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

// Controller The smallest unit to be processed here should be the disk.
// The main thing to do is how to convert the local disk into resources in the cluster
type Controller struct {
	// Mgr k8s runtime controller
	Mgr crmanager.Manager

	// Namespace is the namespace in which v1alpha1 is installed
	NameSpace string

	// NodeName is the node in which v1alpha1 is installed
	NodeName string
}

func NewController(mgr crmanager.Manager) Controller {
	return Controller{
		Mgr:       mgr,
		NameSpace: utils.GetNamespace(),
		NodeName:  utils.GetNodeName(),
	}
}

func (ctr Controller) CreateLocalDisk(ld v1alpha1.LocalDisk) error {
	log.Debugf("Create localDisk for %+v", ld)
	// Setup disk.spec.Reserved if found filesystem or partitions on it already
	if ld.Spec.HasPartition {
		ld.Spec.Reserved = true
	}
	return ctr.Mgr.GetClient().Create(context.Background(), &ld)
}

func (ctr Controller) UpdateLocalDisk(ld v1alpha1.LocalDisk) error {
	newLd := ld.DeepCopy()
	key := client.ObjectKey{Name: ld.GetName(), Namespace: ""}

	oldLd, err := ctr.GetLocalDisk(key)
	if err != nil {
		return err
	}

	ctr.mergerLocalDisk(oldLd, newLd)
	return ctr.Mgr.GetClient().Patch(context.Background(), newLd, client.MergeFrom(&oldLd))
}

func (ctr Controller) IsAlreadyExist(ld v1alpha1.LocalDisk) bool {
	key := client.ObjectKey{Name: ld.GetName(), Namespace: ""}
	if lookLd, err := ctr.GetLocalDisk(key); err != nil {
		return false
	} else {
		return ld.GetName() == lookLd.GetName()
	}
}

func (ctr Controller) GetLocalDisk(key client.ObjectKey) (v1alpha1.LocalDisk, error) {
	ld := v1alpha1.LocalDisk{}
	if err := ctr.Mgr.GetClient().Get(context.Background(), key, &ld); err != nil {
		if errors.IsNotFound(err) {
			return ld, nil
		}
		return ld, err
	}

	return ld, nil
}

func (ctr Controller) ConvertDiskToLocalDisk(disk manager.DiskInfo) (ld v1alpha1.LocalDisk) {
	ld, _ = localdisk.NewBuilder().WithName(ctr.GenLocalDiskName(disk)).
		SetupState().
		SetupRaidInfo(disk.Raid).
		SetupSmartInfo(disk.Smart).
		SetupUUID(disk.GenerateUUID()).
		SetupAttribute(disk.Attribute).
		SetupPartitionInfo(disk.Partitions).
		SetupNodeName(ctr.NodeName).
		Build()
	return
}

func (ctr Controller) mergerLocalDisk(oldLd v1alpha1.LocalDisk, newLd *v1alpha1.LocalDisk) {
	newLd.Status = oldLd.Status
	newLd.TypeMeta = oldLd.TypeMeta
	newLd.ObjectMeta = oldLd.ObjectMeta
	newLd.Spec.ClaimRef = oldLd.Spec.ClaimRef
	newLd.Spec.Owner = oldLd.Spec.Owner
}

func (ctr Controller) GenLocalDiskName(disk manager.DiskInfo) string {
	return utils.ConvertNodeName(ctr.NodeName) + "-" + disk.Name
}
