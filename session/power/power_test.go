/*
 * Copyright (C) 2016 ~ 2018 Deepin Technology Co., Ltd.
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

package power

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarnLevelConfig(t *testing.T) {
	conf := &warnLevelConfig{
		UsePercentageForPolicy: true,

		LowTime:      1200,
		DangerTime:   900,
		CriticalTime: 600,
		ActionTime:   300,

		LowPowerNotifyThreshold: 20,
		remindPercentage:        25,
		LowPercentage:           20,
		DangerPercentage:        15,
		CriticalPercentage:      10,
		ActionPercentage:        5,
	}
	assert.True(t, conf.isValid())
	conf.LowTime = 599
	assert.False(t, conf.isValid())

	conf.LowTime = 1200
	conf.LowPercentage = 9
	assert.False(t, conf.isValid())
}

func Test_getWarnLevel(t *testing.T) {
	config := &warnLevelConfig{
		UsePercentageForPolicy: true,

		LowTime:      1200,
		DangerTime:   900,
		CriticalTime: 600,
		ActionTime:   300,

		LowPowerNotifyThreshold: 20,
		remindPercentage:        25,
		LowPercentage:           20,
		DangerPercentage:        15,
		CriticalPercentage:      10,
		ActionPercentage:        5,
	}

	onBattery := false
	assert.Equal(t, getWarnLevel(config, onBattery, 1.0, 0), WarnLevelNone)

	onBattery = true
	config.UsePercentageForPolicy = true

	assert.Equal(t, getWarnLevel(config, onBattery, 0.0, 0), WarnLevelNone)
	assert.Equal(t, getWarnLevel(config, onBattery, 1.1, 0), WarnLevelAction)
	assert.Equal(t, getWarnLevel(config, onBattery, 5.0, 0), WarnLevelAction)
	assert.Equal(t, getWarnLevel(config, onBattery, 5.1, 0), WarnLevelCritical)
	assert.Equal(t, getWarnLevel(config, onBattery, 10.0, 0), WarnLevelCritical)
	assert.Equal(t, getWarnLevel(config, onBattery, 10.1, 0), WarnLevelDanger)
	assert.Equal(t, getWarnLevel(config, onBattery, 15.0, 0), WarnLevelDanger)
	assert.Equal(t, getWarnLevel(config, onBattery, 15.1, 0), WarnLevelLow)
	assert.Equal(t, getWarnLevel(config, onBattery, 20.0, 0), WarnLevelLow)
	assert.Equal(t, getWarnLevel(config, onBattery, 20.1, 0), WarnLevelNone)
	assert.Equal(t, getWarnLevel(config, onBattery, 50.0, 0), WarnLevelNone)

	config.UsePercentageForPolicy = false
	// use time to empty
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 0), WarnLevelNone)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 61), WarnLevelAction)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 300), WarnLevelAction)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 301), WarnLevelCritical)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 600), WarnLevelCritical)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 601), WarnLevelDanger)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 900), WarnLevelDanger)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 901), WarnLevelLow)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 1200), WarnLevelLow)
	assert.Equal(t, getWarnLevel(config, onBattery, 0, 12001), WarnLevelNone)
}

func TestMetaTasksMin(t *testing.T) {
	tasks := metaTasks{
		metaTask{
			name:  "n1",
			delay: 10,
		},
		metaTask{
			name:  "n2",
			delay: 30,
		},
		metaTask{
			name:  "n3",
			delay: 20,
		},
	}
	assert.Equal(t, tasks.min(), int32(10))

	tasks = metaTasks{}
	assert.Equal(t, tasks.min(), int32(0))

	tasks = metaTasks{
		metaTask{
			name:  "n1",
			delay: 10,
		},
	}
	assert.Equal(t, tasks.min(), int32(10))
}
