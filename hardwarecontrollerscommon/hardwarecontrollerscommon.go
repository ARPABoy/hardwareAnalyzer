package hardwarecontrollerscommon

import (
	"bufio"
	"fmt"
	"hardwareAnalyzer/utils"
	"math/big"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Function as variable in order to be possible to mock it from unitary tests
var GetJbodOsDevice = func(manufacturer, controllerId, eidslot string) (string, error) {
	// fmt.Println("-- getJbodOsDevice --")
	// fmt.Println("manufacturer: ", manufacturer)
	// fmt.Println("controllerId: ", controllerId)
	// fmt.Println("eidslot: ", eidslot)

	//fmt.Println("eidslot: ", eidslot)
	eidData := strings.Split(eidslot, ":")
	eid := eidData[0]
	//fmt.Println("eid: ", eid)
	slot := eidData[1]
	//fmt.Println("slot: ", slot)
	switch manufacturer {
	case "mega", "perc":
		command := "/c" + controllerId + " /e" + eid + " /s" + slot + " show all"
		outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "getJbodOsDevice", command)
		if err != nil {
			color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
		}
		if len(outputStderr.String()) != 0 {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			// https://en.wikipedia.org/wiki/World_Wide_Name
			// Ex:   0 Active 6.0Gb/s   0x5000cca23c1237c9
			matched, err := regexp.MatchString(`\d+ Active .*Gb/s\s+0x\w{16}`, line)
			if err != nil {
				color.Red("++ ERROR: Regexp errror %s", err)
			}
			if matched {
				sasAddressData := strings.Fields(line)[3]
				sasAddress := strings.Split(sasAddressData, "0x")[1]
				//fmt.Println("sasAddress: ", sasAddress)

				// Convert SAS address to decimal format
				sasAddressDecimal := new(big.Int)
				sasAddressDecimal.SetString(sasAddress, 16)
				//fmt.Println("sasAddressDecimal: ", sasAddressDecimal)

				// Decrement by 1 SAS address value: WNN
				decimalOne := big.NewInt(1)
				osWnn := new(big.Int)
				osWnn.Sub(sasAddressDecimal, decimalOne)

				// Convert WNN to hex format
				osWnnHex := osWnn.Text(16)

				//fmt.Println("hex: ", osWnnHex)
				diskPath := "/dev/disk/by-id/wwn-0x" + osWnnHex
				//println("diskPath: ", diskPath)
				osDevice, err := os.Readlink(diskPath)
				if err != nil {
					// I have detected cases where SASAdress-1 doesnt exists under /dev/disk/by-id/wwn-0x
					// OS doesnt knows anything about these disks, maybe bogus hardware
					return "BogusDisk-OSUnknown", nil
				}
				osDevice = strings.ReplaceAll(osDevice, "../", "")
				//println("osDevice: ", osDevice)
				return osDevice, nil
			}
		}
		return "Unknown", nil
	case "sas2ircu":
		command := controllerId + " DISPLAY"
		outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "getJbodOsDevice", command)
		if err != nil {
			color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
		}
		if len(outputStderr.String()) != 0 {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		var enclosure string
		var diskSlot string
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if strings.Contains(line, "Enclosure #") {
				enclosureData := strings.Split(line, ":")
				enclosure = enclosureData[1]
				enclosure = utils.ClearString(enclosure)
			}
			if strings.Contains(line, "Slot #") {
				diskSlotData := strings.Split(line, ":")
				diskSlot = diskSlotData[1]
				diskSlot = strings.ReplaceAll(diskSlot, " ", "")
			}
			if enclosure == eid && diskSlot == slot && strings.Contains(line, "GUID") {
				guidData := strings.Split(line, ":")
				guid := guidData[1]
				osWnn := utils.ClearString(guid)
				diskPath := "/dev/disk/by-id/wwn-0x" + osWnn
				//println("diskPath: ", diskPath)
				osDevice, err := os.Readlink(diskPath)
				if err != nil {
					color.Red("++ ERROR: Readlink: %s", err)
					return "Unknown", err
				}
				osDevice = strings.ReplaceAll(osDevice, "../", "")
				//println("osDevice: ", osDevice)
				return osDevice, nil
			}
		}
		return "Unknown", nil
	case "adaptec":
		// I dont have this hardware configuration available for testing
		return "Unknown", nil
	default:
		return "Unknown", fmt.Errorf("Error: Unknown manufacturer")
	}
}

