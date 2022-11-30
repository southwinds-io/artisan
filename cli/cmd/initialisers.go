/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cmd

func InitialiseRootCmd(artHome string) *RootCmd {
	rootCmd := NewRootCmd()
	utilCmd := InitialiseUtilCommand(artHome)
	specCmd := InitialiseSpecCommand(artHome)
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
	manifCmd := NewManifestCmd(artHome)
	exeCmd := NewExeCmd(artHome)
	exeCCmd := NewExeCCmd(artHome)
	envCmd := InitialiseEnvCommand(artHome)
	pruneCmd := NewPruneCmd(artHome)
	rootCmd.Cmd.AddCommand(
		utilCmd.Cmd,
		specCmd.Cmd,
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

func InitialiseUtilCommand(artHome string) *UtilCmd {
	utilCmd := NewUtilCmd()
	utilPwdCmd := NewUtilPwdCmd()
	utilNameCmd := NewUtilNameCmd()
	utilExtractCmd := NewUtilExtractCmd()
	utilB64Cmd := NewUtilBase64Cmd()
	utilStampCmd := NewUtilStampCmd()
	utilCurlCmd := NewUtilCurlCmd()
	waitCmd := NewWaitCmd()
	langCmd := InitialiseLangCommand(artHome)
	gitSyncCmd := NewGitSyncCmd()
	serveCmd := NewServeCmd()
	replaceCmd := NewUtilReplaceCmd()
	utilCmd.Cmd.AddCommand(
		utilPwdCmd.Cmd,
		utilExtractCmd.Cmd,
		utilNameCmd.Cmd,
		utilB64Cmd.Cmd,
		utilStampCmd.Cmd,
		waitCmd.Cmd,
		utilCurlCmd.Cmd,
		langCmd.Cmd,
		gitSyncCmd.Cmd,
		serveCmd.Cmd,
		replaceCmd.Cmd,
	)
	return utilCmd
}

func InitialiseSpecCommand(artHome string) *SpecCmd {
	specCmd := NewSpecCmd()
	specExportCmd := NewSpecExportCmd(artHome)
	specImportCmd := NewSpecImportCmd(artHome)
	specDownCmd := NewSpecDownCmd()
	specUpCmd := NewSpecUpCmd()
	specPushCmd := NewSpecPushCmd(artHome)
	specInfoCmd := NewSpecInfoCmd()
	specPullCmd := NewSpecPullCmd(artHome)
	specCmd.Cmd.AddCommand(specExportCmd.Cmd)
	specCmd.Cmd.AddCommand(specImportCmd.Cmd)
	specCmd.Cmd.AddCommand(specDownCmd.Cmd)
	specCmd.Cmd.AddCommand(specUpCmd.Cmd)
	specCmd.Cmd.AddCommand(specPushCmd.Cmd)
	specCmd.Cmd.AddCommand(specInfoCmd.Cmd)
	specCmd.Cmd.AddCommand(specPullCmd.Cmd)
	return specCmd
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
