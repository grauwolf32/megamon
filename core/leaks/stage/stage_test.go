package stage

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/utils"
)

func TestMain(m *testing.M) {
	utils.InitLoggers("test.log")

	retCode := m.Run()
	err := os.Remove("test.log")
	if err != nil {
		fmt.Println("Unable to remove test.log")
		fmt.Println(err.Error())
	}
	os.Exit(retCode)
}

func TestFragmener(t *testing.T) {
	text := "this test is made for the fragmenter function. To test its functioning."
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "this test is made for the fragmenter function. To test its functioning."
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	text += "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	ctx := context.Background()
	textQueue := make(chan ReportText, 10)
	textQueue <- ReportText{ReportID: 1, Text: text}
	close(textQueue)

	fragmentQueue := make(chan models.TextFragment, 10)
	keywords := []string{"test"}
	fragmenter(ctx, textQueue, fragmentQueue, keywords)
	close(fragmentQueue)

	result := make([]models.TextFragment, 0, 10)

	for frag := range fragmentQueue {
		result = append(result, frag)
	}

	if len(result) != 2 {
		t.Errorf("Expected number of fragments: %d; got: %d", 2, len(result))
		return
	}

	for _, kw := range result[0].Keywords {
		if kw[1] != 4 {
			t.Errorf("Got incorrect length of fragment. Expected len('test') == 4; got: %d\n %v", kw[1], result[0].Keywords)
			return
		}
	}

	for _, kw := range result[1].Keywords {
		if kw[1] != 4 {
			t.Errorf("Got incorrect length of fragment. Expected len('test') == 4; got: %d\n %v", kw[1], result[1].Keywords)
			return
		}
	}

	return
}
