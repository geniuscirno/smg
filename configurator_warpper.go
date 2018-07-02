package smg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/geniuscirno/smg/configurator"
	_ "github.com/geniuscirno/smg/configurator/etcd"
)

type appConfiguratorWarpper struct {
	configurator configurator.Configurator
	app          *application
	localCfgRoot string
}

func parseConfiguratorTarget(target string) (ret configurator.Target, err error) {
	var ok bool
	ret.Scheme, ret.Endpoint, ok = split2(target, "://")
	if !ok {
		return ret, errors.New("parseConfiguratorTarget: invalid target")
	}
	ret.Authority, ret.Endpoint, _ = split2(ret.Endpoint, "/")
	return ret, nil
}

func newAppConfiguratorWarpper(app *application) (*appConfiguratorWarpper, error) {
	var err error

	warpper := &appConfiguratorWarpper{app: app, localCfgRoot: path.Join(app.opts.prefix, "cfg")}

	cb, ok := configurator.Get(app.parsedTarget.Scheme)
	if !ok {
		return nil, fmt.Errorf("could not get resolver for scheme: %q", app.parsedTarget.Scheme)
	}

	warpper.configurator, err = cb.Build(app.parsedTarget)
	if err != nil {
		return nil, err
	}

	localMeta, err := warpper.loadLocalMeta()
	if err != nil {
		return nil, err
	}

	files, err := warpper.globLocalCfg()
	if err != nil {
		return nil, err
	}

	err = warpper.configurator.Upload(app.CfgRoot()+"/", localMeta, warpper.fileIterator(files))
	if err != nil {
		return nil, err
	}
	return warpper, nil
}

func (warpper *appConfiguratorWarpper) fileIterator(files []string) func() (string, []byte, bool, error) {
	index := 0
	return func() (string, []byte, bool, error) {
		file := files[index]
		b, err := ioutil.ReadFile(path.Join(warpper.localCfgRoot, file))
		if err != nil {
			return "", nil, false, err
		}
		index++
		return filepath.ToSlash(path.Join(warpper.app.CfgRoot(), file)), b, index < len(files), nil
	}
}

func (warpper *appConfiguratorWarpper) globLocalCfgFiles(patterns []string, path string) ([]string, error) {
	uniquePatterns := make(map[string]struct{})
	for _, pattern := range patterns {
		if _, ok := uniquePatterns[pattern]; !ok {
			uniquePatterns[pattern] = struct{}{}
		}
	}

	uniqueMatched := make(map[string]struct{})
	for pattern := range uniquePatterns {
		matched, err := warpper.glob(path, pattern, true)
		if err != nil {
			return nil, err
		}
		for _, m := range matched {
			if _, ok := uniqueMatched[m]; !ok {
				uniqueMatched[m] = struct{}{}
			}
		}
	}

	matched := make([]string, 0, len(uniqueMatched))
	for match := range uniqueMatched {
		matched = append(matched, match)
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i] < matched[j] })
	return matched, nil
}

func (*appConfiguratorWarpper) ignoreLocalCfgFiles(patterns []string, matched []string) ([]string, error) {
	uniquePatterns := make(map[string]struct{})
	for _, pattern := range patterns {
		if _, ok := uniquePatterns[pattern]; !ok {
			uniquePatterns[pattern] = struct{}{}
		}
	}

	alives := make([]string, 0, len(matched))
	for _, match := range matched {
		ignore := false
		for pattern := range uniquePatterns {
			ok, err := filepath.Match(pattern, match)
			if err != nil {
				return nil, err
			}
			if ok {
				ignore = true
				break
			}
		}
		if !ignore {
			alives = append(alives, match)
		}
	}
	return alives, nil
}

func (warpper *appConfiguratorWarpper) globLocalCfg() ([]string, error) {
	matched, err := warpper.globLocalCfgFiles(warpper.app.opts.includeCfgPattern, warpper.localCfgRoot)
	if err != nil {
		return nil, err
	}
	return warpper.ignoreLocalCfgFiles(warpper.app.opts.ignoreCfgPattern, matched)
}

func (*appConfiguratorWarpper) glob(dir string, pattern string, recursive bool) ([]string, error) {
	if !recursive {
		return filepath.Glob(filepath.Join(dir, pattern))
	}

	matched := make([]string, 0)
	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if filePath == filepath.Join(dir, "__META") {
				return nil
			}
			match, err := filepath.Match(pattern, filepath.Base(filePath))
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(dir, filePath)
			if err != nil {
				return err
			}
			if match {
				matched = append(matched, rel)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matched, nil
}

func (warpper *appConfiguratorWarpper) loadLocalMeta() (*configurator.Meta, error) {
	b, err := ioutil.ReadFile(path.Join(warpper.localCfgRoot, "__META"))
	if err != nil {
		return nil, err
	}

	meta := &configurator.Meta{}
	err = json.Unmarshal(b, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
