/*
 * Copyright 2023 Arm Limited.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"os"
	"path/filepath"

	"github.com/ARM-software/bob-build/internal/utils"
	"github.com/google/blueprint"
)

// In Bazel, some properties are transitive.
type TransitiveLibraryProps struct {
	Defines []string
}

func (m *TransitiveLibraryProps) defines() []string {
	return m.Defines
}

type StrictLibraryProps struct {
	Hdrs []string
	// TODO: Header inclusion
	//Textual_hdrs           []string
	//Includes               []string
	//Include_prefixes       []string
	//Strip_include_prefixes []string

	Local_defines []string
	Copts         []string
	Deps          []string
	Out           *string // TODO:
	// unused but needed for the output interface, no easy way to hide it
	TargetType tgtType `blueprint:"mutated"`
}

type strictLibrary struct {
	moduleBase
	simpleOutputProducer // band-aid so legacy don't complain the interface isn't implemented
	Properties           struct {
		StrictLibraryProps
		SourceProps
		TransitiveLibraryProps
		Features
		EnableableProps
		SplittableProps
		InstallableProps
	}

	Shared struct {
		simpleOutputProducer
	}

	Static struct {
		simpleOutputProducer
	}
}

// All libraries must support:
var _ splittable = (*strictLibrary)(nil)
var _ dependentInterface = (*strictLibrary)(nil)
var _ sourceInterface = (*library)(nil)

// shared libraries must supports:
// * producing output using the linker
// * producing a shared library
// * stripping symbols from output
// var _ linkableModule = (*strictLibrary)(nil)
// var _ sharedLibProducer = (*strictLibrary)(nil)
// var _ stripable = (*strictLibrary)(nil)

func (m *strictLibrary) processPaths(ctx blueprint.BaseModuleContext, g generatorBackend) {
	// TODO: Handle Bazel targets & check paths
	prefix := projectModuleDir(ctx)
	m.Properties.SourceProps.processPaths(ctx, g)
	m.Properties.Hdrs = utils.PrefixDirs(m.Properties.Hdrs, prefix)
}

func (m *strictLibrary) filesToInstall(ctx blueprint.BaseModuleContext) []string {
	// TODO: Header only installs
	return append(m.Static.outputs(), m.Shared.outputs()...)
}

func (l *strictLibrary) outputName() string {
	if l.Properties.Out != nil {
		return *l.Properties.Out
	}
	return l.Name()
}

func (m *strictLibrary) outputFileName() string {
	utils.Die("Cannot use outputFileName on strict_library")
	return "badName"
}

func (l *strictLibrary) ObjDir() string {
	return filepath.Join("${BuildDir}", string(l.Properties.TargetType), "objects", l.outputName()) + string(os.PathSeparator)
}

func (l *strictLibrary) getSourceFiles(ctx blueprint.BaseModuleContext) []string {
	return l.Properties.SourceProps.getSourceFiles(ctx)

}
func (l *strictLibrary) getSourceTargets(ctx blueprint.BaseModuleContext) []string {
	return l.Properties.SourceProps.getSourceTargets(ctx)

}
func (l *strictLibrary) getSourcesResolved(ctx blueprint.BaseModuleContext) []string {
	return l.Properties.SourceProps.getSourcesResolved(ctx)
}

func (l *strictLibrary) getSrcs() []string {
	return l.Properties.Srcs
}

func (l *strictLibrary) supportedVariants() (tgts []tgtType) {
	// TODO: Change tgts based on if host or target supported.
	tgts = append(tgts, tgtTypeHost)
	return
}

func (l *strictLibrary) disable() {
	f := false
	l.Properties.Enabled = &f
}

func (l *strictLibrary) setVariant(tgt tgtType) {
	l.Properties.TargetType = tgt
}

func (l *strictLibrary) getTarget() tgtType {
	return l.Properties.TargetType
}

func (l *strictLibrary) getSplittableProps() *SplittableProps {
	return &l.Properties.SplittableProps
}

func (l *strictLibrary) getEnableableProps() *EnableableProps {
	return &l.Properties.EnableableProps
}

func (l *strictLibrary) getInstallableProps() *InstallableProps {
	return &l.Properties.InstallableProps
}

func (m *strictLibrary) GenerateBuildActions(ctx blueprint.ModuleContext) {
	getBackend(ctx).strictLibraryActions(m, ctx)
}

func (m *strictLibrary) shortName() string {
	return m.Name()
}

// Shared Library BoB Interface

func (m *strictLibrary) getTocName() string {
	// TODO: Does this need to be m.getRealName() It is in other impls
	// what does getRealName() look like?
	return m.Name() + tocExt
}

func LibraryFactory(config *BobConfig) (blueprint.Module, []interface{}) {
	module := &strictLibrary{}
	module.Properties.Features.Init(&config.Properties, StrictLibraryProps{})
	return module, []interface{}{&module.Properties,
		&module.SimpleName.Properties}
}
