/*
 * Copyright (C) 2020 ~ 2022 Deepin Technology Co., Ltd.
 *
 * Author:     lichangze <ut001335@uniontech.com>
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
package display

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isInVM(t *testing.T) {
	err := exec.Command("systemd-detect-virt", "-v", "-q").Run()
	if err != nil {
		assert.False(t, isInVM())
	} else {
		assert.True(t, isInVM())
	}
}

func Test_binExist(t *testing.T) {
	const cmdPath = "/usr/bin/apt"
	_, err := os.Stat(cmdPath)
	if err != nil {
		if os.IsNotExist(err) {
			assert.False(t, binExist(cmdPath))
		}
	} else {
		assert.True(t, binExist(cmdPath))
	}
}

func Test_loadRendererConfig(t *testing.T) {
	const cfgPath = "./test_data"

	var cfg RendererConfig
	cfg.BlackList = []string{
		"llvmpipe",
	}
	err := genRendererConfig(&cfg, cfgPath)
	assert.Nil(t, err)
	defer func() {
		_ = os.RemoveAll(cfgPath)
	}()
	c, err := loadRendererConfig(cfgPath)
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func Test_genRendererConfig(t *testing.T) {
	const cfgPath = "./test_data"

	var cfg RendererConfig
	cfg.BlackList = []string{
		"llvmpipe",
	}
	err := genRendererConfig(&cfg, cfgPath)
	assert.Nil(t, err)
	_ = os.RemoveAll(cfgPath)
}
