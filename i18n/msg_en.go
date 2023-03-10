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

package i18n

// english
var msg_en = map[I18NKey]string{
	// error messages
	ERR_CANT_CREATE_REGISTRY_FOLDER: "cannot create local registry folder '%s', user home: '%s'",
	ERR_CANT_DOWNLOAD_LANG:          "cannot download language dictionary from '%s'",
	ERR_CANT_EXEC_FUNC_IN_PACKAGE:   "cannot execute function '%s' in package '%s'",
	ERR_CANT_LOAD_PRIV_KEY:          "cannot load the private key",
	ERR_CANT_PUSH_PACKAGE:           "cannot push package",
	ERR_CANT_READ_RESPONSE:          "cannot read response body",
	ERR_CANT_SAVE_FILE:              "cannot save file",
	ERR_CANT_UPDATE_LANG_FILE:       "cannot update language file",
	ERR_INSUFFICIENT_ARGS:           "insufficient arguments",
	ERR_INVALID_PACKAGE_NAME:        "invalid package name",
	ERR_TOO_MANY_ARGS:               "too many arguments",
	INFO_PUSHED:                     "pushed: %s\n",
	INFO_NOTHING_TO_PUSH:            "nothing to push\n",
	INFO_TAGGED:                     "tagged: %s\n",
	LBL_LS_HEADER:                   "REPOSITORY\t TAG\t PACKAGE ID\t PACKAGE TYPE\t CREATED\t SIZE\t",
	LBL_LS_HEADER_PLUS:              "REPOSITORY\t TAG\t PACKAGE ID\t PACKAGE TYPE\t CREATED\t SIZE\t AUTHOR\t",
}
