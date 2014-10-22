package gojsonschema

import (
	"reflect"
	"strconv"
	"time"
)

// NormalizePayloads() is to change number formatted string to number, or date-time formatted string to unix timestamp, etc

func NormalizePayloads(schema *JsonSchemaDocument, document map[string]interface{}) error {
	_, err := normalizePayloadsRecursive(schema.rootSchema, document)
	return err
}

func normalizePayloadsRecursive(currentSchema *jsonSchema, currentNode interface{}) (interface{}, error) {

	if currentNode == nil {
		return currentNode, nil
	}

	if len(currentSchema.oneOf) > 0 {
		for _, oneOfSchema := range currentSchema.oneOf {
			updatedCurrentNode, err := normalizePayloadsRecursive(oneOfSchema, currentNode)
			if err == nil {
				currentNode = updatedCurrentNode
			}
		}
		return currentNode, nil
	}

	rValue := reflect.ValueOf(currentNode)
	rKind := rValue.Kind()

	switch rKind {

	case reflect.Slice:
		castCurrentNode := currentNode.([]interface{})
		for i := range castCurrentNode {
			updatedNextNode, err := normalizePayloadsRecursive(currentSchema.itemsChildren[0], castCurrentNode[i])
			if err == nil {
				castCurrentNode[i] = updatedNextNode
			}
		}

	case reflect.Map:
		castCurrentNode := currentNode.(map[string]interface{})
		for _, pSchema := range currentSchema.propertiesChildren {
			nextNode, ok := castCurrentNode[pSchema.property]
			if ok {
				updatedNextNode, err := normalizePayloadsRecursive(pSchema, nextNode)
				if err == nil {
					castCurrentNode[pSchema.property] = updatedNextNode
				}
			}
		}

	case reflect.String:
		value := currentNode.(string)

		if currentSchema.format != nil && *currentSchema.format == "number" {
			numberValue, err := strconv.ParseFloat(value, 64)
			if err == nil {
				currentNode = numberValue
			}
		}

		if currentSchema.format != nil && *currentSchema.format == "boolean" {
			if value == "true" {
				currentNode = true
			} else if value == "false" {
				currentNode = false
			}
		}

		if currentSchema.format != nil && *currentSchema.format == "date-time" {
			formats := []string{
				time.ANSIC,
				time.UnixDate,
				time.RubyDate,
				time.RFC822,
				time.RFC822Z,
				time.RFC850,
				time.RFC1123,
				time.RFC1123Z,
				time.RFC3339,
				time.RFC3339Nano}
			for _, format := range formats {
				timeValue, err := time.Parse(format, value)
				if err == nil {
					currentNode = timeValue.UnixNano() / 1e9
					break
				}
			}
		}

	}

	return currentNode, nil
}
