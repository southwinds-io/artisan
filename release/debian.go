/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package release

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/magiconair/properties"
	"gopkg.in/yaml.v2"
	"southwinds.dev/artisan/build"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/registry"
	resx "southwinds.dev/os"
)

// BuildDebianPackage build a package containing debian packages
// If export options are specified, it also exports the package
func BuildDebianPackage(pkgNames []string, opts *ExportOptions) error {
	targetUri := opts.TargetUri
	creds := opts.TargetCreds
	artHome := opts.ArtHome

	if len(pkgNames) == 0 {
		return fmt.Errorf("debian package name is missing : %s", pkgNames)
	}

	pName, err := core.ParseName(aptPkgName())
	if err != nil {
		return fmt.Errorf("invalid artisan package name %s : %s, ", aptPkgName(), err)
	}

	// if a target has been specified
	if len(targetUri) > 0 {
		// if a final slash does not exist add it
		if targetUri[len(targetUri)-1] != '/' {
			targetUri = fmt.Sprintf("%s/", targetUri)
		}
		// automatically adds a tar filename to the URI based on the package name:tag
		targetUri = fmt.Sprintf("%s%s", targetUri, aptPkgTarFileName())
	} else {
		return fmt.Errorf("a destination URI must be specified to export the image")
	}

	// execution path
	tmp, err := core.NewTempDir(artHome)
	if err != nil {
		return fmt.Errorf("cannot create temp folder for processing image archive: %s", err)
	}
	core.DebugLogger.Printf("location of temporary folder %s", tmp)

	// create a target folder for the artisan package
	targetFolder := filepath.Join(tmp, "build")
	err = os.MkdirAll(targetFolder, 0755)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("failed to create build folder : %s", err)
	}
	// create a patch folder where packages to be used for patching will be downloaded
	patchFolder := filepath.Join(targetFolder, "patch")
	err = os.MkdirAll(patchFolder, 0755)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("failed to create patch folder : %s", err)
	}

	// get the debian packages either locally or from remote
	err = getPackages(pkgNames, patchFolder)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("failed to get package and its dependencies for packages %s \n error :- %s", pkgNames, err)
	}

	// get all the package name as string from folder "patch" which will be used to build
	// backup command
	bckupCmds, err := buildBackupCmds(patchFolder)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("failed to read .deb file from path %s, \n error is %s", patchFolder, err)
	}

	err = buildMetadata(patchFolder)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("failed to build metadata file for debian packages, \n error is %s", err)
	}

	err = generateShellScript(targetFolder)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("failed to generate shell script for debian patch application, \n error is %s", err)
	}

	ver, err := getOsInfo(patchFolder)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("failed to get host's operating system version, \n error is %s", err)
	}

	// generate build function build.yaml for this package
	bfBytes, err := generateBuildFunctions(bckupCmds, ver)
	if err != nil {
		return fmt.Errorf("failed to marshall debian package build file: %s", err)
	}

	// create a build file to build the package containing the debian packages tar
	pbfBytes, err := generateArtBuild(aptPkgName())
	if err != nil {
		return fmt.Errorf("failed to marshall debian packaging build file: %s", err)
	}

	// save package build and function build file
	core.InfoLogger.Println("packaging debian packages tarball file")
	err = os.WriteFile(filepath.Join(tmp, "build.yaml"), pbfBytes, 0755)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("cannot save debian packaging build file: %s", err)
	}
	err = os.WriteFile(filepath.Join(targetFolder, "build.yaml"), bfBytes, 0755)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("cannot save debian build function file: %s", err)
	}

	b := build.NewBuilder(artHome)
	b.SetBProc(opts.BuildProc)
	b.Build(tmp, "", "", pName, "", false, false, "", "", ".*", "")
	r := registry.NewLocalRegistry(artHome)
	if opts != nil {
		// export package
		core.InfoLogger.Printf("exporting debian package to tarball file")
		_, err = r.ExportPackage([]core.PackageName{*pName}, "", targetUri, creds)
		if err != nil {
			os.RemoveAll(tmp)
			return fmt.Errorf("cannot save debian package to destination: %s", err)
		}
		// append package name to the spec file
		spec := new(Spec)
		err = yaml.Unmarshal(opts.Specification.content, spec)
		if err != nil {
			os.RemoveAll(tmp)
			return fmt.Errorf("failed unmarshal spec file's content part: %s", err)
		}
		m := spec.Packages
		if m == nil {
			m = make(map[string]string)
			spec.Packages = m
		}
		m["PACKAGE_APT"] = pName.String()
		contents, err := core.ToYamlBytes(spec)
		if err != nil {
			return fmt.Errorf("failed to marshal the spec file's content part: %s", err)
		}
		opts.Specification.content = contents
	}
	return nil
}

