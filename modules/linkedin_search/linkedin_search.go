package linkedin_search

import (
	"github.com/graniet/operative-framework/session"
	"github.com/graniet/go-pretty/table"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"os"
	"strconv"
)

type LinkedinSearchModule struct{
	session.SessionModule
	sess *session.Session
	Stream *session.Stream
	Export []LinkedinSearchExport
}

type LinkedinSearchExport struct{
	Name string `json:"name"`
	Work string `json:"work"`
	Link string `json:"link"`
}

func PushLinkedinSearchModule(s *session.Session) *LinkedinSearchModule{
	mod := LinkedinSearchModule{
		sess: s,
		Stream: &s.Stream,
	}

	mod.CreateNewParam("TARGET", "Enterprise Name", "", true, session.STRING)
	mod.CreateNewParam("limit", "Limit search", "10", false, session.STRING)
	return &mod
}

func (module *LinkedinSearchModule) Name() string{
	return "linkedin_search"
}

func (module *LinkedinSearchModule) Description() string{
	return "Search employee on selected enterprise with Linkedin"
}

func (module *LinkedinSearchModule) Author() string{
	return "Tristan Granier"
}

func (module *LinkedinSearchModule) GetType() string{
	return "enterprise"
}

func (module *LinkedinSearchModule) GetInformation() session.ModuleInformation{
	information := session.ModuleInformation{
		Name: module.Name(),
		Description: module.Description(),
		Author: module.Author(),
		Type: module.GetType(),
		Parameters: module.Parameters,
	}
	return information
}

func (module *LinkedinSearchModule) GetExport() interface{}{
	return module.Export
}

func (module *LinkedinSearchModule) Start(){
	paramEnterprise, _ := module.GetParameter("TARGET")
	target, err := module.sess.GetTarget(paramEnterprise.Value)
	if err != nil{
		module.sess.Stream.Error(err.Error())
		return
	}

	if target.GetType() != module.GetType(){
		module.Stream.Error("Target with type '"+target.GetType()+"' isn't valid module need '"+module.GetType()+"' type.")
		return
	}

	paramLimit, _ := module.GetParameter("limit")
	url := "https://encrypted.google.com/search?num=" + paramLimit.Value + "&start=0&hl=en&q=site:linkedin.com/in+" + target.GetName()
	res, err := http.Get(url)
	if err != nil {
		module.Stream.Error("Argument 'URL' can't be reached.")
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		module.Stream.Error("Argument 'URL' can't be reached.")
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		module.Stream.Error("A error as been occurred with a target.")
		return
	}

	t := module.Stream.GenerateTable()
	t.SetOutputMirror(os.Stdout)
	t.SetAllowedColumnLengths([]int{0, 30,})
	t.AppendHeader(table.Row{"Name", "Work", "Link"})

	resultFound := 0
	doc.Find(".g").Each(func(i int, s *goquery.Selection) {
		line := s.Find("h3").Text()
		line = strings.Replace(line, "| LinkedIn", "", 1)
		line = strings.Replace(line, "LinkedIn", "", 1)
		line = strings.Replace(line, "on LinkedIn", "", 1)
		if strings.Contains(line,"-") && len(strings.Split(strings.TrimSpace(line), "-")) > 1{
			name := strings.Split(strings.TrimSpace(line), "-")[0]
			work := strings.Split(strings.TrimSpace(line), "-")[1]
			link := s.Find("cite").Text()
			separator := target.GetSeparator()
			t.AppendRow([]interface{}{name, work, link})
			result := LinkedinSearchExport{
				Name: name,
				Work: work,
				Link: link,
			}
			module.Export = append(module.Export, result)
			target.Save(module, session.TargetResults{
				Header: "Name" + separator + "Work" + separator + "Link",
				Value: name + separator + work + separator + link,
			})
			resultFound = resultFound + 1

		}
	})
	module.Stream.Render(t)
	module.Stream.Success(strconv.Itoa(resultFound) + " employee(s) found with a Linkedin search.")
}