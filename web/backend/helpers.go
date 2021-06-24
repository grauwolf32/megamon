package backend

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/megamon/core/leaks/fragment"
	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/leaks/stage"
	"github.com/megamon/core/utils"
	"gopkg.in/yaml.v2"
)

func stub(ctx echo.Context) (err error) {
	return
}

//Context : context with db and other stuff
type Context struct {
	backend *Backend
	echo.Context
}

func getFragments(ctx echo.Context) (err error) {
	fragmentType := ctx.Param("datatype")
	rejectID, err := strconv.Atoi(ctx.Param("status"))
	if err != nil {
		return ctx.String(400, err.Error())
	}

	extensions := make([]string, 3)
	valid, err := validate(fragmentType)
	if err != nil {
		return ctx.String(400, err.Error())
	}

	if !valid {
		return ctx.String(400, "wrong data type")
	}

	filterQuery := "AND type='" + fragmentType + "' "
	extensions = append(extensions, filterQuery)

	limitParam := ctx.FormValue("limit")
	if limitParam != "" {
		_, err := strconv.Atoi(limitParam)
		if err != nil {
			return ctx.String(400, err.Error())
		}
		limitQuery := "LIMIT " + limitParam + " "
		extensions = append(extensions, limitQuery)
	}

	offsetParam := ctx.FormValue("offset")
	if offsetParam != "" {
		_, err := strconv.Atoi(offsetParam)
		if err != nil {
			return ctx.String(400, err.Error())
		}

		offsetQuery := "OFFSET " + offsetParam
		extensions = append(extensions, offsetQuery)
	}

	manager := ctx.(Context).backend.DBManager
	fragments, err := manager.SelectTextFragment("reject_id", rejectID, extensions...)

	for i := range fragments {
		alteredKeywords := make([][]int, 0, len(fragments[i].Keywords))
		fragmentKeywords := fragments[i].Keywords
		for _, kwIndices := range fragmentKeywords {
			frag := fragment.Fragment{Offset: kwIndices[0], Length: kwIndices[1]}
			frag.ConvertToRunes(fragments[i].Text)
			alteredKeywords = append(alteredKeywords, []int{frag.Offset, frag.Offset + frag.Length})
		}

		fragments[i].Keywords = alteredKeywords
	}
	if err != nil {
		return ctx.String(500, err.Error())
	}

	reportJSON, err := json.Marshal(fragments)
	if err != nil {
		return ctx.String(500, err.Error())
	}
	return ctx.JSONBlob(200, reportJSON)
}

func getFragmentCount(ctx echo.Context) (err error) {
	fragmentType := ctx.Param("datatype")
	rejectID, err := strconv.Atoi(ctx.Param("status"))
	if err != nil {
		return ctx.String(400, err.Error())
	}

	extensions := make([]string, 3)
	valid, err := validate(fragmentType)
	if err != nil {
		return ctx.String(400, err.Error())
	}

	if !valid {
		return ctx.String(400, "wrong data type")
	}

	filterQuery := "AND type='" + fragmentType + "' "
	extensions = append(extensions, filterQuery)

	manager := ctx.(Context).backend.DBManager
	count, err := manager.CountTextFragments("reject_id", rejectID, extensions...)

	if err != nil {
		return ctx.String(400, err.Error())
	}

	return ctx.JSON(200, struct {
		Count    int    `json:"count"`
		Type     string `json:"type"`
		RejectID int    `json:"reject_id"`
	}{count, fragmentType, rejectID})
}
func getFragmentInfo(ctx echo.Context) (err error) {
	fragID, err := strconv.Atoi(ctx.Param("frag_id"))

	if err != nil {
		return ctx.String(500, err.Error())
	}

	manager := ctx.(Context).backend.DBManager
	frags, err := manager.SelectTextFragment("id", fragID)

	if err != nil {
		return ctx.String(500, err.Error())

	}
	if len(frags) != 1 {
		return ctx.String(500, fmt.Sprintf("There must be exactly one fragment with id: %d; got %d", fragID, len(frags)))
	}

	frag := frags[0]
	reportID := frag.ReportID

	report, err := manager.SelectReportByID(reportID)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	return ctx.JSONBlob(200, report.Data)
}