func getPackages(pkgNames []string, target string) error {
	if len(pkgNames) == 0 {
		return fmt.Errorf("package names to be exported is empty")
	}

	// make sure pkgNames slice contains either all with debian package file name with .deb extension
	// or all will be debian package name (with no .deb extension)
	ext := filepath.Ext(pkgNames[0])
	prePkg := pkgNames[0]
	for _, p := range pkgNames {
		ex := filepath.Ext(p)
		if strings.Compare(strings.TrimSpace(ext), strings.TrimSpace(ex)) != 0 {
			return fmt.Errorf("all package names are not of same format, some has extension some don't [ %s ] [ %s ]", prePkg, p)
		}
		prePkg = p
	}

	// if the pkgNames slice contains deb package file names then copy them all from source folder
	// to target folder patch
	var err error
	if len(ext) > 0 {
		err = copyPackagesFromLocal(pkgNames, target)
	} else {
		err = downloadPackagesFromRemote(pkgNames, target)
	}

	return err
}

func downloadPackagesFromRemote(pkgNames []string, target string) error {

	// convert pkgNames slice to a single string, each element of slice separated by space
	allPkgs := strings.Join(pkgNames, " ")

	// get name of dependent packages
	compList, err := getDependencies(allPkgs, target)
	if err != nil {
		return fmt.Errorf("failed to query dependencies for packages %s \n error :- %s", allPkgs, err)
	}

	// download all the packages incudling dependencies
	err = downloadPackages(compList, target)
	if err != nil {
		return fmt.Errorf("failed to download package [ %s ] and its dependencies \n error :- %s", compList, err)
	}

	return nil
}

func copyPackagesFromLocal(pkgNames []string, target string) error {
	for _, p := range pkgNames {
		ext := filepath.Ext(p)
		if ext != ".deb" {
			return fmt.Errorf("package names extension is not .deb %s", p)
		}
		b, e := resx.ReadFile(p, "")
		if e != nil {
			return nil
		}
		core.DebugLogger.Printf("< byte size [ %d ] \n file path [ %s ]\n >", len(b), filepath.Join(target, filepath.Base(p)))
		e = resx.WriteFile(b, filepath.Join(target, filepath.Base(p)), "")
		if e != nil {
			return nil
		}
	}

	return nil
}

func aptPkgTarFileName() string {
	r := strings.NewReplacer(
		"/", "_",
		".", "_",
	)
	fileWithExt := fmt.Sprintf("%s.%s", r.Replace(aptPkgName()), "tar")
	return fileWithExt
}

func aptPkgName() string {
	return fmt.Sprintf("127.0.0.1/os/packages/apt:%d", time.Now().Unix())
}

func generateBuildFunctions(bckupCmds []string, targetOs string) ([]byte, error) {
	export_yes := true
	bf := data.BuildFile{
		Runtime: "ubi-min",
		Input: &data.Input{
			Var: data.Vars{
				&data.Var{
					Name:        "RELEASE_SNAPSHOT_DEVICE",
					Description: "name of snapshot device which will be used for taking snapshot",
					Required:    true,
					Type:        "string",
				},
			},
		},
		// Labels:  map[string]string{"target_os": targetOs},
		Functions: []*data.Function{
			{
				Name:        "apply",
				Description: "apply debian packages to the operating system. In case of failure, will automatically rollback",
				Export:      &export_yes,
				Run: []string{
					"bash -c './patch-apt.sh -v " + targetOs + "'",
				},
			},
			{
				Name:        "apply-snapshot",
				Description: "apply debian packages to the operating system with snapshot enabled. In case of failure, will automatically rollback",
				Export:      &export_yes,
				Run: []string{
					"bash -c './patch-apt.sh -s yes -v " + targetOs + "'",
				},
			},
			{
				Name:        "apply-snapshot-withdevice",
				Description: "apply debian packages to the operating system with snapshot enabled and snapshot device defined. In case of failure, will automatically rollback",
				Export:      &export_yes,
				Run: []string{
					"bash -c './patch-apt.sh -s yes -v " + targetOs + " -d ${RELEASE_SNAPSHOT_DEVICE}'",
				},
				Input: &data.InputBinding{
					Var: []string{
						"RELEASE_SNAPSHOT_DEVICE",
					},
				},
			},
		},
	}
	return yaml.Marshal(bf)
}

