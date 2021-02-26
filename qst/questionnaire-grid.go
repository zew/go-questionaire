package qst

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/zew/go-questionnaire/css"
)

func tdDummy(a horizontalAlignment, b, payload string) string {
	return payload
}

func wrap(w io.Writer, tagName, forVal, className, style, content string) {

	forKeyVal := ""
	if forVal != "" {
		forKeyVal = fmt.Sprintf("for='%v'", forVal)

	}

	if style != "" {
		styles := strings.Split(style, ";")
		style = strings.Join(styles, ";\n\t\t")
		style = fmt.Sprintf("%v%v%v", "\n\t\t", style, "\n\t\t")
	}

	fmt.Fprintf(w,
		"<%v %v class='%v' style='%v' >\n%v\n</%v>\n",
		tagName,
		forKeyVal,
		className,
		style,
		content,
		tagName,
	)

	// fmt.Fprintf(w, "<!-- /%v -->\n", className)

}
func divWrap(w io.Writer, className, style, content string) {
	wrap(w, "div", "", className, style, content)
}

func (inp inputT) labelDescription(w io.Writer, langCode string) {

	if inp.Label.Empty() && inp.Desc.Empty() {
		return
	}

	if inp.Type == "textblock" {
		if !inp.Label.Empty() {
			fmt.Fprint(w, inp.Label.Tr(langCode))
		}
		if !inp.Desc.Empty() {
			fmt.Fprint(w, inp.Desc.Tr(langCode))
		}
		return
	}

	// classes are only for font-size, font-weight
	// inline-block styles are applied in outer wrapper
	if !inp.Label.Empty() {
		fmt.Fprintf(w, " <span class='input-label-text'       >%v</span>", inp.Label.Tr(langCode))
	}
	if !inp.Desc.Empty() {
		fmt.Fprintf(w, " <span class='input-description-text' >%v</span>", inp.Desc.Tr(langCode))
	}

}

// ShortSuffix appends suffix to the input - using &nbsp;
// for longer text, reverse order of label and control;
// sadly, &nbsp; has no effect;
// we have to use white-space: nowrap; on the grid-item
func (inp *inputT) ShortSuffix(ctrl string, langCode string) string {

	if inp.Suffix.Empty() {
		return ctrl
	}

	ctrl = strings.Trim(ctrl, "\r\n\t ")
	ctrl = fmt.Sprintf("%v%v", ctrl, inp.Suffix.TrSilent(langCode)) // &nbsp; is broken anyway

	return ctrl
}

//
//
// GroupHTMLGrid renders a group of inputs to GroupHTMLGrid
func (q QuestionnaireT) GroupHTMLGrid(pageIdx, grpIdx int) string {

	wCSS := &strings.Builder{}
	gr := q.Pages[pageIdx].Groups[grpIdx]

	//
	//
	if gr.Style == nil {
		gr.Style = css.NewStylesResponsive()
	}
	gr.Style.Desktop.BoxStyle.Display = "grid"
	if gr.Style.Desktop.GridContainerStyle.AutoFlow == "" {
		gr.Style.Desktop.GridContainerStyle.AutoFlow = "row"
	}
	if gr.Style.Desktop.GridContainerStyle.TemplateColumns == "" {
		gr.Style.Desktop.GridContainerStyle.TemplateColumns = strings.Repeat("1fr ", gr.Cols)
	}
	gridContainerClass := fmt.Sprintf("pg%02v-grp%02v", pageIdx, grpIdx)
	fmt.Fprint(wCSS, gr.Style.CSS(gridContainerClass))

	//
	//
	wInner := &strings.Builder{} // inside the container

	for inpIdx, inp := range gr.Inputs {
		if inp.Type == "composit-scalar" {
			continue
		}
		if inp.Type == "composit" {
			continue
		}

		if inp.Style == nil {
			inp.Style = css.NewStylesResponsive()

			// input wrapper is item      to group
			inp.Style.Desktop.GridItemStyle.Col = fmt.Sprintf("auto / span %v", inp.ColSpanLabel+inp.ColSpanControl)

			// input wrapper is container to label and control
			inp.Style.Desktop.BoxStyle.Display = "grid"
			inp.Style.Desktop.GridContainerStyle.AutoFlow = "row"
			inp.Style.Desktop.GridContainerStyle.TemplateColumns = strings.Repeat("1fr ", inp.ColSpanLabel+inp.ColSpanControl)
		}
		gridItemClass := fmt.Sprintf("pg%02v-grp%02v-inp%02v", pageIdx, grpIdx, inpIdx)
		fmt.Fprint(wCSS, inp.Style.CSS(gridItemClass))

		wInp := &strings.Builder{} // label and control of input

		{

			if inp.ColSpanLabel > 0 {
				wLbl := &strings.Builder{}
				if inp.StyleLbl == nil {
					inp.StyleLbl = css.NewStylesResponsive()
					inp.StyleLbl.Desktop.GridItemStyle.Col = fmt.Sprintf("auto / span %v", inp.ColSpanLabel)
					inp.StyleLbl.Desktop.GridItemStyle.AlignSelf = "center"
					// inp.StyleLbl.Desktop.GridItemStyle.Order = 2  // label post control
				}

				lblClass := fmt.Sprintf("pg%02v-grp%02v-inp%02v-lbl", pageIdx, grpIdx, inpIdx)
				fmt.Fprint(wCSS, inp.StyleLbl.CSS(lblClass))

				inp.labelDescription(wLbl, q.LangCode)
				if inp.Type == "textblock" {
					divWrap(wInp, lblClass+" grid-item-lvl-2", "", wLbl.String())
				} else {
					wrap(wInp, "label", inp.Name, lblClass+" grid-item-lvl-2", "", wLbl.String())
				}
			}

			if inp.ColSpanControl > 0 {
				wCtl := &strings.Builder{}
				if inp.StyleCtl == nil {
					inp.StyleCtl = css.NewStylesResponsive()
					inp.StyleCtl.Desktop.GridItemStyle.Col = fmt.Sprintf("auto / span %v", inp.ColSpanControl)
					inp.StyleCtl.Desktop.GridItemStyle.AlignSelf = "center"
					inp.StyleCtl.Desktop.TextStyle.WhiteSpace = "nowrap" // prevent suffix from being wrapped
				}
				ctlClass := fmt.Sprintf("pg%02v-grp%02v-inp%02v-ctl", pageIdx, grpIdx, inpIdx)
				fmt.Fprint(wCSS, inp.StyleCtl.CSS(ctlClass))

				fmt.Fprint(wCtl, q.InputHTMLGrid(pageIdx, grpIdx, inpIdx))
				divWrap(wInp, ctlClass+" grid-item-lvl-2", "", wCtl.String())
			}

		}

		//
		divWrap(wInner, gridItemClass+" grid-item-lvl-1", "", wInp.String())
	}

	//
	//
	wContainer := &strings.Builder{}
	divWrap(wContainer, gridContainerClass+" grid-container", "", wInner.String())

	w := &strings.Builder{}
	fmt.Fprint(w, css.StyleTag(wCSS.String()))
	fmt.Fprint(w, wContainer.String())
	return w.String()

}

