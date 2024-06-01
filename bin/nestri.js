#!/usr/bin/env node

/**
 * Main CLI entrypoint to use the nestri libraries
 * This file is meant to be linked as a "nestri" executable.
 *
 * @name nestri
 */

'use strict';

// Define our basic CLI
var yargs = require('yargs');
var argv = yargs
  .commandDir('../cmds')
  .demand(1)
  .help()
  .global('verbose')
  .count('verbose')
  .alias('v', 'verbose')
  .argv;

// Appease code styles
argv = argv;