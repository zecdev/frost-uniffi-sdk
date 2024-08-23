#!/bin/sh
set -euxo pipefail
sh Scripts/build_swift.sh

sh Scripts/replace_remote_binary_with_local.sh

swift test