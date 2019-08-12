package alpha

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"reflect"
	"time"

	errors "golang.org/x/xerrors"
	yaml "gopkg.in/yaml.v2"
)

// Constant field keys. Each field has a specific expected type.
const (
	fieldDistribution = "distribution" // Type: Distribution (TODO: Rename this to Distribution?)
	fieldChoices      = "choices"      // Type: Slice
	fieldBounds       = "bounds"       // Type: Slice with len=2
)

type searchSpace struct {
	Seed       *int64                 `yaml:"seed,omitempty"`
	Parameters map[string]interface{} `yaml:"parameters"`
}

// Distribution describes a distribution of values and how they should be sampled.
type Distribution string

const (
	// Choice selects a single choice with uniform probability from a list.
	Choice Distribution = "choice"

	// Integer produces uniformly distributed integers in the range of [bounds[0], bounds[1]).
	Integer Distribution = "uniform-int"

	// LogUniform produces logarithmically distributed floats in the range of [bounds[0], bounds[1]).
	LogUniform Distribution = "log-uniform"

	// Uniform produces uniformly distributed floats in the range of [bounds[0], bounds[1]).
	Uniform Distribution = "uniform"
)

// A ParameterSpace samples a set of named distributions.
type ParameterSpace struct {
	seed   int64
	rand   *rand.Rand
	params map[string]distribution
}

func decodeParameterSpace(r io.Reader) (*ParameterSpace, error) {
	var ss searchSpace
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)
	if err := dec.Decode(&ss); err != nil {
		return nil, err
	}

	var ps ParameterSpace
	if ss.Seed == nil {
		ps.seed = time.Now().Unix()
	} else {
		ps.seed = *ss.Seed
	}

	ps.rand = rand.New(rand.NewSource(ps.seed))
	ps.params = make(map[string]distribution, len(ss.Parameters))
	for key, param := range ss.Parameters {
		d, err := parseDistribution(param)
		if err != nil {
			return nil, errors.Errorf("parameter %q: %w", key, err)
		}
		ps.params[key] = d
	}

	return &ps, nil
}

func parseDistribution(param interface{}) (distribution, error) {
	if reflect.TypeOf(param).Kind() != reflect.Map {
		// All distributions are expressed as a map, so this must be a fixed value.
		return fixedValue{param}, nil
	}

	fields := param.(map[interface{}]interface{})
	d, ok := fields[fieldDistribution]
	if !ok {
		return nil, errors.Errorf("no sampling distribution provided")
	}

	switch dist := Distribution(d.(string)); dist {
	case Choice:
		choices, err := getChoices(fields)
		if err != nil {
			return nil, err
		}
		return newChooseOne(choices)

	case Integer:
		min, max, err := getBoundsInt(fields)
		if err != nil {
			return nil, err
		}
		return newUniformInt(min, max)

	case LogUniform:
		min, max, err := getBoundsFloat(fields)
		if err != nil {
			return nil, err
		}
		return newLogFloat(min, max)

	case Uniform:
		min, max, err := getBoundsFloat(fields)
		if err != nil {
			return nil, err
		}
		return newUniformFloat(min, max)

	default:
		return nil, errors.Errorf("sampling distribution %q is not supported", dist)
	}
}

func getChoices(fields map[interface{}]interface{}) ([]interface{}, error) {
	choices, ok := fields[fieldChoices]
	if !ok || reflect.TypeOf(choices).Kind() != reflect.Slice {
		return nil, errors.Errorf("must specify %q as a list", fieldChoices)
	}

	return choices.([]interface{}), nil
}

func getBoundsInt(fields map[interface{}]interface{}) (min, max int64, err error) {
	field, ok := fields[fieldBounds]
	value := reflect.ValueOf(field)
	if !ok || value.Type().Kind() != reflect.Slice || value.Len() != 2 {
		return 0, 0, errors.Errorf("must specify %q as a list of 2 elements", fieldBounds)
	}

	min, err = getInt(value.Index(0))
	if err != nil {
		return 0, 0, errors.Errorf("%s[0]: %w", fieldBounds, err)
	}
	max, err = getInt(value.Index(1))
	if err != nil {
		return 0, 0, errors.Errorf("%s[1]: %w", fieldBounds, err)
	}

	return min, max, nil
}

