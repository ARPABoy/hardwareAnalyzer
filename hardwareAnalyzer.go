package main

// Linux Raid/disks configuration detection tool with auto contained disk tools:
// MegaRaid/PERC/SAS2IRCU/ADAPTEC/SoftRAID/ZFS/Btrfs/LVM/Disks Linux support.

// When net and user functions are involved Go compiles dynamically linked binaries
// go-memexec module uses osusergo functionalities, so force to compile statically.
// GOOS="linux" GOARCH="amd64" go build -tags netgo,osusergo -ldflags "-X main.Author=kr0m -X main.AuthorURL=https://alfaexploit.com"

// N=$(wc -l serverList|awk '{print$1}') && i=1 && for SERVER in $(<serverList); do echo "$i/$N SERVER: $SERVER" && scp -J fw ~/hardwareAnalyzer/hardwareAnalyzer $SERVER: && let i=$i+1; done
// pssh -O StrictHostKeyChecking=no -i -h serverList "sudo ./hardwareAnalyzer"

// arcconf dynamic binary: repacked using packelf.sh
// git clone https://github.com/oufm/packelf.git && cd packelf && ./packelf.sh ~/arcconf ./arcconf && cp arcconf ~/hardwareAnalyzer/

// zpool
// git clone https://github.com/oufm/packelf.git && cd packelf && ./packelf.sh /usr/sbin/zpool ./zpool && cp zpool ~/hardwareAnalyzer/

// btrfs
// git clone https://github.com/oufm/packelf.git && cd packelf && ./packelf.sh /usr/bin/btrfs ./btrfs && cp btrfs ~/hardwareAnalyzer/

import (
	"flag"
	"fmt"
	"hardwareAnalyzer/adaptec"
	"hardwareAnalyzer/btrfs"
	"hardwareAnalyzer/hardwarecontrollerscommon"
	"hardwareAnalyzer/lvm"
	"hardwareAnalyzer/megaraidpercsas2ircu"
	"hardwareAnalyzer/regulardisks"
	"hardwareAnalyzer/softraid"
	"hardwareAnalyzer/utils"
	"hardwareAnalyzer/zfs"

	//"github.com/davecgh/go-spew/spew"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/inancgumus/screen"
)

// We use init to initializar flags in order to not get the error: flag redefined when unit testing code
var showInfo *bool

func init() {
	showInfo = flag.Bool("showInfo", false, "Show binary information.")
}

func checkHardware() (bool, bool, bool, bool, bool, bool, bool, bool) {
	// MegaRaid check:
	megaRaidCheck, err := megaraidpercsas2ircu.CheckMegaraidPerc("mega")
	if err != nil {
		megaRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that MegaRaid controllers cant be checked, continuing.")
	}

	// PERC check:
	color.Set(color.FgCyan)
	fmt.Println("")
	percRaidCheck, err := megaraidpercsas2ircu.CheckMegaraidPerc("perc")
	if err != nil {
		percRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that Dell-PERC controllers cant be checked, continuing.")
	}

	// SAS2IRCU check:
	fmt.Println("")
	color.Set(color.FgCyan)
	sas2ircuRaidCheck, err := megaraidpercsas2ircu.CheckSas2ircuRaid()
	if err != nil {
		sas2ircuRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that MegaRaid SAS2IRCU controllers cant be checked, continuing.")
	}

	// ARCCONF check:
	fmt.Println("")
	color.Set(color.FgCyan)
	// Dont get error return value as executing arcconf in a system without Adaptec controllers always returns: exit status 127
	adaptecRaidCheck, err := adaptec.CheckAadaptecRaid()
	if err != nil {
		adaptecRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that Adaptec controllers cant be checked, continuing.")
	}

	// SoftRaid check:
	fmt.Println("")
	color.Set(color.FgCyan)
	softRaidCheck, err := softraid.CheckSoftRaid()
	if err != nil {
		softRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that Softraid configutations cant be checked, continuing.")
	}

	// ZFS check:
	fmt.Println("")
	color.Set(color.FgCyan)
	zfsRaidCheck, err := zfs.CheckZFSRaid()
	if err != nil {
		zfsRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that ZFS configutations cant be checked, continuing.")
	}

	// Btrfs check:
	fmt.Println("")
	color.Set(color.FgCyan)
	btrfsRaidCheck, err := btrfs.CheckBtrfsRaid()
	if err != nil {
		btrfsRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that Btrfs configutations cant be checked, continuing.")
	}

	// LVM check:
	fmt.Println("")
	color.Set(color.FgCyan)
	lvmRaidCheck, err := lvm.CheckLVMRaid()
	if err != nil {
		lvmRaidCheck = false
		color.Red("++ ERROR: %s", err)
		color.Cyan("Dont worry, it only implies that LVM configutations cant be checked, continuing.")
	}

	return megaRaidCheck, percRaidCheck, sas2ircuRaidCheck, adaptecRaidCheck, softRaidCheck, zfsRaidCheck, btrfsRaidCheck, lvmRaidCheck
}

