package main

// ToolList is a list of tools that are installed as binaries for development usage.
// This list gets installed to go bin directory once `mage init` is run.
// This is for binaries that need to be invoked as cli tools, not packages.
var ToolList = []string{}

// CIToolList is the required tooling to install from source with Mage.
// Keep this as small as possible to improve build performance.
// Use aqua when possible.
var CIToolList = []string{}
