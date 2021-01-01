#!/bin/sh
# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE in the project root for license information.
# ----------------------------------------------------------------

#Using protoc version 3.0.0

cd gauge-proto
PATH=$PATH:$GOPATH/bin protoc --go_out=plugins=grpc:../gauge_messages spec.proto messages.proto services.proto