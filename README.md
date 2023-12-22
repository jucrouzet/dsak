# dsak : Dave's Swiss Army Knife.

Dave's a developer.
Dave needs tools to work.
Dave loves cli.
This is Dave's Swiss Army Knife.
All the tools he needs in one cli.

[![Coverage Status](https://coveralls.io/repos/github/jucrouzet/dsak/badge.svg)](https://coveralls.io/github/jucrouzet/dsak)


## Usage:
  dsak [flags]
  dsak [command]

Available Commands:
  base        Transforms an integer from a base to another
  base64      Base64 tools
  completion  Generate the autocompletion script for the specified shell
  config      Get or set a configuration value
  dns         DNS Tools
  help        Help about any command
  http        HTTP Tools
  timestamp   Timestamp tools

Flags:
  -h, --help            help for dsak
      --jsonlogs        Log output in JSON format
      --no-color        Diable color in output
      --output string   Command output resource (default "stdout")
      --timeout uint    Timeout for command in milliseconds, 0 for unlimited (default 10000)
      --verbose         Run command verbosely

Use "dsak [command] --help" for more information about a command.
