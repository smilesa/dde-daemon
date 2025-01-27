/*
 * Copyright (C) 2017 ~ 2018 Deepin Technology Co., Ltd.
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

package gesture

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/godbus/dbus"
	login1 "github.com/linuxdeepin/go-dbus-factory/org.freedesktop.login1"
	x "github.com/linuxdeepin/go-x11-client"
	"github.com/linuxdeepin/go-x11-client/util/keybind"
	"github.com/linuxdeepin/go-x11-client/util/wm/ewmh"
)

var (
	xconn  *x.Conn
	_dconn *dbus.Conn
	_self  login1.Session
)

const (
	positionTop int32 = iota
	positionRight
	positionBottom
	positionLeft
)

type Rect struct {
	X, Y          int32
	Width, Height uint32
}

func isKbdAlreadyGrabbed() bool {
	if getX11Conn() == nil {
		return false
	}

	var grabWin x.Window

	// 如果是防止安全问题只抓取rootWin就可以了，抓取激活窗口会导致多任务等窗口响应失效。
	rootWin := xconn.GetDefaultScreen().Root
	grabWin = rootWin

	err := keybind.GrabKeyboard(xconn, grabWin)
	if err == nil {
		// grab keyboard successful
		_ = keybind.UngrabKeyboard(xconn)
		return false
	}

	logger.Warningf("GrabKeyboard win %d failed: %v", grabWin, err)

	gkErr, ok := err.(keybind.GrabKeyboardError)
	if ok && gkErr.Status == x.GrabStatusAlreadyGrabbed {
		return true
	}
	return false
}

func getCurrentActionWindowCmd() string {
	win, err := ewmh.GetActiveWindow(xconn).Reply(xconn)
	if err != nil {
		logger.Warning("Failed to get current active window:", err)
		return ""
	}
	pid, err := ewmh.GetWMPid(xconn, win).Reply(xconn)
	if err != nil {
		logger.Warning("Failed to get current window pid:", err)
		return ""
	}
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		logger.Warning("Failed to read cmdline:", err)
		return ""
	}
	return string(data)
}

func isSessionActive(sessionPath dbus.ObjectPath) bool {
	if _dconn == nil {
		conn, err := dbus.SystemBus()
		if err != nil {
			logger.Error("Failed to new system bus:", err)
			return false
		}
		_dconn = conn
	}

	if _self == nil {
		self, err := login1.NewSession(_dconn, sessionPath)
		if err != nil {
			logger.Error("Failed to connect self session:", err)
			return false
		}
		_self = self
	}

	active, err := _self.Active().Get(dbus.FlagNoAutoStart)
	if err != nil {
		logger.Error("Failed to get self active:", err)
		return false
	}
	return active
}

func getX11Conn() *x.Conn {
	if xconn == nil {
		conn, err := x.NewConn()
		if err != nil {
			return nil
		}
		xconn = conn
	}
	return xconn
}

func isInWindowBlacklist(cmd string, list []string) bool {
	for _, v := range list {
		if strings.Contains(cmd, v) {
			return true
		}
	}
	return false
}