func inquireHardwareConfiguration(megaRaidCheck, percRaidCheck, sas2ircuRaidCheck, adaptecRaidCheck, softRaidCheck, zfsRaidCheck, btrfsRaidCheck, lvmRaidCheck bool) ([]utils.ControllerStruct, []utils.PoolStruct, []utils.VolumeGroupStruct, []utils.RaidStruct, []utils.NoRaidDiskStruct) {
	// Get detected raid info:
	fmt.Println("")
	color.Set(color.FgCyan)

	// Variables where final data will be stored
	var controllers []utils.ControllerStruct
	var pools []utils.PoolStruct
	var volumeGroups []utils.VolumeGroupStruct
	var raids []utils.RaidStruct
	var noRaidDisks []utils.NoRaidDiskStruct

	// Temp variables for each check
	var newControllers []utils.ControllerStruct
	var newRaids []utils.RaidStruct
	var newNoRaidDisks []utils.NoRaidDiskStruct

	if megaRaidCheck {
		newControllers, newRaids, newNoRaidDisks, err := megaraidpercsas2ircu.ProcessHWMegaraidPercRaid("mega")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newRaids) > 0 {
			for _, newRaid := range newRaids {
				raids = append(raids, newRaid)
			}
		}
		if len(newNoRaidDisks) > 0 {
			for _, newNoRaidDisk := range newNoRaidDisks {
				noRaidDisks = append(noRaidDisks, newNoRaidDisk)
			}
		}
	}

	color.Set(color.FgCyan)
	if percRaidCheck {
		newControllers, newRaids, newNoRaidDisks, err := megaraidpercsas2ircu.ProcessHWMegaraidPercRaid("perc")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newRaids) > 0 {
			for _, newRaid := range newRaids {
				raids = append(raids, newRaid)
			}
		}
		if len(newNoRaidDisks) > 0 {
			for _, newNoRaidDisk := range newNoRaidDisks {
				noRaidDisks = append(noRaidDisks, newNoRaidDisk)
			}
		}
	}

	color.Set(color.FgCyan)
	if sas2ircuRaidCheck {
		newControllers, newRaids, newNoRaidDisks, err := megaraidpercsas2ircu.ProcessHWSas2ircuRaid("sas2ircu")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newRaids) > 0 {
			for _, newRaid := range newRaids {
				raids = append(raids, newRaid)
			}
		}
		if len(newNoRaidDisks) > 0 {
			for _, newNoRaidDisk := range newNoRaidDisks {
				noRaidDisks = append(noRaidDisks, newNoRaidDisk)
			}
		}
	}

	color.Set(color.FgCyan)
	if adaptecRaidCheck {
		newControllers, newRaids, newNoRaidDisks, err := adaptec.ProcessHWAdaptecRaid("adaptec")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newRaids) > 0 {
			for _, newRaid := range newRaids {
				raids = append(raids, newRaid)
			}
		}
		if len(newNoRaidDisks) > 0 {
			for _, newNoRaidDisk := range newNoRaidDisks {
				noRaidDisks = append(noRaidDisks, newNoRaidDisk)
			}
		}
	}

	// Softraid requires extra steps like hardwarecontrollerscommon.CheckJbodDisks and hardwarecontrollerscommon.CheckHardRaidDisks in order to rename disks if required and fill model, medium disk info
	color.Set(color.FgCyan)
	if softRaidCheck {
		newControllers, newRaids, err := softraid.ProcessSoftRaid("softraid")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check JBOD disks(noRaidDisks) against SoftRaid disks
		if err = hardwarecontrollerscommon.CheckJbodDisks(newRaids, noRaidDisks); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newRaids) > 0 {
			for _, newRaid := range newRaids {
				raids = append(raids, newRaid)
			}
		}
	}

	// ZFS requires extra steps like hardwarecontrollerscommon.CheckJbodDisks, softraid.CheckSoftRaidDisks and hardwarecontrollerscommon.CheckHardRaidDisks in order to rename disks if required and fill model, medium disk info
	color.Set(color.FgCyan)
	if zfsRaidCheck {
		newControllers, newPools, newRaids, err := zfs.ProcessZFSRaid("zfs")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check JBOD disks(noRaidDisks) against ZFS disks
		if err = hardwarecontrollerscommon.CheckJbodDisks(newRaids, noRaidDisks); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check MD disks against ZFS disks
		if err = softraid.CheckSoftRaidDisks(newRaids, raids); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check HardRaid disks against ZFS disks
		if err = hardwarecontrollerscommon.CheckHardRaidDisks(newRaids, raids); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newPools) > 0 {
			for _, newPool := range newPools {
				pools = append(pools, newPool)
			}
		}
		if len(newRaids) > 0 {
			for _, newVdev := range newRaids {
				raids = append(raids, newVdev)
			}
		}
	}

	// Btrfs requires extra steps like hardwarecontrollerscommon.CheckJbodDisks, softraid.checkSoftRaidDisks and hardwarecontrollerscommon.CheckHardRaidDisks in order to rename disks if required and fill model, medium disk info
	color.Set(color.FgCyan)
	if btrfsRaidCheck {
		newControllers, newRaids, err := btrfs.ProcessBtrfsRaid("btrfs")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check JBOD disks(noRaidDisks) against Btrfs disks
		if err = hardwarecontrollerscommon.CheckJbodDisks(newRaids, noRaidDisks); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check MD disks against Btrfs disks
		if err = softraid.CheckSoftRaidDisks(newRaids, raids); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check HardRaid disks against Btrfs disks
		if err = hardwarecontrollerscommon.CheckHardRaidDisks(newRaids, raids); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newRaids) > 0 {
			for _, newRaid := range newRaids {
				raids = append(raids, newRaid)
			}
		}
		if len(newNoRaidDisks) > 0 {
			for _, newNoRaidDisks := range newNoRaidDisks {
				noRaidDisks = append(noRaidDisks, newNoRaidDisks)
			}
		}
	}

	// LVM requires extra steps like hardwarecontrollerscommon.CheckJbodDisks, softraid.CheckSoftRaidDisks and hardwarecontrollerscommon.CheckHardRaidDisks in order to rename disks if required and fill model, medium disk info
	color.Set(color.FgCyan)
	if lvmRaidCheck {
		newControllers, newVolumeGroups, newRaids, err := lvm.ProcessLVMRaid("lvm")
		if err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check JBOD disks(noRaidDisks) against LVM disks
		if err = hardwarecontrollerscommon.CheckJbodDisks(newRaids, noRaidDisks); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check MD disks against LVM disks
		if err = softraid.CheckSoftRaidDisks(newRaids, raids); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Check HardRaid disks against LVM disks
		if err = hardwarecontrollerscommon.CheckHardRaidDisks(newRaids, raids); err != nil {
			color.Red("++ ERROR: %s", err)
		}

		// Append controllers, raids and noraiddisks to already existent
		if len(newControllers) > 0 {
			for _, newController := range newControllers {
				controllers = append(controllers, newController)
			}
		}
		if len(newVolumeGroups) > 0 {
			for _, newVolumeGroup := range newVolumeGroups {
				volumeGroups = append(volumeGroups, newVolumeGroup)
			}
		}
		if len(newRaids) > 0 {
			for _, newRaid := range newRaids {
				raids = append(raids, newRaid)
			}
		}
	}

	// Regular disks
	color.Set(color.FgCyan)
	newControllers, newRaids, err := regulardisks.ProcessRegularDisks(raids, noRaidDisks)
	if err != nil {
		color.Red("++ ERROR: %s", err)
	}

	// Try to fill the most information about disks
	err = utils.CheckPreviousDiskData(newRaids, raids)
	if err != nil {
		color.Red("++ ERROR: %s", err)
	}

	// Append controllers, raids and noraiddisks to already existent
	if len(newControllers) > 0 {
		for _, newController := range newControllers {
			controllers = append(controllers, newController)
		}
	}
	if len(newRaids) > 0 {
		for _, newRaid := range newRaids {
			raids = append(raids, newRaid)
		}
	}

	return controllers, pools, volumeGroups, raids, noRaidDisks
}

