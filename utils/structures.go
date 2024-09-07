package utils

// Controller struct
type ControllerStruct struct {
	Id           string
	Manufacturer string
	Model        string
	Status       string
}

// ZFS pool struct
type PoolStruct struct {
	ControllerId string
	Name         string
	State        string
	Size         string
	OsDevice     string
}

// LVM volumeGroup struct
type VolumeGroupStruct struct {
	ControllerId string
	Name         string
	State        string
	Size         string
}

// Disk struct, when its a hardware raid disk, osDevice will be empty, as osDevice is controller delivered virtual drive
type DiskStruct struct {
	ControllerId string
	Dg           string
	EidSlot      string
	State        string
	Size         string
	Intf         string
	Medium       string
	Model        string
	SerialNumber string
	OsDevice     string
}

// Raid struct, all storcli parsed data as string
type RaidStruct struct {
	ControllerId string
	RaidLevel    int
	Dg           string
	RaidType     string
	State        string
	Size         string
	Disks        []DiskStruct
	OsDevice     string
}

// Every raidStruct object will be binded to AddDisk function
// Inside AddDisk function raid object will be called r
func (r *RaidStruct) AddDisk(disk DiskStruct) {
	r.Disks = append(r.Disks, disk)
}

// Disk struct, all storcli parsed data as string
type NoRaidDiskStruct struct {
	ControllerId string
	EidSlot      string
	State        string
	Size         string
	Intf         string
	Medium       string
	Model        string
	SerialNumber string
	OsDevice     string
}