func buildBackupCmds(patchFolder string) ([]string, error) {
	files, err := os.ReadDir(patchFolder)
	if err != nil {
		return nil, err
	}
	core.DebugLogger.Printf("count of .deb files found at the path [ %s] is [ %d ] ", patchFolder, len(files))
	var bckupCmds []string
	core.InfoLogger.Printf("generating run command for .deb files")
	exp := regexp.MustCompile(`\r?\n`)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".deb" {
			// go to patch folder and get the debian package name
			cmd := fmt.Sprintf("bash -c 'cd %s && dpkg -f %s Package'", patchFolder, file.Name())
			pname, er := build.Exe(cmd, patchFolder, merge.NewEnVarFromSlice([]string{}), false)
			pname = exp.ReplaceAllString(pname, " ")
			if er != nil {
				return nil, er
			}
			bckupCmd := fmt.Sprintf("bash -c \"cd backup && sudo dpkg-query -W %s | awk '{print $1}' | sudo xargs dpkg-repack\"", pname)
			bckupCmds = append(bckupCmds, bckupCmd)
		}
	}

	return bckupCmds, nil
}

func downloadPackages(compList, patchFolder string) error {
	downloadPkg := "bash -c 'apt-get download " + strings.TrimSpace(compList) + "'"
	core.InfoLogger.Printf("downloading package %s and its dependencies", compList)
	core.DebugLogger.Printf("downloading debian package [%s] dependencies using command :\n %s", compList, downloadPkg)
	_, err := build.Exe(downloadPkg, patchFolder, merge.NewEnVarFromSlice([]string{}), false)

	return err
}

func generateArtBuild(pkgName string) ([]byte, error) {
	// create a build file to build the package containing the debian packages tar
	pbf := data.BuildFile{
		Labels: map[string]string{
			"package": pkgName,
		},
		Profiles: []*data.Profile{
			{
				Name:   "debian-package",
				Target: "./build",
				Type:   "content/apt",
			},
		},
	}
	return yaml.Marshal(pbf)
}

func getDependencies(pkgNames, executionPath string) (string, error) {

	// get all dependencies for current package
	pkgNames = strings.TrimSpace(pkgNames)
	qryDependencies := fmt.Sprintf("bash -c 'apt-rdepends %s | grep -v \"^ \" | sed 's/debconf-2.0/debconf/g'' | sed 's/time-daemon/systemd-timesyncd/g'", pkgNames)
	core.DebugLogger.Printf("querying debian package [%s] dependencies using command :\n %s\n ", pkgNames, qryDependencies)
	// execute the command synchronously
	core.InfoLogger.Printf("querying dependencies of package %s", pkgNames)
	// using async because when using sync occassionally it been observed that return list contains
	// status messages key word also like "processing" along with the dependent package names
	dep, err := build.ExeAsync(qryDependencies, executionPath, merge.NewEnVarFromSlice([]string{}), false)
	// note: the dep list will contain the parent package name also for which we looked dependencies
	if err != nil {
		return "", err
	}

	replacer := strings.NewReplacer("\n", " ", "  ", " ", "\t", "")
	dep = replacer.Replace(dep)
	return dedup(dep)
}

func dedup(dep string) (string, error) {
	core.DebugLogger.Printf("dedup input data is \n %s \n", dep)
	var b strings.Builder
	words := strings.Split(dep, " ")

	for _, word := range words {
		// find exact match for the word in the string builder
		w := strings.Replace(fmt.Sprintf("\b%s\b", word), "+", "\\+", -1)
		x, err := regexp.MatchString(w, b.String())
		if err != nil {
			return "", fmt.Errorf("failed dedup process while trying for word %s in new package list %s: %s", word, b.String(), err)
		}

		if x {
			continue
		}
		b.WriteString(word)
		b.WriteString(" ")
	}
	return strings.TrimSpace(b.String()), nil
}