func main() {
	version := "2.8"
	codename := "Sistine Chapel"

	// Set default font color:
	color.Set(color.FgCyan)

	screen.MoveTopLeft()
	screen.Clear()
	fmt.Println("########################################################################################")
	fmt.Printf("| HardwareAnalyzer v%v - CodeName: %v %v                                   |\n", version, codename, emoji.LatinCross)
	fmt.Println("| Coded by kr0m - MegaRaid/PERC/SAS2IRCU/ADAPTEC/SoftRAID/ZFS/Btrfs/LVM/Disks support. |")
	fmt.Println("########################################################################################")
	fmt.Println("")

	if !utils.IsRoot() {
		color.Red("++ ERROR: Binary must be run under root privileges.")
		fmt.Println("")
		return
	}

	var err error

	var isSupported bool
	if isSupported, err = utils.SupportedOS(); !isSupported {
		color.Red("++ ERROR: %s", err)
		fmt.Println("")
		return
	}

	// -info command:
	flag.Parse()
	if *showInfo {
		color.Set(color.FgCyan)

		screen.MoveTopLeft()
		screen.Clear()
		fmt.Println("#################################################################################################################")
		fmt.Printf("| HardwareAnalyzer v%v - CodeName: %v %v                                                            |\n", version, codename, emoji.LatinCross)
		fmt.Println("| Coded by kr0m(https://alfaexploit.com) - MegaRaid/PERC/SAS2IRCU/ADAPTEC/SoftRAID/ZFS/Btrfs/LVM/Disks support. |")
		fmt.Println("|   - storcli: Linux/x86-64-static v007.1408.0000.0000 Apr 16, 2020                                             |")
		fmt.Println("|   - percCLI: Linux/x86-64-static Ver 007.0127.0000.0000 July 13, 2017                                         |")
		fmt.Println("|   - sas2ircu: Linux/x86-64-static Version 20.00.00.00 (2014.09.18)                                            |")
		fmt.Println("|   - arcconf: Linux/x86-64 Version 2.05 (B22932) - Statically repacked: packelf.sh                             |")
		fmt.Println("|   - zpool: Linux/x86-64 Version zfs-2.1.5-1 - Statically repacked: packelf.sh                                 |")
		fmt.Println("|   - btrfs: Linux/x86-64 Version 5.16.2-1 - Statically repacked: packelf.sh                                    |")
		fmt.Println("|   - lvm: Linux/x86-64 Version 2.03.11(2) - Statically repacked: packelf.sh                                    |")
		fmt.Println("#################################################################################################################")
		fmt.Println("")

		return
	}

	megaRaidCheck, percRaidCheck, sas2ircuRaidCheck, adaptecRaidCheck, softRaidCheck, zfsRaidCheck, btrfsRaidCheck, lvmRaidCheck := checkHardware()

	controllers, pools, volumeGroups, raids, noRaidDisks := inquireHardwareConfiguration(megaRaidCheck, percRaidCheck, sas2ircuRaidCheck, adaptecRaidCheck, softRaidCheck, zfsRaidCheck, btrfsRaidCheck, lvmRaidCheck)

	// Show gathered raid info:
	utils.ShowGatheredData(controllers, pools, volumeGroups, raids, noRaidDisks)
	fmt.Println("")
}

// Anyadir doc con comentarios al principio de cada package y antes de cada funcion
// Package comments should begin with “Package” followed by the package name: Package mypackage enables widget management.
// Function comments should begin with the name of the function they describe: MyFunction converts widgets to gizmos.
