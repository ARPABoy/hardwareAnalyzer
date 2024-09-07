package zfs

import (
	"bytes"
	"fmt"
	"hardwareAnalyzer/utils"
	"io/fs"
	"testing"
)

// TestCheckZFSRaid fs.DirEntry interface implementation in order to mock []fs.DirEntry output in TestCheckZFSRaid* unit tests
type CustomDirEntry struct {
	name  string
	isDir bool
}

func (e *CustomDirEntry) Name() string {
	return e.name
}

func (e *CustomDirEntry) IsDir() bool {
	return e.isDir
}

func (e *CustomDirEntry) Type() fs.FileMode {
	return fs.FileMode(0)
}

func (e *CustomDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

// ----------------------------------------------------------------------------------------------------------------------------

// Test GetZFSPoolSize
func TestGetZFSPoolSize(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer

		outputStdout.WriteString(`
			NAME   SIZE  ALLOC   FREE  EXPANDSZ   FRAG    CAP  DEDUP  HEALTH  ALTROOT
			TESTPOOL    123T  74,9T  28,1T         -     5%    72%  1.00x  ONLINE  -
		`)
		return &outputStdout, &outputStderr, nil
	}

	poolSize, err := GetZFSPoolSize("TESTPOOL")

	if err != nil {
		t.Fatalf(`TestGetZFSPoolSize returned error: %s`, err)
	}

	poolSizeWanted := "123 TB"
	if poolSize != poolSizeWanted {
		t.Fatalf(`TestGetZFSPoolSize poolSize: %s should match: %v`, poolSize, poolSizeWanted)
	}
}

// Test GetZFSPoolSize Error
func TestGetZFSPoolSizeError(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer
		return &outputStdout, &outputStderr, fmt.Errorf("RANDOM ERROR")
	}

	_, err := GetZFSPoolSize("TESTPOOL")

	if err == nil {
		t.Fatalf(`TestGetZFSPoolSizeError returned error != nil`)
	}
}

// Test GetZFSPoolSize Empty
func TestGetZFSPoolSizeEmpty(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
	}()

	// Mocked function, this way we can run unit tests in servers without hardware raid controller installed.
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer
		outputStdout.WriteString("")
		return &outputStdout, &outputStderr, nil
	}

	poolSize, err := GetZFSPoolSize("TESTPOOL")

	if err != nil {
		t.Fatalf(`TestGetZFSPoolSizeEmpty returned error: %s`, err)
	}
	if poolSize != "Unknown" {
		t.Fatalf(`TestGetZFSPoolSizeEmpty poolSize: %s should match Unknown`, poolSize)
	}
}

// Test CheckZFSRaid
func TestCheckZFSRaid(t *testing.T) {
	// Copy original functions content
	getZFSsOri := GetZFSs
	// unmock functions content
	defer func() {
		GetZFSs = getZFSsOri
	}()

	// Mocked function, we return a []fs.DirEntry with at least one dir
	GetZFSs = func() ([]fs.DirEntry, error) {
		entries := []fs.DirEntry{
			//&CustomDirEntry{name: "file1.txt", isDir: false},
			&CustomDirEntry{name: "dir1", isDir: true},
		}

		return entries, nil
	}

	zfsRaidCheck, err := CheckZFSRaid()

	if err != nil {
		t.Fatalf(`TestCheckZFSRaid returned error: %s`, err)
	}

	if !zfsRaidCheck {
		t.Fatalf(`TestCheckZFSRaid zfsRaidCheck: %v must be TRUE`, zfsRaidCheck)
	}
}

// Test CheckZFSRaid No Pools
func TestCheckZFSRaidNoPools(t *testing.T) {
	// Copy original functions content
	getZFSsOri := GetZFSs
	// unmock functions content
	defer func() {
		GetZFSs = getZFSsOri
	}()

	// Mocked function, we return a []fs.DirEntry with at least one dir
	GetZFSs = func() ([]fs.DirEntry, error) {
		entries := []fs.DirEntry{
			&CustomDirEntry{name: "file1.txt", isDir: false},
			//&CustomDirEntry{name: "dir1", isDir: true},
		}

		return entries, nil
	}

	zfsRaidCheck, err := CheckZFSRaid()

	if err != nil {
		t.Fatalf(`TestCheckZFSRaidNoPools returned error: %s`, err)
	}

	if zfsRaidCheck {
		t.Fatalf(`TestCheckZFSRaidNoPools zfsRaidCheck: %v must be FALSE`, zfsRaidCheck)
	}
}

