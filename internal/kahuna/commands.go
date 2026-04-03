/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
)

const (
	KahunaCfgFileDefault = "/etc/kowabunga/kahuna.yml"

	flagDescConfig  = "YAML config file to be used"
	flagDescDebug   = "Enable verbose/debug output"
	flagDescMigrate = "Perform any required database schema migration and gracefully exit afterwards"
)

func ParseCommands() (*os.File, bool, bool) {

	configFile := kingpin.Flag("config", flagDescConfig).Short('c').Default(KahunaCfgFileDefault).File()
	debug := kingpin.Flag("debug", flagDescDebug).Short('d').Bool()
	migrate := kingpin.Flag("migrate", flagDescMigrate).Short('m').Bool()

	kingpin.Parse()

	return *configFile, *debug, *migrate
}
