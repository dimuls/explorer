package hf

import (
	"encoding/json"
	"reflect"
	"regexp"

	"google.golang.org/protobuf/proto"
)

var protobufValueTypes = map[string]map[*regexp.Regexp]reflect.Type{}

func RegisterProtobufValueType(chaincodeName string, matchRegexp string,
	format proto.Message) {

	if protobufValueTypes[chaincodeName] == nil {
		protobufValueTypes[chaincodeName] = map[*regexp.Regexp]reflect.Type{}
	}
	protobufValueTypes[chaincodeName][regexp.MustCompile(matchRegexp)] =
		reflect.TypeOf(format)
}

func unmarshalProtobufValue(chaincodeName, key string, rawValue []byte) (
	interface{}, error) {

	ccValueTypes, exists := protobufValueTypes[chaincodeName]
	if !exists {
		return nil, nil
	}

	var format reflect.Type

	for m, f := range ccValueTypes {
		if m.MatchString(key) {
			format = f
			break
		}
	}

	if format == nil {
		return nil, nil
	}

	value := reflect.New(format)

	err := proto.Unmarshal(rawValue, value.Interface().(proto.Message))
	if err != nil {
		return nil, err
	}

	return value.Interface(), nil
}

var jsonValueTypes = map[string][]*regexp.Regexp{}

func RegisterJSONValueType(chaincodeName string, matchRegexp string) {
	jsonValueTypes[chaincodeName] = append(jsonValueTypes[chaincodeName],
		regexp.MustCompile(matchRegexp))
}

func isJSONValue(chaincodeName, key string) bool {
	ccValueTypes, exists := jsonValueTypes[chaincodeName]
	if !exists {
		return false
	}
	for _, m := range ccValueTypes {
		if m.MatchString(key) {
			return true
		}
	}
	return false
}

func parseValue(chaincodeName, key string, rawValue []byte) (json.RawMessage, error) {
	if isJSONValue(chaincodeName, key) {
		return rawValue, nil
	}
	v, err := unmarshalProtobufValue(chaincodeName, key, rawValue)
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}
