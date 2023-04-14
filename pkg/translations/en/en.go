package en

import (
	"fmt"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/en"
)

// translation is a translation to register
type translation struct {
	tag      string // the tag name of the field
	key      string // the key to use for the translation
	text     string // the text to use for the translation
	override bool   // override the default translation
}

// translations is a list of translations to register
var translations = []translation{
	{tag: "dir", key: "dir", text: "{0} is not a valid directory: {1}", override: true},
	{tag: "file", key: "file", text: "{0} is not a valid file: {1}", override: true},
}

// RegisterDefaultTranslations registers the default translations and custom translations
func RegisterDefaultTranslations(v *validator.Validate, trans ut.Translator) (err error) {
	// register default translations
	err = en.RegisterDefaultTranslations(v, trans)
	if err != nil {
		return fmt.Errorf("failed to register default translations: %w", err)
	}

	// register custom translations
	for _, t := range translations {
		err = v.RegisterTranslation(t.tag, trans, func(ut ut.Translator) error {
			return ut.Add(t.key, t.text, t.override)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			// get the value of the field
			value, ok := fe.Value().(string)

			// if the value is not a string, we can't use it
			if !ok {
				value = fmt.Sprintf("%v", fe.Value())
			}

			// translate the error
			t, _ := ut.T(t.key, fe.Field(), value)
			return t
		})

		// if there is an error, return it immediately
		if err != nil {
			return fmt.Errorf("failed to register translation %s: %w", t.tag, err)
		}
	}

	return nil
}
