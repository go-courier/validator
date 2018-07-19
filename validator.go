package validator

import (
	"fmt"
	"sync"

	"github.com/go-courier/reflectx/typesutil"

	"github.com/go-courier/validator/rules"
)

func MustParseRuleStringWithType(ruleStr string, typ typesutil.Type) *Rule {
	r, err := ParseRuleWithType([]byte(ruleStr), typ)
	if err != nil {
		panic(err)
	}
	return r
}

func ParseRuleWithType(ruleBytes []byte, typ typesutil.Type) (*Rule, error) {
	r, err := rules.ParseRule(ruleBytes)
	if err != nil {
		return nil, err
	}
	return &Rule{
		Type: typ,
		Rule: r,
	}, nil
}

func (r *Rule) String() string {
	return typesutil.FullTypeName(r.Type) + string(r.Rule.Bytes())
}

type Rule struct {
	*rules.Rule
	Type typesutil.Type
}

type RuleProcessor func(rule *Rule)

// mgr for compiling validator
type ValidatorMgr interface {
	// compile rule string to validator
	Compile([]byte, typesutil.Type, RuleProcessor) (Validator, error)
}

var ValidatorMgrDefault = NewValidatorFactory()

type ValidatorCreator interface {
	// name and aliases of validator
	// we will register validator to validator set by these names
	Names() []string
	// create new instance
	New(*Rule, ValidatorMgr) (Validator, error)
}

type Validator interface {
	// validate value
	Validate(v interface{}) error
	// stringify validator rule
	String() string
}

func NewValidatorFactory() *ValidatorFactory {
	return &ValidatorFactory{
		validatorSet: map[string]ValidatorCreator{},
	}
}

type ValidatorFactory struct {
	validatorSet map[string]ValidatorCreator
	cache        sync.Map
}

func (f *ValidatorFactory) ResetCache() {
	f.cache = sync.Map{}
}

func (f *ValidatorFactory) Register(validators ...ValidatorCreator) {
	for i := range validators {
		validator := validators[i]
		for _, name := range validator.Names() {
			f.validatorSet[name] = validator
		}
	}
}

func (f *ValidatorFactory) MustCompile(rule []byte, typ typesutil.Type, ruleProcessor RuleProcessor) Validator {
	v, err := f.Compile(rule, typ, ruleProcessor)
	if err != nil {
		panic(err)
	}
	return v
}

func (f *ValidatorFactory) Compile(ruleBytes []byte, typ typesutil.Type, ruleProcessor RuleProcessor) (Validator, error) {
	if len(ruleBytes) == 0 {
		return nil, nil
	}

	rule, err := ParseRuleWithType(ruleBytes, typ)
	if err != nil {
		return nil, err
	}

	if ruleProcessor != nil {
		ruleProcessor(rule)
	}

	key := rule.String()
	if v, ok := f.cache.Load(key); ok {
		return v.(Validator), nil
	}

	validatorCreator, ok := f.validatorSet[rule.Name]
	if !ok {
		return nil, fmt.Errorf("%s not match any validator", rule.Name)
	}

	normalizeValidator := NewValidatorLoader(validatorCreator)

	validator, err := normalizeValidator.New(rule, f)
	if err != nil {
		return nil, err
	}

	f.cache.Store(key, validator)
	return validator, nil
}
