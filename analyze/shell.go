package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

var getJSON = `
package main

import (
	"encoding/json"
	"os"
	"github.com/Logiraptor/nap/fill"
	"reflect"
	"math/rand"
	{{range .Imports}}"{{.}}"
	{{end}}
)

func main() {
	var x {{printType .Type}}

	fill.Fill(reflect.ValueOf(&x).Elem(), rand.New(rand.NewSource(0)))

	buf, _ := json.MarshalIndent(x, "", "\t")
	os.Stdout.Write(buf)
}

`

func getOutput(src string, data interface{}) string {
	fileName := "nap_temp_.go"
	tmpl := template.New("temp").Funcs(template.FuncMap{
		"printType": printTypeWrapper,
	})
	tmpl, err := tmpl.Parse(src)
	if err != nil {
		return err.Error()
	}
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, "temp", data)
	if err != nil {
		return err.Error()
	}
	err = ioutil.WriteFile(fileName, buf.Bytes(), 0666)
	if err != nil {
		return err.Error()
	}
	defer os.Remove(fileName)

	cmd := exec.Command("go", "run", fileName)
	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	return string(output)
}
