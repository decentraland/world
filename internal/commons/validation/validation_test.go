package validation

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type testStruct struct {
	Required string `validate:"required"`
}

func TestDefaultValidator(t *testing.T) {
	v, err := DefaultValidator()
	require.NoError(t, err)

	validateNull := &testStruct{}
	err = v.ValidateStruct(validateNull)
	assert.NotNil(t, err)
	assert.Equal(t, "Required field is required.", err.Error())

	validateEmpty := &testStruct{Required: ""}
	err = v.ValidateStruct(validateEmpty)
	assert.NotNil(t, err)
	assert.Equal(t, "Required field is required.", err.Error())
}


func TestValidatorTranslations(t *testing.T) {
	messages := map[string]string{}
	messages["required"] = "you should set the {0} field!!!!"
	v, err := WithMessages(messages)
	require.NoError(t, err)

	validateEmpty := &testStruct{Required: ""}
	err = v.ValidateStruct(validateEmpty)
	assert.NotNil(t, err)
	assert.Equal(t, "you should set the Required field!!!!", err.Error())
}
