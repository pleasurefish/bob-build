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
	"github.com/ARM-software/bob-build/core/file"
	"github.com/ARM-software/bob-build/core/flag"

	"github.com/google/blueprint"
)

/*
	We are swapping from `bob_generate_source` to `bob_genrule`

`bob_genrule` is made to be a stricter version that is compatible with Android.
For easiest compatibility, we are using Androids format for `genrule`.
Some properties in the struct may not be useful, but it is better to expose as many
features as possible rather than too few. Some are commented out as they would take special
implementation for features we do not already have in place.

*/

type GenruleProps struct {
	Out         []string
	ResolvedOut file.Paths `blueprint:"mutated"`
}

type ModuleGenrule struct {
	ModuleStrictGenerateCommon
	Properties struct {
		GenruleProps
	}
}

type ModuleGenruleInterface interface {
	FileConsumer
	FileResolver
	pathProcessor
}

var _ ModuleGenruleInterface = (*ModuleGenrule)(nil) // impl check

func (m *ModuleGenrule) implicitOutputs() []string {
	return m.OutFiles().ToStringSliceIf(
		func(f file.Path) bool { return f.IsType(file.TypeImplicit) },
		func(f file.Path) string { return f.BuildPath() })
}

func (m *ModuleGenrule) outputs() []string {
	return m.OutFiles().ToStringSliceIf(
		func(f file.Path) bool { return f.IsNotType(file.TypeImplicit) },
		func(f file.Path) string { return f.BuildPath() })
}

func (m *ModuleGenrule) processPaths(ctx blueprint.BaseModuleContext) {
	m.ModuleStrictGenerateCommon.processPaths(ctx)
}

func (m *ModuleGenrule) ResolveFiles(ctx blueprint.BaseModuleContext) {
	m.ModuleStrictGenerateCommon.ResolveFiles(ctx)

	files := file.Paths{}
	for _, out := range m.Properties.Out {
		fp := file.NewPath(out, ctx.ModuleName(), file.TypeGenerated)
		files = files.AppendIfUnique(fp)
	}

	m.Properties.ResolvedOut = files
}

func (m *ModuleGenrule) GetFiles(ctx blueprint.BaseModuleContext) file.Paths {
	return m.ModuleStrictGenerateCommon.Properties.GetFiles(ctx)
}

func (m *ModuleGenrule) GetDirectFiles() file.Paths {
	return m.ModuleStrictGenerateCommon.Properties.GetDirectFiles()
}

func (m *ModuleGenrule) GetTargets() []string {
	return m.ModuleStrictGenerateCommon.Properties.GetTargets()
}

func (m *ModuleGenrule) OutFiles() file.Paths {

	return m.Properties.ResolvedOut
}

func (m *ModuleGenrule) OutFileTargets() (tgts []string) {
	// does not forward any of it's source providers.
	return
}

func (m *ModuleGenrule) FlagsOut() (flags flag.Flags) {
	gc := m.getStrictGenerateCommon()
	for _, str := range gc.Properties.Export_include_dirs {
		flags = append(flags, flag.FromGeneratedIncludePath(str, m, flag.TypeExported))
	}
	return
}

func (m *ModuleGenrule) shortName() string {
	return m.Name()
}

func (m *ModuleGenrule) generateInouts(ctx blueprint.ModuleContext) []inout {
	var io inout

	m.GetFiles(ctx).ForEachIf(
		// TODO: The current generator does pass parse .toc files when consuming generated shared libraries.
		func(fp file.Path) bool { return fp.IsNotType(file.TypeToc) },
		func(fp file.Path) bool {
			if fp.IsType(file.TypeImplicit) {
				io.implicitSrcs = append(io.implicitSrcs, fp.BuildPath())
			} else {
				io.in = append(io.in, fp.BuildPath())
			}
			return true
		})

	io.out = m.Properties.Out

	if depfile, ok := m.OutFiles().FindSingle(
		func(p file.Path) bool { return p.IsType(file.TypeDep) }); ok {
		io.depfile = depfile.UnScopedPath()
	}

	return []inout{io}
}

func (m *ModuleGenrule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	if isEnabled(m) {
		g := getGenerator(ctx)
		g.genruleActions(m, ctx)
	}
}

func (m ModuleGenrule) GetProperties() interface{} {
	return m.Properties
}

func (m *ModuleGenrule) FeaturableProperties() []interface{} {
	return append(m.ModuleStrictGenerateCommon.FeaturableProperties(), &m.Properties.GenruleProps)
}

func (m *ModuleGenrule) getStrictGenerateCommon() *ModuleStrictGenerateCommon {
	return &m.ModuleStrictGenerateCommon
}

func generateRuleAndroidFactory(config *BobConfig) (blueprint.Module, []interface{}) {
	module := &ModuleGenrule{}

	module.ModuleStrictGenerateCommon.init(&config.Properties,
		StrictGenerateProps{}, GenruleProps{}, EnableableProps{})

	return module, []interface{}{&module.ModuleStrictGenerateCommon.Properties, &module.Properties,
		&module.SimpleName.Properties}
}