// Function as variable in order to be possible to mock it from unitary tests
var GetRaidOSDevice = func(manufacturer, controllerId, dg string) (string, error) {
	// fmt.Println("-- getRaidOSDevice --")
	// fmt.Println("manufacturer: ", manufacturer)
	// fmt.Println("controllerId: ", controllerId)
	// fmt.Println("dg: ", dg)

	switch manufacturer {
	case "mega", "perc":
		command := "/c" + controllerId + " /v" + dg + " show all"
		outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "getRaidOSDevice", command)
		if err != nil {
			color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
		}
		if len(outputStderr.String()) != 0 {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if strings.Contains(line, "SCSI NAA Id = ") {
				naaData := strings.Split(line, "=")
				//fmt.Println("naaData: ", naaData)
				naa := naaData[1]
				naa = utils.ClearString(naa)
				//fmt.Println("naa: ", naa)
				diskPath := "/dev/disk/by-id/wwn-0x" + naa
				//println("diskPath: ", diskPath)
				osDevice, err := os.Readlink(diskPath)
				if err != nil {
					color.Red("++ ERROR Readlink: %s", err)
					return "Unknown", err
				}
				osDevice = strings.ReplaceAll(osDevice, "../", "")
				//println("osDevice: ", osDevice)
				return osDevice, nil
			}
		}
		return "Unknown", nil
	case "sas2ircu":
		command := controllerId + " DISPLAY"
		outputStdout, outputStderr, err := utils.GetCommandOutput(manufacturer, "getRaidOSDevice", command)
		if err != nil {
			color.Red("++ ERROR: Something went wrong executing command %s: %v.", command, err)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command %s: %v.", command, err)
		}
		if len(outputStderr.String()) != 0 {
			color.Red("++ ERROR: Something went wrong executing command: %s.", command)
			return "Unknown", fmt.Errorf("Error: Something went wrong executing command: %s.", command)
		}
		//fmt.Println("out:", outputStdout.String(), "err:", outputStderr.String())

		scanner := bufio.NewScanner(strings.NewReader(outputStdout.String()))
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			//fmt.Println("line: ", line)
			if strings.Contains(line, "Volume wwid") {
				wwidData := strings.Split(line, ":")
				wwid := wwidData[1]
				wwid = utils.ClearString(wwid)
				//fmt.Println("wwid:", wwid)
				// Controller architecture must be LittleEndian/BigEndian
				// wwid bytes are reversed, we must unreverse it
				// Ex: 0392cdcec8b85290 -> 9052b8c8cecd9203
				var splittedWwid []string
				runes := []rune(wwid)
				//fmt.Println("runes:", runes)
				n := 2
				for i := 0; i < len(runes); i += n {
					end := i + n
					if end > len(runes) {
						end = len(runes)
					}
					splittedWwid = append(splittedWwid, string(runes[i:end]))
				}
				//fmt.Println("splittedWwid:", splittedWwid)

				for i, j := 0, len(splittedWwid)-1; i < j; i, j = i+1, j-1 {
					splittedWwid[i], splittedWwid[j] = splittedWwid[j], splittedWwid[i]
				}
				//fmt.Println("splittedWwid:", splittedWwid)

				reversedWwid := ""
				for _, part := range splittedWwid {
					reversedWwid += part
				}
				//fmt.Println("reversedWwid:", reversedWwid)

				// Misterious string: 600508e000000000, it seems to be the same for all SAS2IRCU controllers
				diskPath := "/dev/disk/by-id/wwn-0x600508e000000000" + reversedWwid
				//println("diskPath: ", diskPath)
				osDevice, err := os.Readlink(diskPath)
				if err != nil {
					color.Red("++ ERROR: Readlink: %s", err)
					return "Unknown", err
				}
				osDevice = strings.ReplaceAll(osDevice, "../", "")
				//println("osDevice: ", osDevice)
				return osDevice, nil
			}
		}
		return "Unknown", nil
	case "adaptec":
		diskPath := "/dev/disk/by-id/scsi-SAdaptec_" + dg
		//println("diskPath: ", diskPath)
		osDevice, err := os.Readlink(diskPath)
		if err != nil {
			//color.Red("++ ERROR: Readlink: %s", err)
			return "Unknown", err
		}
		osDevice = strings.ReplaceAll(osDevice, "../", "")
		//println("osDevice: ", osDevice)
		return osDevice, nil
	default:
		return "Unknown", fmt.Errorf("Error: Unknown manufacturer")
	}
}

