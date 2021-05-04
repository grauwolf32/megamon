package stage

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"sync"

	"github.com/megamon/core/leaks/fragment"
	"github.com/megamon/core/leaks/models"
)

//Fragmentize : calculate text fragments and process it
func Fragmentize(ctx context.Context, stage Interface, nWorkers int) {
	var wg sync.WaitGroup
	textQueue := make(chan ReportText, MAXCHANCAP)
	fragmentQueue := make(chan models.TextFragment, MAXCHANCAP)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(textQueue)

		reportTexts, err := stage.GetTextsToProcess()
		if err != nil {
			logErr(err)
			return
		}

		for _, reportText := range reportTexts {
			textQueue <- reportText
		}

		return
	}()
	manager := stage.GetDBManager()
	keywords, err := manager.SelectAllKeywords()

	if err != nil {
		logErr(err)
		return
	}

	rules, err := manager.SelectAllRules()

	if err != nil {
		logErr(err)
		return
	}

	for i := 0; i < nWorkers; i++ {
		go fragmenter(ctx, textQueue, fragmentQueue, &keywords, &rules)
	}

	go func() {
		for textFragment := range fragmentQueue {
			err := stage.ProcessTextFragment(textFragment)
			if err != nil {
				logErr(err)
			}
		}
	}()

	wg.Wait()
	close(fragmentQueue)
	return
}

func buildTextFragment(reportText ReportText, context fragment.Fragment, keywords *[]fragment.Fragment, RejectID int) (textFragment models.TextFragment, err error) {
	textFragment.RejectID = RejectID
	textFragment.ReportID = reportText.ReportID
	textFragment.Text, err = context.Apply(reportText.Text)
	if err != nil {
		return
	}

	textFragment.ShaHash = sha1.Sum([]byte(textFragment.Text))
	for _, keyword := range *keywords {
		textFragment.Keywords = append(textFragment.Keywords, []int{keyword.Offset - context.Offset, keyword.Length})
	}

	return
}

//checkKeywordFragment : checks if fragment with keyword matches the expression
//If we throw the keyword from fragment & it still matches, then that is false positive
func checkKeywordFragment(rules *[]models.RejectRule, frag, keyword fragment.Fragment, text string) (match bool, id int, err error) {
	var builder strings.Builder
	fragmentText, err := frag.Apply(text)

	if err != nil {
		return
	}

	if keyword.Offset < frag.Offset || keyword.Offset+keyword.Length > frag.Offset+frag.Length {
		err = fmt.Errorf("keyword is out of the fragment")
		return
	}

	builder.WriteString(text[frag.Offset:keyword.Offset])
	builder.WriteString(text[keyword.Offset+keyword.Length : frag.Offset+frag.Length])
	stripped := builder.String()

	for _, rule := range *rules {

		if rule.Expr.Match([]byte(fragmentText)) {
			if rule.Expr.Match([]byte(stripped)) {
				continue
			} else {
				return true, id, err
			}
		}
	}

	return false, -1, err
}

func filterKeywordContexts(ctx context.Context, reportText ReportText, keyword string, fragmentQueue chan models.TextFragment, rules *[]models.RejectRule) (keywords, contexts []fragment.Fragment) {
	keywordFragments := fragment.GetKeywordFragments(reportText.Text, keyword)
	checkedFragments := make([]fragment.Fragment, 0, len(keywordFragments))
	kwContexts := make([]fragment.Fragment, 0, len(keywordFragments))

	for _, keyword := range keywordFragments {
		kwContext := fragment.GetKeywordContext(reportText.Text, CONTEXTLEN, keyword)

		match, id, err := checkKeywordFragment(rules, kwContext, keyword, reportText.Text)
		if err != nil {
			logErr(err)
			continue
		}

		if match {
			fragmentKeywords := []fragment.Fragment{{Offset: keyword.Length, Length: keyword.Offset}}
			textFragment, err := buildTextFragment(reportText, kwContext, &fragmentKeywords, id)

			if err != nil {
				logErr(err)
				continue
			}

			select {
			case <-ctx.Done():
				return

			case fragmentQueue <- textFragment:
			}

		} else {
			kwContexts = append(kwContexts, kwContext)
			checkedFragments = append(checkedFragments, keyword)
		}
	}
	return checkedFragments, kwContexts
}

func fragmenter(ctx context.Context, textQueue chan ReportText, fragmentQueue chan models.TextFragment, keywords *[]models.Keyword, rules *[]models.RejectRule) {
	if len(*keywords) == 0 {
		return
	}

	for reportText := range textQueue {
		var mergedContexts []fragment.Fragment
		var mergedKeywords []fragment.Fragment

		for _, keyword := range *keywords {
			fragmentKeywords, fragmentContexts := filterKeywordContexts(ctx, reportText, keyword.Value, fragmentQueue, rules)
			mergedKeywords = fragment.Merge(&mergedKeywords, &fragmentKeywords)
			mergedContexts = fragment.Merge(&mergedContexts, &fragmentContexts)
		}

		mergedContexts = fragment.Join(&mergedContexts, MAXCONTEXTLEN)
		kwInFrags := fragment.GetKeywordsInFragments(mergedKeywords, mergedContexts)

		for id := range kwInFrags {
			keywordIDs := kwInFrags[id]
			context := mergedContexts[id]

			fragKeywords := make([]fragment.Fragment, 0, len(keywordIDs))
			for _, kwID := range keywordIDs {
				fragKeywords = append(fragKeywords, mergedKeywords[kwID])
			}

			textFragment, err := buildTextFragment(reportText, context, &fragKeywords, 0)

			if err != nil {
				logErr(err)
				continue
			}

			select {
			case <-ctx.Done():
				return

			case fragmentQueue <- textFragment:
			}
		}
	}

	return
}
