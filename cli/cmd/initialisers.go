/*
   Artisan Core - Automation Manager
   Copyright (C) 2022-Present SouthWinds Tech Ltd - www.southwinds.io

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package cmd

func InitialiseRootCmd(artHome string) *RootCmd {
	rootCmd := NewRootCmd()
	utilCmd := InitialiseUtilCommand(artHome)
	buildCmd := NewBuildCmd(artHome)
	lsCmd := NewListCmd(artHome)
	pushCmd := NewPushCmd(artHome)
	rmCmd := NewRmCmd(artHome)
	tagCmd := NewTagCmd(artHome)
	runCmd := NewRunCmd(artHome)
	runCCmd := NewRunCCmd(artHome)
	runACmd := NewRunACmd(artHome)
	mergeCmd := NewMergeCmd()
	pullCmd := NewPullCmd(artHome)
	openCmd := NewOpenCmd(artHome)
	flowCmd := InitialiseFlowCommand(artHome)
	manifCmd := InitialiseManifestCmd(artHome)
	exeCmd := NewExeCmd(artHome)
	exeCCmd := NewExeCCmd(artHome)
	envCmd := InitialiseEnvCommand(artHome)
	pruneCmd := NewPruneCmd(artHome)
	rootCmd.Cmd.AddCommand(
		utilCmd.Cmd,
		buildCmd.Cmd,
		lsCmd.Cmd,
		pushCmd.Cmd,
		rmCmd.Cmd,
		tagCmd.Cmd,
		runCmd.Cmd,
		runCCmd.Cmd,
		runACmd.Cmd,
		mergeCmd.Cmd,
		pullCmd.Cmd,
		openCmd.cmd,
		flowCmd.Cmd,
		manifCmd.Cmd,
		exeCmd.cmd,
		exeCCmd.Cmd,
		envCmd.Cmd,
		pruneCmd.Cmd,
	)
	return rootCmd
}

func InitialiseManifestCmd(home string) *ManifestCmd {
	mCmd := NewManifestCmd()
	mGetCmd := NewManifestGetCmd(home)
	mFxCmd := NewManifestFxCmd(home)
	mCmd.Cmd.AddCommand(mGetCmd.Cmd, mFxCmd.Cmd)
	return mCmd
}

func InitialiseUtilCommand(artHome string) *UtilCmd {
	utilCmd := NewUtilCmd()
	utilPwdCmd := NewUtilPwdCmd()
	utilNameCmd := NewUtilNameCmd()
	utilExtractCmd := NewUtilExtractCmd()
	utilB64Cmd := NewUtilBase64Cmd()
	utilStampCmd := NewUtilStampCmd()
	utilCurlCmd := NewUtilCurlCmd()
	utilWaitCmd := NewUtilWaitCmd()
	utilSysInfoCmd := NewUtilSysInfoCmd()
	langCmd := InitialiseLangCommand(artHome)
	serveCmd := NewServeCmd()
	replaceCmd := NewUtilReplaceCmd()
	upconfCmd := NewUpConfCmd()
	utilCmd.Cmd.AddCommand(
		utilPwdCmd.Cmd,
		utilExtractCmd.Cmd,
		utilNameCmd.Cmd,
		utilB64Cmd.Cmd,
		utilStampCmd.Cmd,
		utilWaitCmd.Cmd,
		utilSysInfoCmd.Cmd,
		utilCurlCmd.Cmd,
		langCmd.Cmd,
		serveCmd.Cmd,
		replaceCmd.Cmd,
		upconfCmd.Cmd,
	)
	return utilCmd
}

func InitialiseEnvCommand(artHome string) *EnvCmd {
	envCmd := NewEnvCmd()
	envPackageCmd := NewEnvPackageCmd(artHome)
	envFlowCmd := NewEnvFlowCmd()
	envCmd.Cmd.AddCommand(envFlowCmd.Cmd, envPackageCmd.Cmd)
	return envCmd
}

func InitialiseLangCommand(artHome string) *LangCmd {
	langCmd := NewLangCmd()
	langFetchCmd := NewLangFetchCmd(artHome)
	langUpdateCmd := NewLangUpdateCmd(artHome)
	langCmd.Cmd.AddCommand(langFetchCmd.Cmd, langUpdateCmd.Cmd)
	return langCmd
}

func InitialiseFlowCommand(artHome string) *FlowCmd {
	flowCmd := NewFlowCmd()
	flowMergeCmd := NewFlowMergeCmd(artHome)
	flowRunCmd := NewFlowRunCmd(artHome)
	flowCmd.Cmd.AddCommand(flowMergeCmd.Cmd, flowRunCmd.Cmd)
	return flowCmd
}
