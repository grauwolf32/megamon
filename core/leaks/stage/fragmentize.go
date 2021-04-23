package stage

import (
	"context"
	"crypto/sha1"
	"sync"

	"github.com/megamon/core/config"
	"github.com/megamon/core/leaks/fragment"
	"github.com/megamon/core/leaks/helpers"
)

//Fragmentize : calculate text fragments and process it
func Fragmentize(ctx context.Context, stage *Interface, nWorkers int) {
	var wg sync.WaitGroup
	textQueue := make(chan ReportText, MAXCHANCAP)
	fragmentQueue := make(chan helpers.TextFragment, MAXCHANCAP)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(textQueue)

		reportTexts, err := (*stage).GetTextsToProcess()
		if err != nil {
			logErr(err)
			return
		}

		for _, reportText := range reportTexts {
			textQueue <- reportText
		}

		return
	}()

	keywords := config.Settings.LeakGlobals.Keywords

	for i := 0; i < nWorkers; i++ {
		go fragmenter(ctx, textQueue, fragmentQueue, keywords)
	}

	go func() {
		for textFragment := range fragmentQueue {
			err := (*stage).ProcessTextFragment(textFragment)
			if err != nil {
				logErr(err)
			}
		}
	}()

	wg.Wait()
	close(fragmentQueue)
	return
}

func fragmenter(ctx context.Context, textQueue chan ReportText, fragmentQueue chan helpers.TextFragment, keywords []string) {
	if len(keywords) == 0 {
		return
	}

	for reportText := range textQueue {
		var kwContextFragments [][]fragment.Fragment
		var kwFragments [][]fragment.Fragment

		for _, keyword := range keywords {
			kwFragment := fragment.GetKeywordFragments(reportText.Text, keyword)
			kwFragments = append(kwFragments, kwFragment)

			kwFragmentForContext := make([]fragment.Fragment, len(kwFragment))
			copy(kwFragmentForContext, kwFragment)

			kwContextFragment := fragment.GetKeywordContext(reportText.Text, CONTEXTLEN, kwFragmentForContext)
			kwContextFragments = append(kwContextFragments, kwContextFragment)
		}

		mergedContexts := fragment.MergeFragments(kwContextFragments, MAXCONTEXTLEN)
		mergedKeywords := fragment.MergeSort(kwFragments)
		kwInFrags := fragment.GetKeywordsInFragments(mergedKeywords, mergedContexts)

		for id := range kwInFrags {
			var textFragment helpers.TextFragment
			keywordIDs := kwInFrags[id]

			frag := mergedContexts[id]
			fragText, err := frag.Apply(reportText.Text)

			if err != nil {
				logErr(err)
				continue
			}

			textFragment.Text = fragText
			textFragment.ShaHash = sha1.Sum([]byte(fragText))
			textFragment.ReportID = reportText.ReportID

			for _, kwID := range keywordIDs {
				kw := mergedKeywords[kwID]
				textFragment.Keywords = append(textFragment.Keywords, []int{kw.Offset - frag.Offset, kw.Length})
			}

			fragmentQueue <- textFragment
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	return
}
