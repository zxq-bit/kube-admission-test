package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
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

	IsCRD bool
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
	if c.IsCRD && strings.HasPrefix(c.PkgPath, "k8s.io") {
		return fmt.Errorf("config is in k8s repo but set as IsCRD")
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
	flag.StringVar(&outputPath, "outputPath", "./output", "gen output file path")
	flag.Parse()
	klog.Infof("parse templatePath: %v", templatePath)
	klog.Infof("parse configPath: %v", configPath)
	klog.Infof("parse outputPath: %v", outputPath)
}

func initTemplate() *template.Template {
	b, e := ioutil.ReadFile(templatePath)
	if e != nil {
		klog.Exitf("read template file failed, %v", e)
	}
	raw := string(b)
	name := ""
	t, e := template.New(name).Parse(raw)
	if e != nil {
		klog.Exitf("parse template %s failed, %v", name, e)
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
	klog.V(3).Infoln("parsed configs:\n" + toYaml(configs))
	// validate
	for i, c := range configs {
		// validate
		if e = c.Validate(); e != nil {
			klog.Exitf("[%d] %v validate failed, %v", i, c, e)
		}
	}
	// exec
	wr := new(bytes.Buffer)
	if e = t.Execute(wr, configs); e != nil {
		klog.Exitf("execute failed, %v", e)
	}
	// write
	e = ioutil.WriteFile(outputPath, wr.Bytes(), 0664)
	if e != nil {
		klog.Exitf("write %s failed, %v", outputPath, e)
	}
}

func toYaml(v interface{}) string {
	b, _ := yaml.Marshal(v)
	return string(b)
}
