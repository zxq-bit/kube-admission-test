package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
	"k8s.io/klog"
)

type Config struct {
	APIGroup   string
	APIVersion string

	PkgPath      string
	PkgName      string
	ResourceName string
	KindName     string
}

func (c *Config) Validate() error {
	if c.APIGroup == "" {
		return fmt.Errorf("config APIGroup is empty")
	}
	if c.APIVersion == "" {
		return fmt.Errorf("config APIVersion is empty")
	}

	if c.PkgPath == "" {
		return fmt.Errorf("config PkgPath is empty")
	}
	if c.PkgName == "" {
		return fmt.Errorf("config PkgName is empty")
	}
	if c.ResourceName == "" {
		return fmt.Errorf("config ResourceName is empty")
	}
	if c.KindName == "" {
		return fmt.Errorf("config KindName is empty")
	}
	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("[%v.%v]", c.PkgName, c.KindName)
}

func (c *Config) OutputDir() string {
	return filepath.Join(c.APIGroup, c.APIVersion)
}
func (c *Config) OutputFileName(suffixes ...string) string {
	vec := []string{c.ResourceName}
	for _, s := range suffixes {
		if s != "" {
			vec = append(vec, s)
		}
	}
	return fmt.Sprintf("%s.go", strings.Join(vec, "_"))
}

var (
	templatePath string
	configPath   string
	outputPath   string
)

func flagInit() {
	flag.StringVar(&templatePath, "templatePath", "object.gohtml", "gen config file path")
	flag.StringVar(&configPath, "configPath", "./config.yaml", "output template file path")
	flag.StringVar(&outputPath, "outputPath", "./output", "gen output base dir path")
	flag.Parse()
	klog.Infof("parse templatePath: %v", templatePath)
	klog.Infof("parse configPath: %v", configPath)
	klog.Infof("parse outputPath: %v", outputPath)
}

func initTemplate() (tm map[string]*template.Template) {
	var (
		splitString = "\n<<<---=== i am split ===--->>>\n"
		names       = []string{"types", "operations"}
	)
	b, e := ioutil.ReadFile(templatePath)
	if e != nil {
		klog.Exitf("read template file failed, %v", e)
	}
	raw := string(b)
	texts := strings.Split(raw, splitString)
	switch len(texts) {
	case 1:
		tm = make(map[string]*template.Template, 1)
		name := ""
		t, e := template.New(name).Parse(raw)
		if e != nil {
			klog.Exitf("parse template %s failed, %v", name, e)
		}
		tm[name] = t
	case len(names):
		tm = make(map[string]*template.Template, len(names))
		for i, name := range names {
			text := texts[i]
			t, e := template.New(name).Parse(text)
			if e != nil {
				klog.Exitf("parse template %s failed, %v", name, e)
			}
			tm[name] = t
		}
	default:
		klog.Exitf("parse template failed, split not correct")
	}
	return
}

func main() {
	flagInit()
	tm := initTemplate()
	// read
	b, e := ioutil.ReadFile(configPath)
	if e != nil {
		klog.Exitf("read config failed, %v", e)
	}
	// parse
	var configs []Config
	if e = yaml.Unmarshal(b, &configs); e != nil {
		klog.Exitf("unmarshal configs failed, %v", e)
	}
	// validate
	for i, c := range configs {
		// validate
		if e = c.Validate(); e != nil {
			klog.Exitf("[%d] %v validate failed, %v", i, c, e)
		}
		// dir
		if e = os.MkdirAll(filepath.Join(outputPath, c.OutputDir()), 0775); e != nil {
			klog.Exitf("[%d] %v mkdir failed, %v", i, c, e)
		}
		for name, t := range tm {
			// exec
			wr := new(bytes.Buffer)
			if e = t.Execute(wr, c); e != nil {
				klog.Exitf("[%d] %v execute %s failed, %v", i, c, name, e)
			}
			// write
			e = ioutil.WriteFile(filepath.Join(outputPath, c.OutputDir(), c.OutputFileName(name)), wr.Bytes(), 0664)
			if e != nil {
				klog.Exitf("[%d] %v write %s failed, %v", i, c, name, e)
			}
		}
	}
}
