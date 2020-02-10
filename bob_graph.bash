#!/bin/bash

# Copyright 2018-2020 Arm Limited.
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# Example usage
# ./bob_graph --graph-start-nodes=libMy,libOther
#
# To view users of libOther
# ./bob_graph --graph-start-nodes=libOther --graph-rev-deps

# Switch to the build directory
cd "$(dirname "${BASH_SOURCE[0]}")"

# Read settings written by bootstrap.bash
source ".bob.bootstrap"

# Switch to the working directory
cd "${WORKDIR}"

BOB_BUILDER_TARGET=".bootstrap/bin/bob"
BOB_BUILDER="${BUILDDIR}/${BOB_BUILDER_TARGET}"
BOB_BUILDER_NINJA="${BUILDDIR}/.bootstrap/build.ninja"

if [ ! -f "${BOB_BUILDER_NINJA}" ]; then
	echo "Missing ${BOB_BUILDER_NINJA}"
	echo "Please build your project first"
	exit 1
fi

# Make sure Bob is built
ninja -f "${BOB_BUILDER_NINJA}" "${BOB_BUILDER_TARGET}"

echo "
#
# Legend description
#
# Nodes
# green node            - static library
# orange node           - shared library
# gray node             - binary
# yellow node           - default

# Marked node
# double circle         - marked node

# Edges
# orange edges          - content of shared_libs
# orange-dashed edges   - content of export_shared_libs
# green edges           - content of static_libs
# green-dashed edges    - content of export_static_libs
# red edges             - content of whole_static
# yellow edges          - content of defaults
#
"

"${BOB_BUILDER}" -l "${BLUEPRINT_LIST_FILE}" -b "${BUILDDIR}" "$@" "${SRCDIR}/${TOPNAME}"
