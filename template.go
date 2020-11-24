package lazygo

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type AssetRegister func(name string) ([]byte, error)

type Template struct {
	tpl     *template.Template
	prefix  string
	suffix  string
	TplData map[string]interface{}
	Res     http.ResponseWriter
	Req     *http.Request
	Header  map[string]string
	asset   AssetRegister
}

func NewTemplate(res http.ResponseWriter, req *http.Request, prefix string, suffix string, asset AssetRegister) *Template {
	return &Template{
		prefix:  prefix,
		suffix:  suffix,
		TplData: make(map[string]interface{}),
		Res:     res,
		Req:     req,
		Header:  map[string]string{},
		asset:   asset,
	}
}

func (t *Template) Parse(content string) *Template {
	var err error
	t.tpl, err = t.tpl.Parse(content)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) getDefinedTpl(file string) []string {
	file = t.prefix + file + t.suffix
	tpls := []string{
		file,
	}

	return tpls
}

func (t *Template) ParseFiles(tplName ...string) *Template {
	commFiles := t.getDefinedTpl(tplName[0])
	for _, f := range tplName {
		ff := t.prefix + f + t.suffix
		commFiles = append(commFiles, ff)
	}

	var err error
	t.tpl, err = t.parseFiles(commFiles)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) parseFiles(filenames []string) (*template.Template, error) {
	for _, filename := range filenames {
		if t.asset == nil {
			return nil, fmt.Errorf("未注册资源")
		}
		b, err := t.asset(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		name := filepath.Base(filename)
		var tmpl *template.Template
		if t.tpl == nil {
			t.tpl = template.New(name)
		}
		if name == t.tpl.Name() {
			tmpl = t.tpl
		} else {
			tmpl = t.tpl.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t.tpl, nil
}

func (t *Template) AssignMap(data map[string]interface{}) *Template {
	for k, v := range data {
		t.TplData[k] = v
	}
	return t
}

func (t *Template) Assign(key string, data interface{}) *Template {
	t.TplData[key] = data
	return t
}

func (t *Template) Display(tpl string) {
	err := t.ParseFiles(tpl).tpl.Execute(t.Res, t.TplData)
	if err != nil {
		panic(err.Error())
	}
}

func (t *Template) DisplayMulti(tpl string, subTpl []string) {
	tpls := []string{tpl}
	tpls = append(tpls, subTpl...)
	err := t.ParseFiles(tpls...).tpl.Execute(t.Res, t.TplData)
	if err != nil {
		panic(err.Error())
	}
}
