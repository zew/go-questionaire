package qst

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zew/go-questionnaire/cfg"
	"github.com/zew/go-questionnaire/trl"
)

type dynFuncT func(*QuestionnaireT) (string, error)

var dynFuncs = map[string]dynFuncT{
	"RepsonseStatistics":             RepsonseStatistics,
	"PersonalLink":                   PersonalLink,
	"HasEuroQuestion":                ResponseTextHasEuro,
	"FederalStateAboveOrBelowMedian": FederalStateAboveOrBelowMedian,
	"PatLogos":                       PatLogos,
}

// Statistics returns the percentage of
// answers responded to.
// It is helper to RepsonseStatistics().
func (q *QuestionnaireT) Statistics() (int, int, float64) {
	responses := 0
	counter := 0
	radioDoubles := map[string]int{}
	for _, p := range q.Pages {
		for _, gr := range p.Groups {
			for _, i := range gr.Inputs {
				if i.IsLayout() {
					continue
				}
				if radioDoubles[i.Name] > 0 {
					continue
				}
				radioDoubles[i.Name]++
				counter++
				if i.Response != "" && i.Response != "0" {
					responses++
				}
			}
		}

	}
	return responses, counter, 100 * float64(responses) / float64(counter)
}

// RepsonseStatistics returns the percentage of
// answers responded to.
func RepsonseStatistics(q *QuestionnaireT) (string, error) {

	responses, inputs, pct := q.Statistics()
	ct := q.Survey.Deadline
	// ct = ct.Truncate(time.Hour)
	cts := ct.Format("02.01.2006 15:04")
	nextDay := q.Survey.Deadline.Add(24 * time.Hour)
	nextDayS := nextDay.Format("02.01.2006")

	s1 := fmt.Sprintf(cfg.Get().Mp["percentage_answered"].Tr(q.LangCode), responses, inputs, pct)
	s2 := fmt.Sprintf(cfg.Get().Mp["survey_ending"].Tr(q.LangCode), cts, nextDayS)
	ret := s1 + s2
	// log.Print("RepsonseStatistics: " + ret)
	return ret, nil
}

// PersonalLink returns the entry link
func PersonalLink(q *QuestionnaireT) (string, error) {
	closed := q.FinishedEntirely()
	ret := ""
	if closed {
		ret = cfg.Get().Mp["finished_by_participant"].Tr(q.LangCode)
		ret = fmt.Sprintf(ret, q.ClosingTime.Format("02.01.2006 15:04"))
	} else {
		ret = cfg.Get().Mp["review_by_personal_link"].Tr(q.LangCode)
	}
	log.Printf("PersonalLink: closed is %v", closed)
	return ret, nil
}

// ResponseTextHasEuro yields texts => want to keep € - want to have €
func ResponseTextHasEuro(q *QuestionnaireT) (string, error) {

	if q.Attrs == nil {

	}

	attr1, ok1 := q.Attrs["euro-member"]
	attr2, ok2 := q.Attrs["country"] // country of residence - not language - ISO

	if !ok1 || !ok2 {
		return "Question requires known euro-membership and residence code.", nil
	}

	cntry := trl.Countries[attr2]

	cntry["en"] = strings.Replace(cntry["en"], "Czech Republic", "Czechia", -1)
	cntry["de"] = strings.Replace(cntry["de"], "Tschechische Republik", "Tschechien", -1)
	cntry["fr"] = strings.Replace(cntry["fr"], "République tchèque", "Tchéquie", -1)
	cntry["it"] = strings.Replace(cntry["it"], "Repubblica Ceca", "Cechia", -1)

	hl := trl.S{
		"de": "Wirtschaftlicher Nutzen des Euro<br>",
		"en": "Economic benefits of the euro<br>",
		"fr": "Avantages économiques de l'euro<br>",
		"it": "Benefici economici dell'Euro<br>",
	}
	desc := ""
	ret := ""

	if attr1 == "yes" {
		s1 := trl.S{
			"de": fmt.Sprintf("Den Euro in %v als die offizielle Währung zu haben, ist wirtschaftlich vorteilhaft.",
				cntry["de"]),
			"en": fmt.Sprintf("Having the euro in %v as the official currency is economically beneficial.",
				cntry["en"]),
			"fr": fmt.Sprintf("Avoir l'euro en %v comme monnaie officielle est économiquement avantageux.",
				cntry["fr"]),
			"it": fmt.Sprintf("Avere l'Euro come valuta ufficiale in %v è economicamente vantaggioso.",
				cntry["it"]),
		}
		desc = s1[q.LangCode]

	} else {
		s1 := trl.S{
			"de": fmt.Sprintf("Den Euro in %v als offizielle Währung einzuführen, wäre wirtschaftlich vorteilhaft. ",
				cntry["de"]),
			"en": fmt.Sprintf("Introducing the euro in %v as the official currency would be economically beneficial.",
				cntry["en"]),
			"fr": fmt.Sprintf("L'introduction de l'euro dans %v en tant que monnaie officielle serait économiquement avantageuse.",
				cntry["fr"]),
			"it": fmt.Sprintf("Introdurre l'Euro come valuta ufficiale in %v sarebbe economicamente vantaggioso.",
				cntry["it"]),
		}
		desc = s1[q.LangCode]
	}

	ret = fmt.Sprintf("<b> %v </b> %v", hl[q.LangCode], desc)

	return ret, nil

}

// FederalStateAboveOrBelowMedian returns "besser" or "schlechter";
// depending on the user's federal state education ranking
func FederalStateAboveOrBelowMedian(q *QuestionnaireT) (string, error) {

	attr1, ok := q.Attrs["aboveOrBelowMedian"]

	if !ok {
		return "Question requires known euro-membership and residence code.", nil
	}
	return attr1, nil

}

// PatLogos - only for the img src URLs
func PatLogos(q *QuestionnaireT) (string, error) {

	return fmt.Sprintf(
		`
		<div class="uni-logos  logo-imgs-in-content">
			<img src="%v"  style="width:61%%;"  alt=""  >
			<img src="%v"  style="width:33%%;"  alt=""  >
			<img src="%v"  style="width:50%%;"  alt=""  >
			<img src="%v"  style="width:44%%;"  alt=""  >
			<img src="%v"  style="width:28%%;"  alt=""  >
		</div>
		
		<br>
		
		`,
		cfg.Pref("/img/pat/uni-mannheim-wide.png"),
		cfg.Pref("/img/pat/uni-koeln.png"),
		cfg.Pref("/img/pat/uni-muenster.png"),
		cfg.Pref("/img/pat/uni-zurich.png"),
		cfg.Pref("/img/pat/zew.png"),
	), nil

}
