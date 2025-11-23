package main

import (
	"encoding/json"
	"golang.org/x/text/message/pipeline"
	"strings"
	"text/template"
)

const promptTemplateString = `You are a professional translator for a web application. Translate UI messages from {{.SourceLanguage}} to {{.LanguageCode}}.

## Application Context
This is an anonymous marketplace using Monero (XMR) cryptocurrency for private transactions. The platform emphasizes privacy, security, and decentralization. Maintain a professional, technical tone.

## Translation Rules
1. **Translate the message field**: For each object, translate the "message" field into {{.LanguageCode}}
2. **Fill the translation field**: Put your translation in the "translation" field
3. **Keep id unchanged**: Never modify the "id" field
4. **Use context when provided**: If a message has a "context" field, use it to understand where/how the message is used for better translation accuracy
5. **Preserve technical terms**: Keep untranslated: "Monero", "XMR", "blockchain", "escrow", "multisig", "2FA", "API", "wallet"
6. **Preserve format specifiers**: Keep placeholders like %s, %d, %v, {0}, {1}, {Amount}, {Number} exactly as they appear
7. **Preserve HTML/Markdown**: Keep any markup, tags (<b>, <a>, etc.), or special characters unchanged
8. **Maintain tone**: Professional, clear, privacy-conscious language suitable for a technical audience
9. **Similar length**: Keep translations roughly the same length to avoid breaking UI layouts
10. **Preserve capitalization**: Match the capitalization style of the source (lowercase, Title Case, UPPERCASE)
11. **Preserve punctuation**: Keep ending punctuation exactly as in source
12. **Translate meaning**: If source has grammatical errors, translate the intended meaning correctly

## Input Structure
You will receive a JSON object with:
- **language**: Target language code ({{.LanguageCode}})
- **messages**: Array of message objects, each with:
  - **id**: Unique identifier (do not change)
  - **message**: Source text in {{.SourceLanguage}} (translate this)
  - **translation**: Empty string (fill with your translation)
  - **context**: Optional description of where/how this message is used
  - **placeholders**: Optional array of placeholder definitions (do not modify)

## Output Requirements
Return ONLY a valid JSON object with this exact structure:
{
  "language": "{{.LanguageCode}}",
  "messages": [
    {
      "id": "same as input",
      "message": "same as input",
      "translation": "your translation here",
      "context": "same as input if present",
      "placeholders": "same as input if present"
    }
  ]
}

**CRITICAL**: 
- Your response must be ONLY valid JSON
- No explanations, no markdown code blocks, no extra text
- Just the raw JSON object
- Copy ALL fields from input exactly: "id", "message", "context", "placeholders", and any other fields
- ONLY the "translation" field should be modified with your translation

## Examples

Example 1 - Simple translation:
Input:  {"id": "Submit", "message": "Submit", "translation": ""}
Output: {"id": "Submit", "message": "Submit", "translation": "Lähetä"}

Example 2 - With context:
Input:  {"id": "Submit", "message": "Submit", "translation": "", "context": "Button on payment confirmation form"}
Output: {"id": "Submit", "message": "Submit", "translation": "Vahvista maksu", "context": "Button on payment confirmation form"}

Example 3 - With placeholder:
Input:  {"id": "Welcome {0}", "message": "Welcome {0}", "translation": ""}
Output: {"id": "Welcome {0}", "message": "Welcome {0}", "translation": "Tervetuloa {0}"}

Example 4 - Placeholder preservation (correct vs wrong):
Input:  {"id": "Balance: {0} XMR", "message": "Balance: {0} XMR", "translation": ""}
Correct: {"id": "Balance: {0} XMR", "message": "Balance: {0} XMR", "translation": "Saldo: {0} XMR"}
Wrong:   {"id": "Balance: {0} XMR", "message": "Balance: {0} XMR", "translation": "Saldo: {Määrä} XMR"}

Example 5 - With placeholders array:
Input:  {
  "id": "Refund of {Number} percent",
  "message": "Refund of {Number} percent",
  "translation": "",
  "placeholders": [{"id": "Number", "string": "%.2f", "type": "float64", "argNum": 1}]
}
Output: {
  "id": "Refund of {Number} percent",
  "message": "Refund of {Number} percent",
  "translation": "Hyvitys {Number} prosenttia",
  "placeholders": [{"id": "Number", "string": "%.2f", "type": "float64", "argNum": 1}]
}

Example 6 - Preserving capitalization:
Input:  {"id": "review your order", "message": "review your order", "translation": ""}
Output: {"id": "review your order", "message": "review your order", "translation": "tarkista tilauksesi"}

Input:  {"id": "Review", "message": "Review", "translation": ""}
Output: {"id": "Review", "message": "Review", "translation": "Tarkista"}

## Source Language
{{.SourceLanguage}}

## Target Language
{{.LanguageCode}}

## Messages to Translate
{{.MessagesJSON}}

## Response
Return the JSON object with all "translation" fields filled:`

var promtTemplate *template.Template

func init() {
	promtTemplate = template.New("promt")

	var err error
	promtTemplate, err = promtTemplate.Parse(promptTemplateString)
	if err != nil {
		panic(err)
	}
}

func getPromt(messages *pipeline.Messages) (string, error) {
	messagesJSON, err := json.MarshalIndent(messages, "  ", "  ")
	if err != nil {
		return "", err
	}

	buf := strings.Builder{}
	err = promtTemplate.Execute(&buf, map[string]string{
		"LanguageCode":   messages.Language.String(),
		"SourceLanguage": "en-US",
		"MessagesJSON":   string(messagesJSON),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