func generateShellScript(targetFolder string) error {
	er := os.WriteFile(filepath.Join(targetFolder, "patch-apt.sh"), []byte(shellScript), 0755)
	return er
}

func getOsInfo(executionPath string) (string, error) {

	cmd := "cat /etc/os-release"
	core.DebugLogger.Printf("using command to find host operating system version :%s\n ", cmd)
	// using async because when using sync occassionally it been observed that return list contains
	// status messages key word also like "processing" along with the dependent package names
	output, err := build.ExeAsync(cmd, executionPath, merge.NewEnVarFromSlice([]string{}), false)
	// note: the dep list will contain the parent package name also for which we looked dependencies
	if err != nil {
		return "", err
	}

	p := properties.MustLoadString(output)
	id := p.MustGetString("ID")
	version_id := p.MustGetString("VERSION_ID")
	tagetOsVersion := id + "-" + version_id
	replacer := strings.NewReplacer("\"", "")
	tagetOsVersion = replacer.Replace(tagetOsVersion)
	return tagetOsVersion, err
}

func buildMetadata(patchFolder string) error {
	cmd := "bash -c 'dpkg-scanpackages . /dev/null > Release'"
	_, er := build.Exe(cmd, patchFolder, merge.NewEnVarFromSlice([]string{}), false)
	if er != nil {
		return er
	}

	cmd = "bash -c 'dpkg-scanpackages . /dev/null | gzip -9c > Packages.gz'"
	_, er = build.Exe(cmd, patchFolder, merge.NewEnVarFromSlice([]string{}), false)

	return er
}

