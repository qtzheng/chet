package chet

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	chetTplFuncMap  template.FuncMap
	ChetTemplates   map[string]*template.Template
	ChetTemplateExt []string = []string{"html", "tpl"}
	TemplateLeft             = "{{"
	TemplateRight            = "}}"
)

type templateDir struct {
	root string
}
type templateFile struct {
	filePath string
	root     string
	data     string
}

func (t *templateDir) scanTemplate(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	unixPath := filepath.ToSlash(path) //为了使windows，类unix系统之间的目录格式能统一
	a := []byte(unixPath)
	a = a[len([]byte(t.root)):]
	rel_path := strings.TrimLeft(string(a), "/")
	fmt.Println(rel_path)

	tplFile := getTemplate(t.root, rel_path, "")
	temp, err_t := template.New(rel_path).Parse(tplFile.data)
	if err_t != nil {
		return err_t
	}
	ChetTemplates[rel_path] = temp
	return nil
}
func init() {
	chetTplFuncMap = make(template.FuncMap)
	ChetTemplates = make(map[string]*template.Template)
}
func BuildTemplates(tplDirPath string) error {
	if _, err := os.Stat(tplDirPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s is not exist", tplDirPath)
		} else {
			return fmt.Errorf("dir:%s open error", tplDirPath)
		}
	}
	tplDir := &templateDir{tplDirPath}
	err := filepath.Walk(tplDir.root, func(path string, f os.FileInfo, err error) error {
		return tplDir.scanTemplate(path, f, err)
	})
	return err
}
func getTemplate(root, filePath, parentPath string) *templateFile {
	var absPath string
	if strings.HasPrefix(filePath, "../") {
		absPath = filepath.Join(root, filepath.Dir(parentPath), filePath)
	} else {
		absPath = filepath.Join(root, filePath)
	}
	if _, err := os.Stat(absPath); err != nil {
		panic("can't find template file:" + absPath)
	}
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		panic("can't read template file:" + absPath)
	}
	tplFile := &templateFile{
		root:     root,
		filePath: filePath,
		data:     string(data),
	}
	reg := regexp.MustCompile(TemplateLeft + "[ ]*template[ ]+\"([^\"]+)\".*}}")
	allSub := reg.FindAllStringSubmatch(tplFile.data, -1)
	for _, m := range allSub {
		if len(m) == 2 {
			t := getTemplate(root, m[1], filePath)
			replace := strings.NewReplacer(m[0], t.data)
			tplFile.data = replace.Replace(tplFile.data)
		}
	}
	return tplFile
}
