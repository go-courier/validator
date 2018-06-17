package validator

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/go-courier/validator/rules"
)

var BuiltInValidators = []ValidatorCreator{
	&StringValidator{},
	&UintValidator{},
	&IntValidator{},
	&FloatValidator{},

	&StructValidator{},
	&MapValidator{},
	&SliceValidator{},
}

type ValidatorMgr interface {
	Compile(rule []byte, tpe reflect.Type, processor RuleProcessor) (Validator, error)
}

type RuleProcessor func(rule *rules.Rule)

type ValidatorCreator interface {
	// name and aliases of validator
	// we will register validator to validator set by these names
	Names() []string
	// create new instance
	New(rule *rules.Rule, tpe reflect.Type, validateMgr ValidatorMgr) (Validator, error)
}

type Validator interface {
	// validate value
	Validate(v interface{}) error
	// stringify validator rule
	String() string
}

func NewValidatorFactory(validators ...ValidatorCreator) *ValidatorFactory {
	validatorSet := map[string]ValidatorCreator{}
	for _, validator := range validators {
		for _, name := range validator.Names() {
			validatorSet[name] = validator
		}
	}

	return &ValidatorFactory{
		validatorSet: validatorSet,
	}
}

type ValidatorFactory struct {
	validatorSet map[string]ValidatorCreator
	cache        sync.Map
}

func (f *ValidatorFactory) MustCompile(rule []byte, tpe reflect.Type, ruleProcessor RuleProcessor) Validator {
	v, err := f.Compile(rule, tpe, ruleProcessor)
	if err != nil {
		panic(err)
	}
	return v
}

func (f *ValidatorFactory) Compile(rule []byte, tpe reflect.Type, ruleProcessor RuleProcessor) (Validator, error) {
	if len(rule) == 0 {
		return nil, nil
	}

	r, err := rules.ParseRule(rule)
	if err != nil {
		return nil, err
	}
	if ruleProcessor != nil {
		ruleProcessor(r)
	}

	key := tpe.String() + string(r.Bytes())
	if v, ok := f.cache.Load(key); ok {
		return v.(Validator), nil
	}

	validatorCreator, ok := f.validatorSet[r.Name]
	if !ok {
		return nil, fmt.Errorf("%s not match any validator", r.Name)
	}

	normalizeValidator := NewValidatorLoader(validatorCreator)

	validator, err := normalizeValidator.New(r, tpe, f)
	if err != nil {
		return nil, err
	}

	f.cache.Store(key, validator)
	return validator, nil
}
