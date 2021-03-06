package lgn

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-playground/form"
	"github.com/zew/go-questionnaire/lgn/shuffler"
	"github.com/zew/util"
)

// ShufflesToCSV computes random but deterministic shufflings for usage outside the app
func ShufflesToCSV(w http.ResponseWriter, r *http.Request) {

	errMsg := ""

	_, ok := r.PostForm["token"]
	if ok {
		err := ValidateFormToken(r.PostForm.Get("token"))
		if err != nil {
			errMsg += fmt.Sprintf("Invalid request token: %v\n", err)
		}
	} else if !ok && r.Method == "POST" {
		errMsg += fmt.Sprintf("Missing request token\n")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	src := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title>Shufflings</title>
</head>
<body>
	<form method="post" action="{{.SelfURL}}"  style="margin: 50px;"  >
		
		{{if gt (len .ErrMsg) 0 }} <p style='white-space: pre; color:#E22'>{{.ErrMsg}}</p>{{end}}
		
		Create Shufflings<br>

		                <input name="token"       type="hidden"   value="{{.Token}}" />

		Variations:     <input name="variations"  type="text"     value="{{.DL.Variations}}"  placeholder="8" ><br>

		<br>
		Groups on page  (questions on page):  <br> 
			Page 0 		<input name="grop[0]"      type="text"     value="{{index .DL.GroupsOnPage 0}}"   ><br>
			Page 1 		<input name="grop[1]"      type="text"     value="{{index .DL.GroupsOnPage 1}}"   ><br>
			Page 2 		<input name="grop[2]"      type="text"     value="{{index .DL.GroupsOnPage 2}}"   ><br>
			Page 3 		<input name="grop[3]"      type="text"     value="{{index .DL.GroupsOnPage 3}}"   ><br>
			Page 4 		<input name="grop[4]"      type="text"     value="{{index .DL.GroupsOnPage 4}}"   ><br>

		<br>
		User ID:<br>
			Start 	    <input name="start"       type="text"     value="{{.DL.Start}}"><br>
			Stop 	    <input name="stop"        type="text"     value="{{.DL.Stop}}" ><br>

		<input type="submit"   name="submitclassic" accesskey="s"><br>

		{{if  gt (len .List   ) 0 }} <p style='white-space: pre; color:#444'>{{.List   }}</p>{{end}}


	</form>

</body>
</html>
`
	pages := 5

	type formEntryT struct {
		Token string `json:"token"`

		Variations    int    `json:"variations"`
		GroupsOnPage  []int  `json:"grop"`
		Start         int    `json:"start"`
		Stop          int    `json:"stop"`
		SubmitClassic string `json:"submitclassic"`
	}
	fe := formEntryT{}

	//
	// dec := for--mam.NewDecoder(&for--mam.DecoderOptions{TagName: "json"})
	dec := form.NewDecoder()
	dec.SetTagName("json") // recognizes and ignores ,omitempty
	err := dec.Decode(&fe, r.Form)
	if err != nil {
		errMsg += fmt.Sprintf("Decoding error: %v\n", err)
	}
	log.Printf(util.IndentedDump(fe))

	if fe.Variations == 0 {
		fe.Variations = 6
	}

	defaultV := 2
	for len(fe.GroupsOnPage) < pages {
		fe.GroupsOnPage = append(fe.GroupsOnPage, defaultV)
		defaultV++
	}

	if fe.Start == 0 {
		fe.Start = 1000
	}
	if fe.Stop == 0 {
		fe.Stop = 1020
	}

	list := ""

	for i1 := fe.Start; i1 <= fe.Stop; i1++ {
		for i2 := 0; i2 < pages; i2++ {
			grOP := fe.GroupsOnPage[i2]
			sh := shuffler.New(fmt.Sprintf("%v", i1), fe.Variations, grOP)
			order := sh.Slice(i2)
			list += fmt.Sprintf("%5v\t%v\t%v\t%v\n", i1, i2, grOP, order)
		}
	}

	type dataT struct {
		SelfURL string
		Token   string

		ErrMsg string
		DL     formEntryT
		List   string
	}
	data := dataT{
		SelfURL: r.URL.Path,
		Token:   FormToken(),
		ErrMsg:  errMsg,
		DL:      fe,
		List:    list,
	}

	tpl := template.New("anyname.html")
	tpl, err = tpl.Parse(src)
	if err != nil {
		fmt.Fprintf(w, "Error parsing inline template: %v", err)
	}

	err = tpl.Execute(w, data)
	if err != nil {
		fmt.Fprintf(w, "Error executing inline template: %v", err)
	}

}