const shellScript = `
#!/usr/bin/env bash
DEB_DIR=/tmp/patch
APT_DIR=/etc/apt
SNAP_DEVICE=""
STD=$(date +"%Y/%m/%d at %H:%M:%S")
SNAPS_ENABLED=""
SNAP_LV_NAME=tsbackup
SNAP_DEV_SIZE=6
SNAP_NO_TO_KEEP=1
OS_ID=$(grep "^ID=" /etc/os-release | awk -F"=" '{print $2}')
VERSION_ID=$(grep "VERSION_ID" /etc/os-release | awk -F"=" '{print $2}' | sed -e 's/^"//' -e 's/"$//')
OS_RELEASE="$OS_ID-$VERSION_ID"
TARGET_OS=""
#DISTRIB_ID=$(grep DISTRIB_ID /etc/lsb-release | awk -F"=" '{print $2}')

#################  Functions  ############################################

check_timeshift() {
    if [ ! $(command -v timeshift) ]; then
    echo "########################################################"
    echo ""
    echo "Timeshift not available!"
    echo ""
    echo "Timeshift package not installed or not in the PATH."
    echo ""
    echo "########################################################"
    exit 1
    fi
}

check_lvm() {
    if [ ! $(command -v lvs) ]; then
    echo "########################################################"
    echo ""
    echo "LVM not available!"
    echo ""
    echo "LVM package not installed or not in the PATH."
    echo ""
    echo "########################################################"
    fi
}

snapshot_number() {
## Check the numer of timeshift snapshots
    sudo timeshift --list --snapshot-device ${SNAP_DEVICE} |grep 'No snapshots on this device' > /dev/null
    if [ $? -eq 0 ]; then
        SNAP_NO=0
    else
        SNAP_NO=$(sudo timeshift --list  --snapshot-device ${SNAP_DEVICE}  | grep snapshots | awk '{print $1}')
    fi
}

oldes_snapshot() {
## Find snapshot name
    SNAP_NAME=$(sudo timeshift --list --snapshot-device ${SNAP_DEVICE} |grep Num -A2 | tail -n 1 |awk '{print $3}')
}

delete_snapshot() {
## Delete the snapshot
    echo "########################################################"
    echo ""
    echo "Deleting ${SNAP_NAME}"
    echo ""
    echo "########################################################"
    sudo timeshift --delete --snapshot ${SNAP_NAME} --snapshot-device ${SNAP_DEVICE}
}

delete_oldest_snapshot() {
## Delete the snapshot
    echo "########################################################"
    echo ""
    echo "Deleting ${SNAP_NAME}"
    echo ""
    echo "########################################################"
    sudo timeshift --delete --snapshot ${SNAP_NAME} --snapshot-device ${SNAP_DEVICE}
}

create_snapshot() {
    sudo timeshift --create --comments "before patching ${STD}" --snapshot-device ${SNAP_DEVICE}
    CREATE_SNAP_STATUS=$?
    sudo timeshift --list  --snapshot-device ${SNAP_DEVICE}  | grep "${STD}"
    CHECK_SNAP_STATUS=$?
    if [ $CREATE_SNAP_STATUS -eq 0 ] && [ $CHECK_SNAP_STATUS -eq 0 ]; then
        NEW_SNAP_NAME=$(sudo timeshift --list  --snapshot-device ${SNAP_DEVICE}  | grep "${STD}" | awk '{print $3}')
        echo "${SNAP_DEVICE};${NEW_SNAP_NAME}"| sudo tee $HOME/.timeshift_snap_info
        echo "########################################################"
        echo ""
        echo "Timeshift snapshot successfully created"
        echo ""
        echo "########################################################"
    else
        echo "########################################################"
        echo ""
        echo "Timeshift snapshot not created!"
        echo ""
        echo "########################################################"
        exit 1
    fi
}

apply_patch() {
## Copy the patch tirectory to ${DEB_DIR} to avoid permission issue

    cp -r patch ${DEB_DIR}

## Make sure Packages and Relese have correct access
    sudo chown -R _apt: ${DEB_DIR}
    sudo chmod a+r ${DEB_DIR}/Packages.gz
    sudo chmod a+r ${DEB_DIR}/Release

## Make a copy of the sources.list file

    sudo mv ${APT_DIR}/sources.list ${APT_DIR}/sources.list.bkp

## Create a new sources list file

    sudo echo "deb [trusted=yes] file:${DEB_DIR} ./" | sudo tee ${APT_DIR}/sources.list

## Patch the system

    sudo apt-get update \
    && sudo  DEBIAN_FRONTEND=noninteractive apt-get -qy \
    -o 'Dpkg::Options::=--force-confdef' -o 'Dpkg::Options::=--force-confold' upgrade -y

    if [ $? -eq 0 ]; then
        echo "########################################################"
        echo ""
        echo "Successfully patched"
        echo ""
        echo "########################################################"
    else
        echo "########################################################"
        echo ""
        echo "Patching failed! Check system logs"
        echo ""
        echo "########################################################"
        sudo mv ${APT_DIR}/sources.list.bkp ${APT_DIR}/sources.list
        exit 1
    fi
  
## Restore the sources.list file

    sudo mv ${APT_DIR}/sources.list.bkp ${APT_DIR}/sources.list

## Delete DEB_DIR

    sudo rm -rf ${DEB_DIR} 
}

usage() {
      echo "Usage: $0 [ -s yes ] [ -v os_version ] [ -d snapshot_device ]" 1>&2
    }
    
exit_abnormal() {
     usage
     exit 1
    }

fs_type() {
    FS_TYPE=$(sudo lsblk -f $SNAP_DEVICE -o FSTYPE | tail -n1)
}

fs_size() {
    if [[ "$FS_TYPE" =~ "ext" ]]; then
        B_SIZE=$(sudo tune2fs -l $SNAP_DEVICE | grep 'Block size' | awk -F ":" '{print $2}' | xargs)
        FREE_B_COUNT=$(sudo tune2fs -l $SNAP_DEVICE |  grep 'Free blocks' | awk -F ":" '{print $2}' | xargs)
        SNAP_DEV_FREE=$(echo "$B_SIZE*$FREE_B_COUNT/1073741824" | bc -l)   
    elif [[ "$FS_TYPE" =~ "xfs" ]]; then
        B_SIZE=$(sudo xfs_info $SNAP_DEVICE | grep ^data. | awk '{print $3}' | awk -F"=" '{print $2}' | sed -e 's/^,//' -e 's/,$//')
        FREE_B_COUNT=$(sudo xfs_db -r "-c freesp -s" $SNAP_DEVICE |grep "total free blocks" | awk '{ print $4 }')
        SNAP_DEV_FREE=$(echo "$B_SIZE*$FREE_B_COUNT/1073741824" | bc -l)
    else
        echo "########################################################"
        echo ""
        echo "Unsupported FS type. Use EXT or XFS"
        echo ""
        echo "########################################################"
        exit 1
    fi    
}

#################  Functions  ############################################

## Perform OS patching with or without timeshift snapshot
## To enable timeshift snapshoting use '-s yes' option 

while getopts ":s:v:d:" options; do
    
      case "${options}" in
        s)
          SNAPS_ENABLED=$(echo ${OPTARG} | tr '[:upper:]' '[:lower:]')
          if [ ${SNAPS_ENABLED} != 'yes' ]; then
             echo "Error: 'yes' required with -s"
             exit_abnormal
             exit 1
         
           fi
          ;;
        v)
          TARGET_OS=$(echo ${OPTARG} )
          if [ ${TARGET_OS} = '' ]; then
             echo "Error: 'os version' required with -v"
             exit_abnormal
             exit 1
         
           fi
          ;;
        d)
          SNAP_DEVICE=$(echo ${OPTARG} )
          if [ ${SNAP_DEVICE} = '' ]; then
             echo "Error: 'snapshot device required with -d"
             exit_abnormal
             exit 1
         
           fi
          ;;
        :)
          echo "Error: -${OPTARG} requires an argument."
          exit_abnormal
          ;;
        *)
          exit_abnormal
          ;;
      esac
    done

## Make sure this is a Debian like system

if [ ! -f /etc/debian_version ] && [ ! -f /etc/lsb-release ]; then
echo "########################################################"
echo ""
echo "OS distribution mismatch!"
echo ""
echo "The package should be executed on a Debian like system!"
echo ""
echo "########################################################"
exit 1
fi

## Make sure distribution version matches
 
if [ $TARGET_OS != $OS_RELEASE ]; then
echo "########################################################"
echo ""
echo "OS version mismatch!"
echo ""
echo "The package has been prepared for $TARGET_OS and is executed on $OS_RELEASE"
echo ""
echo "########################################################"
exit 1
fi

#################  Main script  ############################################

if [ ! -z ${SNAP_DEVICE} ] && [ -z ${SNAPS_ENABLED} ]; then
    echo "########################################################"
    echo ""
    echo "Option [ -d snapshot_device ] must be used with [ -s yes ]"
    echo ""
    echo "########################################################"
    exit 1
fi

if [ ${SNAPS_ENABLED} = 'yes' ] && [ ! -z ${SNAP_DEVICE} ]; then
    fs_type
    fs_size 
    if (( $(echo "$SNAP_DEV_FREE > $SNAP_DEV_SIZE" | bc -l) )); then
        ## Make sure timeshift has been installed on the system
        check_timeshift
        echo "########################################################"
        echo ""
        echo "Patching with timeshift snapshot"
        echo ""
        echo "########################################################"
        snapshot_number
        ## If SNAP_NO > SNAP_NO_TO_KEEP remove the oldes timeshift snapshot
        if [ ${SNAP_NO} -gt ${SNAP_NO_TO_KEEP} ]; then
            oldes_snapshot
            delete_oldest_snapshot
        fi
        ## Create snapshot with timeshift
        create_snapshot
        ## Patch the system
        apply_patch
    else
        echo "########################################################"
        echo ""
        echo "Not enough free disk space on $SNAP_DEVICE - at least "$SNAP_DEV_SIZE"G required"
        echo ""
        echo "Choose another device or add free space to the device"
        echo ""
        echo "########################################################"
        exit 1        
   fi

elif [ ${SNAPS_ENABLED} = 'yes' ] && [ -z ${SNAP_DEVICE} ]; then

    ## Make sure timeshift has been installed on the system

    check_timeshift

    echo "########################################################"
    echo ""
    echo "Patching with timeshift snapshot"
    echo ""
    echo "########################################################"

    ## Check if LVM is on the system
    command -v lvs > /dev/null
    if [ $? -ne 0 ] || [ $(sudo vgs --noheadings | wc -l) -eq 0 ]; then
        PWD_SIZE=$(df --output=avail -B 1 "$PWD" | tail -n 1 | numfmt --to="iec" | sed 's/[A-Za-z]*//g')
    ## Check if there is enough disk space on the working directory device
        if (( $(echo "$PWD_SIZE > $SNAP_DEV_SIZE" | bc -l) )); then 
            SNAP_DEVICE=$(df $PWD --output=source | tail -n 1)
            ## Check the numer of timeshift snapshots
            snapshot_number
            ## If SNAP_NO > SNAP_NO_TO_KEEP remove the oldes timeshift snapshot
            if [ ${SNAP_NO} -gt ${SNAP_NO_TO_KEEP} ]; then
                oldes_snapshot
                delete_oldest_snapshot
            fi
            ## Create snapshot with timeshift
            create_snapshot
            ## Patch the system
            apply_patch
                    
        else 
            SNAP_DEVICE=$(df $PWD --output=source | tail -n 1)
            echo "########################################################"
            echo ""
            echo "Not enough free disk space on $SNAP_DEVICE to create timeshift snapshot and LVM not installed."
            echo ""
            echo "Please create a filesystem for timeshift snapshots."
            echo ""
            echo "########################################################"
            exit 1
        ## Check if SNAP_FS exists on the system
        fi
    else
        sudo lvs |grep ${SNAP_LV_NAME} > /dev/null
        if [ $? -eq 0 ]; then
            SNAP_DEVICE=/dev/mapper/$(sudo lvs |grep ${SNAP_LV_NAME} |awk '{print $2}')-${SNAP_LV_NAME}     
            ## Check the numer of timeshift snapshots
            snapshot_number
            ## If SNAP_NO > SNAP_NO_TO_KEEP remove the oldes timeshift snapshot
            if [ ${SNAP_NO} -gt ${SNAP_NO_TO_KEEP} ]; then
                oldes_snapshot
                delete_oldest_snapshot
            fi
            ## Create snapshot with timeshift
            create_snapshot
            ## Patch the system
            apply_patch

        else
            ## Check if there is enough free space in PWD
            FREE_SPACE=$(df --output=avail -B 1 "$PWD" | tail -n 1 | numfmt --to="iec" | sed 's/[A-Za-z]*//g')
            if (( $(echo "$FREE_SPACE > $SNAP_DEV_SIZE" |bc -l) )); then
                SNAP_DEVICE=$(df $PWD --output=source | tail -n 1)
                ## Check the numer of timeshift snapshots
                snapshot_number
                ## If SNAP_NO > 2 remove the oldes timeshift snapshot
                if [ ${SNAP_NO} -gt ${SNAP_NO_TO_KEEP} ]; then
                    oldes_snapshot
                    delete_oldest_snapshot
                fi
                ## Create snapshot with timeshift
                create_snapshot
                ## Patch the system
                apply_patch
            else
                ## Check if there is enough free space in vg to create a volume
                VG_FREE_SIZE=$(sudo vgs --units G --no-suffix --noheadings -o vg_name,vg_free |sort -k 2 -r |head -n1 | awk '{print $2}')
                VG_NAME=$(sudo vgs --units G --no-suffix --noheadings -o vg_name,vg_free |sort -k 2 -r |head -n1 | awk '{print $1}')      
                if (( $(echo "$VG_FREE_SIZE > $SNAP_DEV_SIZE" |bc -l) )); then
                    sudo lvcreate -n ${SNAP_LV_NAME} -L ${SNAP_DEV_SIZE}G ${VG_NAME}
                    sudo mkfs.ext4 /dev/mapper/${VG_NAME}-${SNAP_LV_NAME}
                    SNAP_DEVICE=/dev/mapper/${VG_NAME}-${SNAP_LV_NAME}
                    ## Check the numer of timeshift snapshots
                    snapshot_number
                    ## If SNAP_NO > 2 remove the oldes timeshift snapshot
                    if [ ${SNAP_NO} -gt ${SNAP_NO_TO_KEEP} ]; then
                        oldes_snapshot
                        delete_oldest_snapshot
                    fi
                    ## Create snapshot with timeshift
                    create_snapshot
                    ## Patch the system
                    apply_patch

                else
                    echo "########################################################"
                    echo ""
                    echo "Not enough free space in VGs to create LV for timeshift snapshot"
                    echo ""
                    echo "At least $SNAP_DEV_SIZE GB required"
                    echo ""
                    echo "########################################################"
                    exit 1
                fi
            fi
        fi        
    fi   
else

echo "########################################################"
echo ""
echo "Patching without timeshift snapshot"
echo ""
echo "########################################################"

## Patch the system
apply_patch
fi
`