// Test CheckZFSRaid No Kernel Support
func TestCheckZFSRaidNoKernelSupport(t *testing.T) {
	// Copy original functions content
	getZFSsOri := GetZFSs
	// unmock functions content
	defer func() {
		GetZFSs = getZFSsOri
	}()

	// Mocked function, we return a []fs.DirEntry with at least one dir
	GetZFSs = func() ([]fs.DirEntry, error) {
		entries := []fs.DirEntry{
			&CustomDirEntry{name: "file1.txt", isDir: false},
			//&CustomDirEntry{name: "dir1", isDir: true},
		}
		return entries, fmt.Errorf("RANDOM ERROR")
	}

	zfsRaidCheck, err := CheckZFSRaid()

	if err != nil {
		t.Fatalf(`TestCheckZFSRaidNoKernelSupport returned error: %s`, err)
	}

	if zfsRaidCheck {
		t.Fatalf(`TestCheckZFSRaidNoPools zfsRaidCheck: %v must be FALSE`, zfsRaidCheck)
	}
}

// Test ProcessZFSRaid
func TestProcessZFSRaid(t *testing.T) {
	// Copy original functions content
	getCommandOutputOri := utils.GetCommandOutput
	getZFSPoolSizeOri := GetZFSPoolSize
	getDiskDataOri := utils.GetDiskData
	getDiskPartitionSizeOri := utils.GetDiskPartitionSize
	// unmock functions content
	defer func() {
		utils.GetCommandOutput = getCommandOutputOri
		GetZFSPoolSize = getZFSPoolSizeOri
		utils.GetDiskData = getDiskDataOri
		utils.GetDiskPartitionSize = getDiskPartitionSizeOri
	}()

	// Mocked function
	utils.GetCommandOutput = func(manufacturer string, callingFunction string, command string) (*bytes.Buffer, *bytes.Buffer, error) {
		//fmt.Println("-- Executing mocked getCommandOutput function")
		var outputStdout, outputStderr bytes.Buffer

		outputStdout.WriteString(`
		pool: lxd
		state: ONLINE
		config:

			NAME        STATE     READ WRITE CKSUM
			lxd         ONLINE       0     0     0
			raidz2-0  ONLINE         0     0     0
				sdb     ONLINE       0     0     0
				sdc     ONLINE       0     0     0
				sdd     ONLINE       0     0     0
				sde     ONLINE       0     0     0
				sdf     ONLINE       0     0     0
				sdg     ONLINE       0     0     0

		errors: No known data errors
		`)
		return &outputStdout, &outputStderr, nil
	}

	// Mocked function
	GetZFSPoolSize = func(poolName string) (string, error) {
		poolSize := "60 TB"
		return poolSize, nil
	}

	// Mocked function
	utils.GetDiskData = func(diskDrive string) (string, string, string, string, error) {
		diskSerialNumber := "SERIALNUMBER-" + diskDrive
		diskModel := "MODEL-" + diskDrive
		diskIntf := "SATA"
		diskMedium := "SSD"
		return diskSerialNumber, diskModel, diskIntf, diskMedium, nil
	}

	// Mocked function
	utils.GetDiskPartitionSize = func(diskDrive string) (string, error) {
		diskSize := "10 TB"
		return diskSize, nil
	}

	// test it
	newControllers, newPools, newRaids, err := ProcessZFSRaid("zfs")

	if err != nil {
		t.Fatalf(`TestProcessZFSRaid returned error: %s`, err)
	}

	// fmt.Println("------------- newControllers --------------")
	// spew.Dump(newControllers)
	for _, controller := range newControllers {
		controllerIdWanted := "zfs-0"
		if controller.Id != controllerIdWanted {
			t.Fatalf(`TestProcessZFSRaid: controller.Id: %v muts match %v`, controller.Id, controllerIdWanted)
		}

		controllerManufacturerWanted := "zfs"
		if controller.Manufacturer != controllerManufacturerWanted {
			t.Fatalf(`TestProcessZFSRaid: controller.Manufacturer: %v muts match %v`, controller.Manufacturer, controllerManufacturerWanted)
		}

		controllerModelWanted := "ZFS"
		if controller.Model != controllerModelWanted {
			t.Fatalf(`TestProcessZFSRaid: controller.Model: %v muts match %v`, controller.Model, controllerModelWanted)
		}

		controllerStatusWanted := "Good"
		if controller.Status != controllerStatusWanted {
			t.Fatalf(`TestProcessZFSRaid: controller.Status: %v muts match %v`, controller.Status, controllerStatusWanted)
		}
	}

	// fmt.Println("------------- newPools --------------")
	// spew.Dump(newPools)
	for _, pool := range newPools {
		poolControllerIdWanted := "zfs-0"
		if pool.ControllerId != poolControllerIdWanted {
			t.Fatalf(`TestProcessZFSRaid: pool.ControllerId: %v muts match %v`, pool.ControllerId, poolControllerIdWanted)
		}

		poolNameWanted := "lxd"
		if pool.Name != poolNameWanted {
			t.Fatalf(`TestProcessZFSRaid: pool.Name: %v muts match %v`, pool.Name, poolNameWanted)
		}

		poolStateWanted := "ONLINE"
		if pool.State != poolStateWanted {
			t.Fatalf(`TestProcessZFSRaid: pool.State: %v muts match %v`, pool.State, poolStateWanted)
		}

		poolSizeWanted := "60 TB"
		if pool.Size != poolSizeWanted {
			t.Fatalf(`TestProcessZFSRaid: pool.Size: %v muts match %v`, pool.Size, poolSizeWanted)
		}

		poolOsDeviceWanted := "/lxd"
		if pool.OsDevice != poolOsDeviceWanted {
			t.Fatalf(`TestProcessZFSRaid: pool.OsDevice: %v muts match %v`, pool.OsDevice, poolOsDeviceWanted)
		}
	}

	// fmt.Println("------------- newRaids --------------")
	// spew.Dump(newRaids)
	for _, raid := range newRaids {
		raidControllerIdWanted := "zfs-0"
		if raid.ControllerId != raidControllerIdWanted {
			t.Fatalf(`TestProcessZFSRaid: raid.ControllerId: %v muts match %v`, raid.ControllerId, raidControllerIdWanted)
		}

		raidRaidLevelWanted := 0
		if raid.RaidLevel != raidRaidLevelWanted {
			t.Fatalf(`TestProcessZFSRaid: raid.RaidLevel: %v muts match %v`, raid.RaidLevel, raidRaidLevelWanted)
		}

		raidDgWanted := "lxd"
		if raid.Dg != raidDgWanted {
			t.Fatalf(`TestProcessZFSRaid: raid.Dg: %v muts match %v`, raid.Dg, raidDgWanted)
		}

		raidRaidTypeWanted := "raidz2"
		if raid.RaidType != raidRaidTypeWanted {
			t.Fatalf(`TestProcessZFSRaid: raid.RaidType: %v muts match %v`, raid.RaidType, raidRaidTypeWanted)
		}

		raidStateWanted := "ONLINE"
		if raid.State != raidStateWanted {
			t.Fatalf(`TestProcessZFSRaid: raid.State: %v muts match %v`, raid.State, raidStateWanted)
		}

		raidSizeWanted := ""
		if raid.Size != raidSizeWanted {
			t.Fatalf(`TestProcessZFSRaid: raid.Size: %v muts match %v`, raid.Size, raidSizeWanted)
		}

		for i, raidDisk := range raid.Disks {
			raidDiskControllerIdWanted := "zfs-0"
			if raidDisk.ControllerId != raidDiskControllerIdWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.ControllerId: %v muts match %v`, raidDisk.ControllerId, raidDiskControllerIdWanted)
			}

			raidDiskDgWanted := "lxd"
			if raidDisk.Dg != raidDiskDgWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.Dg: %v muts match %v`, raidDisk.Dg, raidDiskDgWanted)
			}

			raidDiskEidSlotWanted := ""
			if raidDisk.EidSlot != raidDiskEidSlotWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.EidSlot: %v muts match %v`, raidDisk.EidSlot, raidDiskEidSlotWanted)
			}

			raidDiskStateWanted := "ONLINE"
			if raidDisk.State != raidDiskStateWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.State: %v muts match %v`, raidDisk.State, raidDiskStateWanted)
			}

			raidDiskSizeWanted := "10 TB"
			if raidDisk.Size != raidDiskSizeWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.Size: %v muts match %v`, raidDisk.Size, raidDiskSizeWanted)
			}

			raidDiskIntfWanted := "SATA"
			if raidDisk.Intf != raidDiskIntfWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.Intf: %v muts match %v`, raidDisk.Intf, raidDiskIntfWanted)
			}

			raidDiskMediumWanted := "SSD"
			if raidDisk.Medium != raidDiskMediumWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.Medium: %v muts match %v`, raidDisk.Medium, raidDiskMediumWanted)
			}

			raidDiskModelWanted := ""
			if i == 0 {
				raidDiskModelWanted = "MODEL-sdb"
			} else if i == 1 {
				raidDiskModelWanted = "MODEL-sdc"
			} else if i == 2 {
				raidDiskModelWanted = "MODEL-sdd"
			} else if i == 3 {
				raidDiskModelWanted = "MODEL-sde"
			} else if i == 4 {
				raidDiskModelWanted = "MODEL-sdf"
			} else {
				raidDiskModelWanted = "MODEL-sdg"
			}
			if raidDisk.Model != raidDiskModelWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.Model: %v muts match %v`, raidDisk.Model, raidDiskModelWanted)
			}

			raidDiskSerialNumberWanted := ""
			if i == 0 {
				raidDiskSerialNumberWanted = "SERIALNUMBER-sdb"
			} else if i == 1 {
				raidDiskSerialNumberWanted = "SERIALNUMBER-sdc"
			} else if i == 2 {
				raidDiskSerialNumberWanted = "SERIALNUMBER-sdd"
			} else if i == 3 {
				raidDiskSerialNumberWanted = "SERIALNUMBER-sde"
			} else if i == 4 {
				raidDiskSerialNumberWanted = "SERIALNUMBER-sdf"
			} else {
				raidDiskSerialNumberWanted = "SERIALNUMBER-sdg"
			}
			if raidDisk.SerialNumber != raidDiskSerialNumberWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.SerialNumber: %v muts match %v`, raidDisk.SerialNumber, raidDiskSerialNumberWanted)
			}

			raidDiskOsDeviceWanted := ""
			if i == 0 {
				raidDiskOsDeviceWanted = "sdb"
			} else if i == 1 {
				raidDiskOsDeviceWanted = "sdc"
			} else if i == 2 {
				raidDiskOsDeviceWanted = "sdd"
			} else if i == 3 {
				raidDiskOsDeviceWanted = "sde"
			} else if i == 4 {
				raidDiskOsDeviceWanted = "sdf"
			} else {
				raidDiskOsDeviceWanted = "sdg"
			}
			if raidDisk.OsDevice != raidDiskOsDeviceWanted {
				t.Fatalf(`TestProcessZFSRaid: raidDisk.OsDevice: %v muts match %v`, raidDisk.OsDevice, raidDiskOsDeviceWanted)
			}
		}

		raidOsDeviceWanted := ""
		if raid.OsDevice != raidOsDeviceWanted {
			t.Fatalf(`TestProcessZFSRaid: raid.OsDevice: %v muts match %v`, raid.OsDevice, raidOsDeviceWanted)
		}
	}
}
