/*
 * Copyright (C) 2014 ~ 2018 Deepin Technology Co., Ltd.
 *
 * Author:     jouyouyun <jouyouwen717@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package bluetooth

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/godbus/dbus"
)

const (
	deviceStateDisconnected = 0
	// device state is connecting or disconnecting, mark them as device state doing
	deviceStateConnecting    = 1
	deviceStateConnected     = 2
	deviceStateDisconnecting = 3
)

type deviceState uint32

func (s deviceState) String() string {
	switch s {
	case deviceStateDisconnected:
		return "Disconnected"
	case deviceStateConnecting:
		return "Connecting"
	case deviceStateConnected:
		return "Connected"
	case deviceStateDisconnecting:
		return "Disconnecting"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

type DeviceInfo struct {
	Path        dbus.ObjectPath
	AdapterPath dbus.ObjectPath

	Alias            string
	Trusted          bool
	Paired           bool
	State            deviceState
	ServicesResolved bool
	ConnectState     bool

	UUIDs   []string
	Name    string
	Icon    string
	RSSI    int16
	Address string
}

func unmarshalDeviceInfo(data string) (*DeviceInfo, error) {
	var device DeviceInfo
	err := json.Unmarshal([]byte(data), &device)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

type DeviceInfoMap struct {
	mu    sync.Mutex
	infos map[dbus.ObjectPath]DeviceInfos
}

type DeviceInfos []DeviceInfo

func (infos DeviceInfos) getDevice(path dbus.ObjectPath) (int, *DeviceInfo) {
	for idx, info := range infos {
		if info.Path == path {
			return idx, &info
		}
	}
	return -1, nil
}

func (infos DeviceInfos) removeDevice(path dbus.ObjectPath) (DeviceInfos, bool) {
	idx, _ := infos.getDevice(path)
	if idx == -1 {
		return infos, false
	}
	return append(infos[:idx], infos[idx+1:]...), true
}

func (m *DeviceInfoMap) getDeviceNoLock(adapterPath dbus.ObjectPath,
	devPath dbus.ObjectPath) (int, *DeviceInfo) {
	devices := m.infos[adapterPath]
	return devices.getDevice(devPath)
}

func (m *DeviceInfoMap) getDevice(adapterPath dbus.ObjectPath, devPath dbus.ObjectPath) (int, *DeviceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getDeviceNoLock(adapterPath, devPath)
}

func (m *DeviceInfoMap) getDevices(adapterPath dbus.ObjectPath) DeviceInfos {
	m.mu.Lock()
	defer m.mu.Unlock()
	devices := m.infos[adapterPath]
	devicesCopy := make(DeviceInfos, len(devices))
	copy(devicesCopy, devices)
	return devicesCopy
}

func (m *DeviceInfoMap) addOrUpdateDevice(devInfo *DeviceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	devices := m.infos[devInfo.AdapterPath]
	idx, _ := devices.getDevice(devInfo.Path)
	if idx != -1 {
		// 更新
		devices[idx] = *devInfo
		return
	}
	m.infos[devInfo.AdapterPath] = append(devices, *devInfo)
}

func (m *DeviceInfoMap) removeDevice(adapterPath, devPath dbus.ObjectPath) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	devices := m.infos[adapterPath]
	newDevices, ok := devices.removeDevice(devPath)
	if ok {
		m.infos[adapterPath] = newDevices
	}
	return ok
}

func (m *DeviceInfoMap) clear() {
	m.mu.Lock()
	m.infos = make(map[dbus.ObjectPath]DeviceInfos)
	m.mu.Unlock()
}

func (m *DeviceInfoMap) findFirst(fn func(devInfo *DeviceInfo) bool) *DeviceInfo {
	if fn == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, infos := range m.infos {
		for _, info := range infos {
			// #nosec G601
			if fn(&info) {
				return &info
			}
		}
	}
	return nil
}

func (m *DeviceInfoMap) getDeviceWithPath(devPath dbus.ObjectPath) *DeviceInfo {
	return m.findFirst(func(devInfo *DeviceInfo) bool {
		return devInfo.Path == devPath
	})
}
