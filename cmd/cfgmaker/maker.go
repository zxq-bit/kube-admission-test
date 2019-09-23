package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
func (c *Config) OutputFileName() string {
	return fmt.Sprintf("%s.go", c.ResourceName)
}

var (
	templatePath string
	configPath   string
	outputPath   string
)

func flagInit() {
	flag.StringVar(&templatePath, "templatePath", "object.go.yaml", "gen config file path")
	flag.StringVar(&configPath, "configPath", "./config.yaml", "output template file path")
	flag.StringVar(&outputPath, "outputPath", "./output", "gen output base dir path")
	flag.Parse()
	klog.Infof("parse templatePath: %v", templatePath)
	klog.Infof("parse configPath: %v", configPath)
	klog.Infof("parse outputPath: %v", outputPath)
}

func initTemplate() (t *template.Template) {
	b, e := ioutil.ReadFile(templatePath)
	if e != nil {
		klog.Exitf("read template file failed, %v", e)
	}
	t, e = template.New("object").Parse(string(b))
	if e != nil {
		klog.Exitf("parse template failed, %v", e)
	}
	return t
}

func main() {
	flagInit()
	t := initTemplate()
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
		// exec
		wr := new(bytes.Buffer)
		if e = t.Execute(wr, c); e != nil {
			klog.Exitf("[%d] %v execute failed, %v", i, c, e)
		}
		// write
		e = ioutil.WriteFile(filepath.Join(outputPath, c.OutputDir(), c.OutputFileName()), wr.Bytes(), 0664)
		if e != nil {
			klog.Exitf("[%d] %v write failed, %v", i, c, e)
		}
	}
}
