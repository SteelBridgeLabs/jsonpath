/*
 * Copyright 2020 VMware, Inc.
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Changes:
 *   - Changed package name from github.com/vmware-labs/yamlpath to github.com/SteelBridgeLabs/jsonpath
 *   - Removed YAML implementation and added JSON implementation
 */

package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/SteelBridgeLabs/jsonpath"
)

func main() {
	tmpl := template.New("template")
	tmpl, err := tmpl.Parse(`<style type="text/css">
.tg  {border-collapse:collapse;border-spacing:0;}
.tg td{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
  overflow:hidden;padding:10px 5px;word-break:normal;}
.tg th{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
  font-weight:normal;overflow:hidden;padding:10px 5px;word-break:normal;}
.tg .tg-zv4m{border-color:#ffffff;text-align:left;vertical-align:top}
textarea, pre, input {font-family:Consolas,monospace; font-size:14px}
h1, body, label {font-family: Lato,proxima-nova,Helvetica Neue,Arial,sans-serif}
textarea, input {
	box-sizing: border-box;
	border: 1px solid;
	background-color: #f8f8f8;
	resize: none;
  }
</style>
{{if .Version}}
<span title="version: {{ .Version }}">
{{end}}
<h1>JsonPath evaluator</h1>
{{if .Version}}
</span>
{{end}}
<table class="tg">
<thead>
  <tr valign="top">
	<th class="tg-zv4m">
<form method="POST">
<label>JSON document</label>:<br />
<pre>
<textarea name="JSON document" cols="80" rows="30" placeholder="JSON...">{{ .JSON }}</textarea>
</pre><br /><br />
<label>JSON path</label>
(<a href="https://github.com/vmware-labs/yaml-jsonpath/tree/{{ .Version }}#syntax" target="_blank">syntax</a>):<br />
<pre>
<input type="text" size="80" name="JSON path" placeholder="JSON path..." value="{{ .JSONPath }}"><br />
<input type="submit" value="Evaluate">
</pre>
</form>

	</th>
	<th class="tg-zv4m">
	   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	</th>
	<th class="tg-zv4m">
	<label>Output:</label><br /><br />
{{if .JSONError}}
	<br />{{ .JSONError }}<br />
{{end}}
{{if .JSONPathError}}
    <br />Invalid JSON path: {{ .JSONPathError }}<br />
{{end}}
<pre>
{{ .Output }}<br />
</pre>
	</th>
  </tr>
</thead>
</table>
`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type output struct {
			JSON          string
			JSONError     error
			JSONPath      string
			JSONPathError error
			Success       bool
			Output        string
			Version       string
		}

		op := output{
			Version: os.Getenv("GAE_VERSION"),
		}

		if r.Method != http.MethodPost {
			if e := tmpl.Execute(w, op); e != nil {
				respondWithError(w, e)
			}
			return
		}

		y := r.FormValue("JSON document")
		op.JSON = y

		problem := false

		// parse JSON
		var value interface{}
		if err := json.Unmarshal([]byte(y), &value); err != nil {
			problem = true
			op.JSONError = err
		}

		j := r.FormValue("JSON path")
		op.JSONPath = j
		path, err := jsonpath.NewPath(j)
		if err != nil {
			problem = true
			op.JSONPathError = err
		}

		if problem {
			if e := tmpl.Execute(w, op); e != nil {
				respondWithError(w, e)
			}
			return
		}

		results, err := path.Evaluate(value)
		if err != nil {
			respondWithError(w, err)
		}

		// encode results
		op.Output, _ = encode(results)
		op.Success = true

		if e := tmpl.Execute(w, op); e != nil {
			respondWithError(w, e)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func encode(value any) (string, error) {
	// buffer
	var buffer bytes.Buffer
	// json encoder
	encoder := json.NewEncoder(&buffer)
	// encode '<'...
	encoder.SetEscapeHTML(true)
	// pretty print
	encoder.SetIndent("", "  ")
	// encode value
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func respondWithError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
