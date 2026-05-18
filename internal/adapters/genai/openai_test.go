package genai

import "testing"

func TestOutputTextReadsResponsesOutput(t *testing.T) {
	body := []byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"{\"executive_summary\":\"summary\",\"probable_cause\":\"cause\",\"blast_radius\":\"radius\",\"recommended_actions\":[\"patch\"],\"servicenow_work_notes\":[\"note\"]}"}]}]}`)
	text, err := outputText(body)
	if err != nil {
		t.Fatal(err)
	}
	if text == "" || text[0] != '{' {
		t.Fatalf("expected JSON text, got %q", text)
	}
}

func TestOutputTextReadsConvenienceField(t *testing.T) {
	text, err := outputText([]byte(`{"output_text":"{\"executive_summary\":\"summary\"}"}`))
	if err != nil {
		t.Fatal(err)
	}
	if text != `{"executive_summary":"summary"}` {
		t.Fatalf("unexpected text: %s", text)
	}
}
