// +build linux

package sys

/*
#cgo LDFLAGS: -ludev
#include <string.h>
#include <libudev.h>

struct udev *udev;

static struct udev *gosys_udev_init() {
	if (!udev)
		udev = udev_new();

	return udev;
}

static const char *gosys_udev_device_get_uuid(struct udev_device *dev) {
	if (!dev)
		return NULL;

	const char *data = NULL;
	const char *uuid = NULL;

	if (data = udev_device_get_property_value(dev, "ID_FS_UUID")) {
		uuid = strdup(data);
	}

	return uuid;
}

static const char *gosys_udev_get_uuid_by_name(char *name) {
	if (!gosys_udev_init())
		return NULL;

	struct udev_device *dev = udev_device_new_from_subsystem_sysname(udev, "block", name);

	if (!dev)
		return NULL;

	const char *uuid = gosys_udev_device_get_uuid(dev);

	udev_device_unref(dev);

	return uuid;
}

static const char *gosys_udev_get_uuid_by_devno(dev_t devno) {
	if (!gosys_udev_init())
		return NULL;

	struct udev_device *dev = udev_device_new_from_devnum(udev, 'b', devno);

	if (!dev)
		return NULL;

	const char *uuid = gosys_udev_device_get_uuid(dev);

	udev_device_unref(dev);

	return uuid;
}
*/
import "C"

import (
	"syscall"

	"github.com/google/uuid"
)

func deviceForPath(path string) (uint64, error) {
	stat := syscall.Stat_t{}

	if err := syscall.Lstat(path, &stat); err != nil {
		return 0, err
	}

	return stat.Dev, nil
}

func RootDiskUUID() (uuid.UUID, error) {
	devno, err := deviceForPath("/")

	if err != nil {
		return uuid.Nil, err
	}

	uuidStr := C.GoString(C.gosys_udev_get_uuid_by_devno(C.dev_t(devno)))

	return uuid.Parse(uuidStr)
}
