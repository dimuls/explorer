package hf

import (
	"encoding/json"
	"reflect"
	"regexp"

	"google.golang.org/protobuf/proto"
)

type protobufvalueType struct {
	StateType string
	Format    reflect.Type
}

var protobufValueTypes = map[string]map[*regexp.Regexp]protobufvalueType{}

func RegisterProtobufValueType(chaincodeName string, keyMatchRegexp string,
	stateType string, format proto.Message) {

	if protobufValueTypes[chaincodeName] == nil {
		protobufValueTypes[chaincodeName] = map[*regexp.Regexp]protobufvalueType{}
	}

	protobufValueTypes[chaincodeName][regexp.MustCompile(keyMatchRegexp)] =
		protobufvalueType{
			StateType: stateType,
			Format:    reflect.TypeOf(format),
		}
}

func unmarshalProtobufValue(chaincodeName, key string, rawValue []byte) (
	string, interface{}, error) {

	ccValueTypes, exists := protobufValueTypes[chaincodeName]
	if !exists {
		return "", nil, nil
	}

	var (
		stateType string
		format    reflect.Type
	)

	for m, f := range ccValueTypes {
		if m.MatchString(key) {
			stateType = f.StateType
			format = f.Format
			break
		}
	}

	if format == nil {
		return "", nil, nil
	}

	value := reflect.New(format)

	err := proto.Unmarshal(rawValue, value.Interface().(proto.Message))
	if err != nil {
		return "", nil, err
	}

	return stateType, value.Interface(), nil
}

type jsonValueType struct {
	regexp    *regexp.Regexp
	stateType string
}

var jsonValueTypes = map[string][]jsonValueType{}

func RegisterJSONValueType(chaincodeName string, keyMatchRegexp string,
	stateType string) {

	jsonValueTypes[chaincodeName] = append(jsonValueTypes[chaincodeName],
		jsonValueType{
			regexp:    regexp.MustCompile(keyMatchRegexp),
			stateType: stateType,
		})
}

func isJSONValue(chaincodeName, key string) (string, bool) {

	ccValueTypes, exists := jsonValueTypes[chaincodeName]
	if !exists {
		return "", false
	}

	for _, vt := range ccValueTypes {
		if vt.regexp.MatchString(key) {
			return vt.stateType, true
		}
	}

	return "", false
}

func parseValue(chaincodeName, key string, rawValue []byte) (string, json.RawMessage, error) {

	stateType, isJSON := isJSONValue(chaincodeName, key)
	if isJSON {
		return stateType, rawValue, nil
	}

	stateType, v, err := unmarshalProtobufValue(chaincodeName, key, rawValue)
	if err != nil {
		return "", nil, err
	}

	vJSON, err := json.Marshal(v)
	if err != nil {
		return "", nil, err
	}

	return stateType, vJSON, nil
}