// SoftRaid/ZFS/Btrfs/LVM over JBOD disks
func CheckJbodDisks(newRaids []utils.RaidStruct, noRaidDisks []utils.NoRaidDiskStruct) error {
	//fmt.Println("-- checkJbodDisks --")
	//fmt.Println(newRaids)
	// Check new disks against already saved noRaidDisks(JBOD), if any of them match, append string and set medium/model disk
	for _, newRaid := range newRaids {
		for i := range newRaid.Disks {
			newDisk := &newRaid.Disks[i]
			newDiskOsDevice := newDisk.OsDevice
			// Remove digits from newDiskOsDevice: SDA3 -> SDA as JBOD mode passtroughts the whole disk
			re := regexp.MustCompile(`\d`)
			newDiskOsDevice = re.ReplaceAllString(newDiskOsDevice, "")
			// noRaidDisks -> JBOD disks
			for i := range noRaidDisks {
				noRaidDisk := &noRaidDisks[i]
				//fmt.Printf("noRaidDisk.OsDevice: %v -- newDiskOsDevice: %v\n", noRaidDisk.OsDevice, newDiskOsDevice)
				if noRaidDisk.OsDevice == "JBOD-"+newDiskOsDevice {
					//fmt.Println("JBOD disk found")
					switch newDisk.ControllerId {
					case "softraid-0":
						noRaidDisk.OsDevice = noRaidDisk.OsDevice + " SoftRaid"
					case "zfs-0":
						noRaidDisk.OsDevice = noRaidDisk.OsDevice + " ZFS"
					case "btrfs-0":
						noRaidDisk.OsDevice = noRaidDisk.OsDevice + " Btrfs"
					case "lvm-0":
						noRaidDisk.OsDevice = noRaidDisk.OsDevice + " LVM"
					default:
						return nil
					}
					newDisk.Model = noRaidDisk.Model
					newDisk.Intf = noRaidDisk.Intf
					newDisk.Medium = noRaidDisk.Medium
					break
				}
				// We check values against all possible appends due to one JBOD disk can be used by various soft/zfs/btrfs raids, remember that disk partitions exist
				// When second raid is checked, string has already appended raid type, but we need to update disk medium and model too.
				if noRaidDisk.OsDevice == "JBOD-"+newDiskOsDevice+" SoftRaid" || noRaidDisk.OsDevice == "JBOD-"+newDiskOsDevice+" ZFS" || noRaidDisk.OsDevice == "JBOD-"+newDiskOsDevice+" Btrfs" || noRaidDisk.OsDevice == "JBOD-"+newDiskOsDevice+" LVM" {
					newDisk.Model = noRaidDisk.Model
					newDisk.Intf = noRaidDisk.Intf
					newDisk.Medium = noRaidDisk.Medium
				}
			}
		}
	}
	return nil
}

// ZFS/Btrfs/LVM over HardwareRaid disks
// Check if HW raid SDx(whole disk) is used by ZFS/Btrfs/LVM and append string if required
func CheckHardRaidDisks(newRaids []utils.RaidStruct, raids []utils.RaidStruct) error {
	//fmt.Println("-- checkHardRaidDisks --")
	// Check new disks against HardwareDisks, if any of them match, append string
	for _, newRaid := range newRaids {
		for i := range newRaid.Disks {
			newDisk := &newRaid.Disks[i]
			newDiskOsDevice := newDisk.OsDevice
			// Previous Raids
			for i := range raids {
				raid := &raids[i]
				//fmt.Printf("Checking newDiskOsDevice: %s -> raid.osDevice: %s\n", newDiskOsDevice, raid.osDevice)
				if newDiskOsDevice == raid.OsDevice {
					switch newDisk.ControllerId {
					// case "softraid-0": I dont have any MD softraid using whole Hw raid unit disk, so I cant test it
					case "zfs-0":
						raid.OsDevice = raid.OsDevice + " ZFS"
					case "btrfs-0":
						raid.OsDevice = raid.OsDevice + " Btrfs"
					case "lvm-0":
						raid.OsDevice = raid.OsDevice + " LVM"
					default:
						return nil
					}
					// Btrfs over MD device, check MD device info for model and medium info
					newDisk.Model = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
					newDisk.Intf = ""
					newDisk.Medium = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
					newDisk.SerialNumber = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
				}
				// We check values against all possible appends due to one hardDisk can be used by various zfs/btrfs raids, remember that disk partitions exist
				// When second raid is checked, string has already appended raid type, but we need to update disk medium and model too.
				if raid.OsDevice == newDiskOsDevice+" ZFS" || raid.OsDevice == newDiskOsDevice+" Btrfs" || raid.OsDevice == newDiskOsDevice+" LVM" {
					newDisk.Model = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
					newDisk.Intf = ""
					newDisk.Medium = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
					newDisk.SerialNumber = "Check " + strings.ToUpper(raid.OsDevice) + " disks."
				}
			}
		}
	}
	return nil
}