// InputHTMLGrid renders an input to HTML
func (q QuestionnaireT) InputHTMLGrid(pageIdx, grpIdx, inpIdx int) string {

	gr := q.Pages[pageIdx].Groups[grpIdx]
	inp := *q.Pages[pageIdx].Groups[grpIdx].Inputs[inpIdx]
	nm := inp.Name

	switch inp.Type {

	case "textblock":
		return ""

	case "button":
		lbl := fmt.Sprintf("<button type='submit' name='%v' value='%v' class='%v' accesskey='%v'><b>%v</b> %v</button>\n",
			inp.Name, inp.Response, inp.CSSControl, inp.AccessKey,
			inp.Label.TrSilent(q.LangCode), inp.Desc.TrSilent(q.LangCode),
		)
		lbl = tdDummy(inp.HAlignControl, colWidth(inp.ColSpanControl, gr.Cols), lbl)
		return lbl

	case "text", "number", "hidden", "textarea", "checkbox", "dropdown":

		ctrl := ""
		val := inp.Response

		checked := ""
		if inp.Type == "checkbox" {
			if val == ValSet {
				checked = "checked=\"checked\""
			}
			val = ValSet
		}

		if inp.Type == "textarea" {
			width := ""
			colsRows := fmt.Sprintf(" cols='%v' rows='1' ", inp.MaxChars+1)
			if inp.MaxChars > 80 {
				colsRows = fmt.Sprintf(" cols='80' rows='%v' ", inp.MaxChars/80+1)
				// width = fmt.Sprintf("width: %vem;", int(float64(80)*1.05))
				width = "width: 98%;"
			}
			ctrl += fmt.Sprintf("<textarea        name='%v' id='%v' title='%v %v' class='%v' style='%v' MAXLENGTH='%v' %v>%v</textarea>\n",
				nm, nm, inp.Label.TrSilent(q.LangCode), inp.Desc.TrSilent(q.LangCode), inp.CSSControl, width, inp.MaxChars, colsRows, val)

		} else if inp.Type == "dropdown" {

			// i.DD = &DropdownT{}
			inp.DD.SetName(inp.Name)
			inp.DD.LC = q.LangCode
			inp.DD.SetTitle(inp.Label.TrSilent(q.LangCode) + " " + inp.Desc.TrSilent(q.LangCode))
			inp.DD.Select(inp.Response)
			inp.DD.SetAttr("class", inp.CSSControl)

			sort.Sort(inp.DD)

			ctrl += inp.DD.RenderStr()

		} else {
			// input
			inputMode := ""
			if inp.Type == "number" {
				if inp.Step != 0 {
					if inp.Step >= 1 {
						inputMode = fmt.Sprintf(" step='%.0f'  ", inp.Step)
					} else {
						prec := int(math.Log10(1 / inp.Step))
						f := fmt.Sprintf(" step='%%.%vf'  ", prec)
						inputMode = fmt.Sprintf(f, inp.Step)
					}
				}
			}
			ctrl += fmt.Sprintf(
				`<input type='%v'  %v  name='%v' id='%v' title='%v %v' 
				class='%v' style='width:%vrem'  size='%v' maxlength=%v min='%v' max='%v'  %v  value='%v' />
				`,
				inp.Type, inputMode,
				nm, nm, inp.Label.TrSilent(q.LangCode), inp.Desc.TrSilent(q.LangCode),
				inp.CSSControl, fmt.Sprintf("%.2f", float32(inp.MaxChars)*0.65), inp.MaxChars, inp.MaxChars, inp.Min, inp.Max, checked, val)
		}

		// the checkbox "empty catcher" must follow *after* the actual checkbox input,
		// since http.Form.Get() fetches the first value.
		if inp.Type == "checkbox" {
			ctrl += fmt.Sprintf("<input type='hidden' name='%v' id='%v_hidd' value='0' />\n", nm, nm)
		}

		// append suffix
		ctrl = inp.ShortSuffix(ctrl, q.LangCode)

		// append error message
		if inp.ErrMsg.Set() {
			ctrl += fmt.Sprintf("<span class='go-quest-label %v' >%v</span>\n", inp.CSSLabel, inp.ErrMsg.TrSilent(q.LangCode))
		}

		ctrl = tdDummy(inp.HAlignControl, colWidth(inp.ColSpanControl, gr.Cols), ctrl)

		return ctrl

	case "radiogroup", "checkboxgroup":
		ctrl := ""
		innerType := "radio"
		if inp.Type == "checkboxgroup" {
			innerType = "checkbox"
		}

		for _, rad := range inp.Radios {
			one := ""
			checked := ""
			if inp.Response == rad.Val {
				checked = "checked=\"checked\""
			}

			radio := fmt.Sprintf(
				// 2021-01 - id must be unique
				"<input type='%v' name='%v' id-disabled='%v' title='%v %v' class='%v' value='%v' %v />\n",
				innerType, nm, nm, inp.Label.TrSilent(q.LangCode), inp.Desc.TrSilent(q.LangCode),
				inp.CSSControl,
				rad.Val, checked,
			)

			lbl := ""
			if rad.Label.Empty() {
				one = radio
			} else {

				if rad.HAlign == HLeft {
					lbl = fmt.Sprintf(
						"<span class=' vert-correct-left-right' >%v</span>",
						rad.Label.Tr(q.LangCode),
					)
					radio = strings.Trim(radio, "\r\n\t ")
					one = fmt.Sprintf("%v%v", lbl, radio)
				}

				if rad.HAlign == HCenter {
					// no i.CSSLabel to prevent left margins/paddings to uncenter
					lbl = fmt.Sprintf(
						"<span class=' vert-correct-center '>%v</span>\n",
						rad.Label.Tr(q.LangCode),
					)
					lbl += vspacer0
					one = lbl + radio
				}

				if rad.HAlign == HRight {
					lbl = fmt.Sprintf(
						"<span class=' vert-correct-left-right'>%v</span>",
						rad.Label.Tr(q.LangCode),
					)
					radio = strings.Trim(radio, "\r\n\t ")
					one = fmt.Sprintf("%v%v", radio, lbl)
				}

			}

			cssNoLeft := inp.CSSLabel
			if rad.HAlign == HCenter {
				cssNoLeft = strings.Replace(inp.CSSLabel, "special-input-left-padding", "", -1)
			}
			one = fmt.Sprintf("<span class='go-quest-label %v'>%v</span>\n", cssNoLeft, one)

			ctrl += one

		}
		// The checkbox "empty catcher" must follow *after* the actual checkbox input,
		// since golang http.Form.Get() fetches the *first* value.
		//
		// The radio "empty catcher" becomes necessary,
		// if no radio was selected by the participant;
		// but a "must..." validation rule is registered
		if innerType == "radio" || innerType == "checkbox" {
			ctrl += fmt.Sprintf("<input type='hidden' name='%v' id='%v_hidd' value='%v' />\n",
				nm, nm, valEmpty)
		}

		// append suffix
		ctrl = inp.ShortSuffix(ctrl, q.LangCode)

		if inp.ErrMsg.Set() {
			ctrl += fmt.Sprintf("<div class='error errorblock' >%v</div>\n", inp.ErrMsg.TrSilent(q.LangCode)) // ugly layout - but radiogroup and checkboxgroup won't have validation errors anyway
		}

		return ctrl

	case "dynamic":
		return fmt.Sprintf("<span class='go-quest-label %v'>%v</span>\n", inp.CSSLabel, inp.Label.Tr(q.LangCode))

	case "composit", "composit-scalar":
		// rendered at group level -  rendered by composit
		return ""

	default:
		return fmt.Sprintf("input %v: unknown type '%v'  - allowed are %v\n", nm, inp.Type, implementedTypes)
	}

}