func markFragment(ctx echo.Context) (err error) {
	fragID, err := strconv.Atoi(ctx.Param("frag_id"))
	if err != nil {
		return ctx.String(500, err.Error())
	}

	rejectID, err := strconv.Atoi(ctx.Param("status"))
	if err != nil {
		return ctx.String(500, err.Error())
	}

	manager := ctx.(Context).backend.DBManager
	frags, err := manager.SelectTextFragment("id", fragID)

	if err != nil {
		return ctx.String(500, err.Error())

	}
	if len(frags) != 1 {
		return ctx.String(500, fmt.Sprintf("There must be exactly one fragment with id: %d; got %d", fragID, len(frags)))
	}

	frag := frags[0]
	reportID := frag.ReportID

	//fmt.Printf("%v %d %d %d\n", frag, fragID, reportID, rejectID)

	err = manager.UpdateTextFragmentRejectID(frag.ID, rejectID)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	if rejectID == models.RULEVERIFIED {
		frags, err := manager.SelectTextFragment("report_id", reportID)
		if err != nil {
			return ctx.String(500, err.Error())
		}

		for _, f := range frags {
			if f.ID != fragID && f.RejectID != models.RULEMANUAL {
				err = manager.UpdateTextFragmentRejectID(f.ID, models.RULEAUTOREMOVED)
				if err != nil {
					return ctx.String(500, err.Error())
				}
			}
		}

		manager.UpdateReportStatus(reportID, stage.VALIDATED)
		return ctx.String(200, "OK")
	}

	count, err := manager.CountTextFragments("reject_id", models.RULENONE)

	if count == 0 {
		manager.UpdateReportStatus(reportID, stage.CLOSED)
	}

	return ctx.String(200, "OK")
}

func getSettings(ctx echo.Context) (err error) {
	settings, err := json.Marshal(utils.Settings)
	if err != nil {
		return ctx.String(500, err.Error())
	}
	return ctx.JSONBlob(200, settings)
}

func updateSettings(ctx echo.Context) (err error) {
	var updated utils.GlobalSettings
	err = ctx.Bind(&updated)
	if err != nil {
		return ctx.String(400, err.Error())
	}

	if updated.AdminCredentials.Password != "" {
		shaHash := sha1.New().Sum([]byte(updated.AdminCredentials.Password))
		utils.Settings.AdminCredentials.Password = fmt.Sprintf("%x", shaHash)
	}

	if updated.AdminCredentials.Username != "" {
		utils.Settings.AdminCredentials.Username = updated.AdminCredentials.Username
	}

	utils.Settings.Github.Langs = updated.Github.Langs
	utils.Settings.Github.Tokens = updated.Github.Tokens

	yamlSettings, err := yaml.Marshal(utils.Settings)

	if err != nil {
		return ctx.String(500, err.Error())
	}

	err = ioutil.WriteFile("../../config/config.yaml", yamlSettings, 0644)
	return ctx.String(200, "OK")
}

func getRegexps(ctx echo.Context) (err error) {
	if err != nil {
		return ctx.String(500, err.Error())
	}

	manager := ctx.(Context).backend.DBManager
	rules, err := manager.SelectAllRules()

	if err != nil {
		return ctx.String(500, err.Error())
	}

	jsonRules, err := json.Marshal(rules)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	return ctx.JSONBlob(200, jsonRules)
}

func delRegexp(ctx echo.Context) (err error) {
	ID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.String(400, err.Error())
	}
	manager := ctx.(Context).backend.DBManager

	err = manager.DeleteRuleByID(ID)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	return ctx.String(200, "OK")
}

func addRegexp(ctx echo.Context) (err error) {
	var rule models.RejectRule
	err = ctx.Bind(&rule)
	if err != nil {
		return ctx.String(400, err.Error())
	}

	rule.Name = "regexp"
	manager := ctx.(Context).backend.DBManager
	ID, err := manager.InsertRule(rule)
	rule.ID = ID

	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	return ctx.JSONBlob(200, ruleJSON)
}

func getKeywords(ctx echo.Context) (err error) {
	manager := ctx.(Context).backend.DBManager
	keywords, err := manager.SelectAllKeywords()
	if err != nil {
		return ctx.String(500, err.Error())
	}

	keywordsJSON, err := json.Marshal(keywords)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	return ctx.JSONBlob(200, keywordsJSON)
}

func delKeyword(ctx echo.Context) (err error) {
	ID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.String(400, err.Error())
	}

	manager := ctx.(Context).backend.DBManager
	err = manager.DeleteKeyword(ID)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	return ctx.String(200, "OK")
}

func addKeyword(ctx echo.Context) (err error) {
	var keyword models.Keyword
	err = ctx.Bind(&keyword)
	if err != nil {
		return ctx.String(500, err.Error())
	}

	manager := ctx.(Context).backend.DBManager
	ID, err := manager.InsertKeyword(keyword.Value, keyword.Type)
	keyword.ID = ID

	keywordJSON, err := json.Marshal(keyword)
	return ctx.JSONBlob(200, keywordJSON)
}

func validate(text string) (bool, error) {
	alphabetic, err := regexp.Compile("[a-zA-Z]+")
	return alphabetic.Match([]byte(text)), err
}
