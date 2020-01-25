package main

// #cgo LDFLAGS: -L. -lcryptsetup
// #include <unistd.h>
// #include <sys/reboot.h>
// #include <stdlib.h>
// #include <libcryptsetup.h>
import "C"

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"unsafe"
)

type blockDevice struct {
	Name     string        `json:"name"`
	Children []blockDevice `json:"children"`
	FSType   string        `json:"fstype"`
	Parent   *blockDevice
}

type blockList struct {
	Blockdevices []blockDevice `json:"blockdevices"`
}

func listCryptDevices() ([]string, error) {
	out, err := exec.Command("lsblk", "--fs", "--json").Output()
	if err != nil {
		fmt.Println("errored")
		return nil, err
	}

	var list blockList
	err = json.Unmarshal(out, &list)
	if err != nil {
		return nil, err
	}

	devices := []*blockDevice{}
	for i := 0; i < len(list.Blockdevices); i++ {
		device := &list.Blockdevices[i]
		devices = append(devices, device)
	}

	cryptDevices := []string{}

	for i := 0; i < len(devices); i++ {
		device := devices[i]
		// fmt.Println(device.Name, device.FSType)
		if device.FSType == "crypto_LUKS" {
			cryptDevices = append(cryptDevices, device.Name)
		}

		for j := 0; j < len(device.Children); j++ {
			child := &device.Children[j]
			child.Parent = device
			devices = append(devices, child)
		}
	}

	return cryptDevices, nil
}

type Luks struct {
	cd 			*C.struct_crypt_device
	luksType	string
}

func LoadLuks(device string, luksType string) (*Luks, error) {
	dev := C.CString(fmt.Sprintf("/dev/%s", device))
	defer C.free(unsafe.Pointer(dev))
	var cd *C.struct_crypt_device
	r := C.crypt_init(&cd, dev)
	if r < 0 {
		return nil, fmt.Errorf("Could not crypt_init. Are you root?")
	}
	luks := &Luks{
		cd,
		luksType,
	}	
	_type := C.CString(luksType)
	defer C.free(unsafe.Pointer(_type))
	r = C.crypt_load(cd, _type, nil)
	if r < 0 {
		return nil, fmt.Errorf("Could not crypt_load. Are you root?")
	}

	fmt.Println(device)
	fmt.Printf("\tcipher used: %s\n", C.GoString(C.crypt_get_cipher(cd)))
	fmt.Printf("\tcipher mode: %s\n", C.GoString(C.crypt_get_cipher_mode(cd)))
	fmt.Printf("\tdevice UUID: %s\n", C.GoString(C.crypt_get_uuid(cd)))

	return luks, nil
}

func (luks *Luks) Close() {
	C.crypt_free(luks.cd)
}

func (luks *Luks) ShowKeyslots()  {
	_type := C.CString(luks.luksType)
	defer C.free(unsafe.Pointer(_type))
	numKeyslots := C.crypt_keyslot_max(_type)
	for k := 0; k < int(numKeyslots); k++ {
		info := C.crypt_keyslot_status(luks.cd, C.int(k))

		switch info {
		case C.CRYPT_SLOT_INVALID:
			fmt.Println(k, "CRYPT_SLOT_INVALID")
			break

		case C.CRYPT_SLOT_INACTIVE:
			fmt.Println(k, "CRYPT_SLOT_INACTIVE")
			break

		case C.CRYPT_SLOT_ACTIVE:
			fmt.Println(k, "CRYPT_SLOT_ACTIVE")
			break

		case C.CRYPT_SLOT_ACTIVE_LAST:
			fmt.Println(k, "CRYPT_SLOT_ACTIVE_LAST")
			break

			// Only in LUKS2, which doesn't seem to be in the package I have.
			// case C.CRYPT_SLOT_UNBOUND:
			// 	fmt.Println(k, "CRYPT_SLOT_UNBOUND")
			// 	break
		}
	}
}

func (luks *Luks) WipeKeyslots() {
	_type := C.CString(luks.luksType)
	defer C.free(unsafe.Pointer(_type))
	numKeyslots := C.crypt_keyslot_max(_type)
	for k := 0; k < int(numKeyslots); k++ {		
		info := C.crypt_keyslot_status(luks.cd, C.int(k))
		if info == C.CRYPT_SLOT_ACTIVE || info == C.CRYPT_SLOT_ACTIVE_LAST {
			r := C.crypt_keyslot_destroy(luks.cd, C.int(k))
			if r == 0 {
				fmt.Printf("Wiped keyslot %w\n", k)
			} else {
				fmt.Printf("Failed to wipe keyslot %w\n", k)
			}
		}
	}
}

func Shutdown() {
	C.reboot(C.RB_POWER_OFF)
}


// cryptTypes := []string{
// 	C.CRYPT_PLAIN,
// 	C.CRYPT_LUKS1,
// 	C.CRYPT_LUKS2,
// 	C.CRYPT_PLAIN,
// 	C.CRYPT_LOOPAES,
// 	C.CRYPT_VERITY,
// 	C.CRYPT_TCRYPT,
// 	C.CRYPT_INTEGRITY}


func main() {
	devices, err := listCryptDevices()
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		luks, err := LoadLuks(device, C.CRYPT_LUKS1)		
		if err != nil {
			log.Fatal(err)
		}
		luks.ShowKeyslots()
		fmt.Println()		
		luks.WipeKeyslots()
		fmt.Println()
		luks.ShowKeyslots()
	}
	// Shutdown()
}