func getBoundsFloat(fields map[interface{}]interface{}) (min, max float64, err error) {
	field, ok := fields[fieldBounds]
	value := reflect.ValueOf(field)
	if !ok || value.Type().Kind() != reflect.Slice || value.Len() != 2 {
		return 0, 0, errors.Errorf("must specify %q as a list of 2 elements", fieldBounds)
	}

	min, err = getFloat(value.Index(0))
	if err != nil {
		return 0, 0, errors.Errorf("%s[0]: %w", fieldBounds, err)
	}
	max, err = getFloat(value.Index(1))
	if err != nil {
		return 0, 0, errors.Errorf("%s[1]: %w", fieldBounds, err)
	}

	return min, max, nil
}

func getInt(value reflect.Value) (int64, error) {
	switch value.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i := int64(value.Uint())
		if i < 0 {
			return 0, errors.New("value is too large")
		}
		return i, nil
	case reflect.Interface:
		return getInt(reflect.ValueOf(value.Interface()))
	default:
		fmt.Println(value.Type().Kind())
		return 0, errors.Errorf("value is not an integer: %s", value.String())
	}
}

func getFloat(value reflect.Value) (float64, error) {
	switch value.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint()), nil
	case reflect.Interface:
		return getFloat(reflect.ValueOf(value.Interface()))
	default:
		return 0, errors.Errorf("value is not a number: %s", value.String())
	}
}

// Sample returns a collection of sampled values from a search space.
func (ps *ParameterSpace) Sample() map[string]interface{} {
	// TODO: Iterate by sorted keys to ensure a fixed seed always produces identical results.
	result := make(map[string]interface{}, len(ps.params))
	for key, param := range ps.params {
		result[key] = param.Sample(ps.rand)
	}
	return result
}

type distribution interface {
	Sample(r *rand.Rand) interface{}
}

type fixedValue struct {
	value interface{}
}

func (d fixedValue) Sample(r *rand.Rand) interface{} {
	return d.value
}

type uniformInt struct {
	min, max int64
}

func newUniformInt(min, max int64) (uniformInt, error) {
	if min >= max {
		return uniformInt{}, errors.New("min must be less than max")
	}
	return uniformInt{min, max}, nil
}

func (d uniformInt) Sample(r *rand.Rand) interface{} {
	return r.Int63n(d.max-d.min) + d.min
}

type uniformFloat struct {
	min, max float64
}

func newUniformFloat(min, max float64) (uniformFloat, error) {
	if min >= max {
		return uniformFloat{}, errors.New("min must be less than max")
	}
	return uniformFloat{min, max}, nil
}

func (d uniformFloat) Sample(r *rand.Rand) interface{} {
	return r.Float64()*(d.max-d.min) + d.min
}

type logFloat struct {
	min, max float64
}

func newLogFloat(min, max float64) (logFloat, error) {
	if min <= 0 || max <= 0 {
		return logFloat{}, errors.New("min and max must be positive")
	}
	if min >= max {
		return logFloat{}, errors.New("min must be less than max")
	}
	return logFloat{math.Log(min), math.Log(max)}, nil
}

func (d logFloat) Sample(r *rand.Rand) interface{} {
	return math.Exp(r.Float64()*(d.max-d.min) + d.min)
}

type chooseOne struct {
	choices []interface{}
}

func newChooseOne(choices []interface{}) (chooseOne, error) {
	if len(choices) == 0 {
		return chooseOne{}, errors.New("at least one choice must be provided")
	}
	return chooseOne{choices}, nil
}

func (d chooseOne) Sample(r *rand.Rand) interface{} {
	return d.choices[r.Intn(len(d.choices))]
}
