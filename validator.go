package validator

import (
	"fmt"
	"sync"
)

var BuiltInValidators = []Validator{
	&MapValidator{},
	&SliceValidator{},
	&StringValidator{},
	&UintValidator{},
	&IntValidator{},
}

type Validator interface {
	// name and aliases of validator
	// we will register validator to validator set by these names
	Names() []string
	// validate value
	Validate(v interface{}) error
	// create new instance
	New(rule *Rule) (Validator, error)
	// stringify validator rule
	String() string
}

func NewValidatorSet(validators ...Validator) ValidatorSet {
	validatorSet := ValidatorSet{}
	for _, validator := range validators {
		for _, name := range validator.Names() {
			validatorSet[name] = validator
		}
	}
	return validatorSet
}

type ValidatorSet map[string]Validator

func (validatorSet ValidatorSet) Get(name string) (Validator, bool) {
	validator, ok := validatorSet[name]
	return validator, ok
}

func (validatorSet *ValidatorSet) Compile(rule []byte) (Validator, error) {
	r, err := ParseRule(rule)
	if err != nil {
		return nil, err
	}
	validator, ok := validatorSet.Get(r.Name)
	if !ok {
		return nil, fmt.Errorf("%s not match any validator", r.Name)
	}
	return validator.New(r)
}

func NewValidatorFactory(validators ...Validator) *ValidatorFactory {
	return &ValidatorFactory{
		validators: NewValidatorSet(validators...),
	}
}

type ValidatorFactory struct {
	validators ValidatorSet
	cache      sync.Map
}

func (f *ValidatorFactory) Compile(rule string) (Validator, error) {
	if v, ok := f.cache.Load(rule); ok {
		return v.(Validator), nil
	}
	validator, err := f.validators.Compile([]byte(rule))
	if err != nil {
		return nil, err
	}
	f.cache.Store(rule, validator)
	return validator, nil
}
