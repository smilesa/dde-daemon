// Code generated by "dbusutil-gen em -type Manager,WarnLevelConfigManager"; DO NOT EDIT.

package power

import (
	"pkg.deepin.io/lib/dbusutil"
)

func (v *Manager) GetExportedMethods() dbusutil.ExportedMethods {
	return dbusutil.ExportedMethods{
		{
			Name: "Reset",
			Fn:   v.Reset,
		},
		{
			Name:   "SetPrepareSuspend",
			Fn:     v.SetPrepareSuspend,
			InArgs: []string{"suspendState"},
		},
	}
}
func (v *WarnLevelConfigManager) GetExportedMethods() dbusutil.ExportedMethods {
	return dbusutil.ExportedMethods{
		{
			Name: "Reset",
			Fn:   v.Reset,
		},
	}
}