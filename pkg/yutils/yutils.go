package yutils

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// FindMappingNode finds the next yaml node that matches the given key. if the value is not a mapping node, an error is returned.
func FindMappingNode(key string, node *yaml.Node) (found *yaml.Node, err error) {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			if node.Content[i+1].Kind != yaml.MappingNode {
				return nil, fmt.Errorf("key %q is not a mapping node", key)
			}
			return node.Content[i+1], nil
		}
	}
	return nil, nil
}

// Add adds a yaml node.
func Add(key string, node *yaml.Node) (newNode *yaml.Node) {
	var keyNode, valueNode yaml.Node
	keyNode.SetString(key)
	node.Content = append(node.Content, &keyNode, &valueNode)
	return &valueNode
}

// SetValueAtPath sets the value at the given path.
// all intermediary keys must exist as mappings or will be created as mappings.
func SetValueAtPath(keys []string, value string, data *yaml.Node) (err error) {
	current := data.Content[0]
	for idx := 0; idx < len(keys); idx++ {
		newCurrent, err := FindMappingNode(keys[idx], current)
		if err != nil {
			return err
		}
		if newCurrent != nil {
			current = newCurrent
			continue
		}

		current = Add(keys[idx], current)
		if idx != len(keys)-1 {
			current.Kind = yaml.MappingNode
			current.Tag = "!!map"
		}
	}
	current.SetString(value)
	return nil
}
