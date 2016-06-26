package schema

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/mapstructure"
)

// DiffFieldReader reads fields out of a diff structures.
//
// It also requires access to a Reader that reads fields from the structure
// that the diff was derived from. This is usually the state. This is required
// because a diff on its own doesn't have complete data about full objects
// such as maps.
//
// The Source MUST be the data that the diff was derived from. If it isn't,
// the behavior of this struct is undefined.
//
// Reading fields from a DiffFieldReader is identical to reading from
// Source except the diff will be applied to the end result.
//
// The "Exists" field on the result will be set to true if the complete
// field exists whether its from the source, diff, or a combination of both.
// It cannot be determined whether a retrieved value is composed of
// diff elements.
type DiffFieldReader struct {
	Diff   *terraform.InstanceDiff
	Source FieldReader
	Schema map[string]*Schema
}

func (r *DiffFieldReader) ReadField(address []string) (FieldReadResult, error) {
	schemaList := addrToSchema(address, r.Schema)
	if len(schemaList) == 0 {
		return FieldReadResult{}, nil
	}

	schema := schemaList[len(schemaList)-1]
	switch schema.Type {
	case TypeBool, TypeInt, TypeFloat, TypeString:
		return r.readPrimitive(address, schema)
	case TypeList:
		return r.readList(address, schema)
	case TypeMap:
		return r.readMap(address, schema)
	case TypeSet:
		return r.readSet(address, schema)
	case typeObject:
		return readObjectField(r, address, schema.Elem.(map[string]*Schema))
	default:
		panic(fmt.Sprintf("Unknown type: %#v", schema.Type))
	}
}

func (r *DiffFieldReader) readMap(
	address []string, schema *Schema) (FieldReadResult, error) {
	result := make(map[string]interface{})
	resultSet := false

	// First read the map from the underlying source
	source, err := r.Source.ReadField(address)
	if err != nil {
		return FieldReadResult{}, err
	}
	if source.Exists {
		result = source.Value.(map[string]interface{})
		resultSet = true
	}

	// Next, read all the elements we have in our diff, and apply
	// the diff to our result.
	prefix := strings.Join(address, ".") + "."
	for k, v := range r.Diff.Attributes {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		if strings.HasPrefix(k, prefix+"%") {
			// Ignore the count field
			continue
		}

		resultSet = true

		k = k[len(prefix):]
		if v.NewRemoved {
			delete(result, k)
			continue
		}

		result[k] = v.New
	}

	var resultVal interface{}
	if resultSet {
		resultVal = result
	}

	return FieldReadResult{
		Value:  resultVal,
		Exists: resultSet,
	}, nil
}

func (r *DiffFieldReader) readPrimitive(
	address []string, schema *Schema) (FieldReadResult, error) {
	result, err := r.Source.ReadField(address)
	if err != nil {
		return FieldReadResult{}, err
	}

	attrD, ok := r.Diff.Attributes[strings.Join(address, ".")]
	if !ok {
		return result, nil
	}

	var resultVal string
	if !attrD.NewComputed {
		resultVal = attrD.New
		if attrD.NewExtra != nil {
			result.ValueProcessed = resultVal
			if err := mapstructure.WeakDecode(attrD.NewExtra, &resultVal); err != nil {
				return FieldReadResult{}, err
			}
		}
	}

	result.Computed = attrD.NewComputed
	result.Exists = true
	result.Value, err = stringToPrimitive(resultVal, false, schema)
	if err != nil {
		return FieldReadResult{}, err
	}

	return result, nil
}

func (r *DiffFieldReader) readList(
	address []string, schema *Schema) (FieldReadResult, error) {

	prefix := strings.Join(address, ".") + "."

	var count int

	// If we have an empty list, then return an empty list
	diffCount, ok := r.Diff.Attributes[strings.Join(append(address, "#"), ".")]
	if ok {
		i, err := strconv.Atoi(diffCount.New)
		if err != nil {
			count = 0
		}
		count = i
	} else {
		originalCount, err := r.Source.ReadField(append(address, "#"))
		if err != nil {
			return FieldReadResult{
				Value:    []interface{}{},
				Exists:   false,
				Computed: false,
			}, err
		}
		count, ok = originalCount.Value.(int)
		if !ok {
			count = 0
		}
	}

	diffContainsField := false
	for k, _ := range r.Diff.Attributes {
		if strings.HasPrefix(k, address[0]+".") {
			diffContainsField = true
		}
	}
	if !diffContainsField {
		return FieldReadResult{
			Value:  []interface{}{},
			Exists: false,
		}, nil
	}

	// Create the list that will be our result
	list := []interface{}{}

	// Go through the map and find all the list items
	for i := 0; i < count; i++ {
		k := strings.Join(append(address, strconv.Itoa(i)), ".")

		d, ok := r.Diff.Attributes[k]
		if ok {
			if d.NewRemoved {
				// If the field is removed, we always ignore it
				continue
			}
		}
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		if strings.HasSuffix(k, "#") {
			// Ignore any count field
			continue
		}

		// Split the key, since it might be a sub-object like "idx.field"
		parts := strings.Split(k[len(prefix):], ".")
		idx := parts[0]

		raw, err := r.ReadField(append(address, idx))
		if err != nil {
			return FieldReadResult{}, err
		}
		if !raw.Exists {
			// This shouldn't happen because we just verified it does exist
			panic("missing field in set: " + k + "." + idx)
		}
		list = append(list, raw.Value)
	}

	// Determine if the list "exists". It exists if there are items or if
	// the diff explicitly wanted it empty.
	exists := len(list) > 0
	if !exists {
		// We could check if the diff value is "0" here but I think the
		// existence of "#" on its own is enough to show it existed. This
		// protects us in the future from the zero value changing from
		// "0" to "" breaking us (if that were to happen).
		if _, ok := r.Diff.Attributes[prefix+"#"]; ok {
			exists = true
		}
	}

	return FieldReadResult{
		Value:  list,
		Exists: exists,
	}, nil
}

func (r *DiffFieldReader) readSet(
	address []string, schema *Schema) (FieldReadResult, error) {
	prefix := strings.Join(address, ".") + "."

	// Create the set that will be our result
	set := schema.ZeroValue().(*Set)

	// Check if we're not supposed to remove it
	v, ok := r.Diff.Attributes[prefix+"#"]
	if ok && v.New == "0" {
		return FieldReadResult{
			Value:  set,
			Exists: true, // not entirely sure what's the point of returning empty set, but...
		}, nil
	}

	var originalSet *Set
	var err error

	// Compose common list of keys
	var keys []string
	diffContainsField := false
	for k, _ := range r.Diff.Attributes {
		if strings.HasPrefix(k, address[0]+".") {
			diffContainsField = true
		}
		keys = append(keys, k)
	}
	if !diffContainsField {
		return FieldReadResult{
			Value:  set,
			Exists: false,
		}, nil
	}
	originalResult, err := r.Source.ReadField(address)
	if err != nil {
	}
	if originalResult.Exists {
		originalSet = originalResult.Value.(*Set)
		originalMap := originalSet.Map()
		for k, _ := range originalMap {
			key := prefix + k
			_, ok := r.Diff.Attributes[key]
			if !ok {
				keys = append(keys, key)
			}
		}
	}

	sort.Strings(keys)

	// Go through the map and find all the set items
	for _, k := range keys {
		d, ok := r.Diff.Attributes[k]
		if ok {
			if d.NewRemoved {
				// If the field is removed, we always ignore it
				continue
			}
		}
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		if strings.HasSuffix(k, "#") {
			// Ignore any count field
			continue
		}

		// Split the key, since it might be a sub-object like "idx.field"
		parts := strings.Split(k[len(prefix):], ".")
		idx := parts[0]

		raw, err := r.ReadField(append(address, idx))
		if err != nil {
			return FieldReadResult{}, err
		}
		if !raw.Exists {
			// This shouldn't happen because we just verified it does exist
			panic("missing field in set: " + k + "." + idx)
		}

		set.Add(raw.Value)
	}

	// Determine if the set "exists". It exists if there are items or if
	// the diff explicitly wanted it empty.
	exists := set.Len() > 0
	if !exists {
		// We could check if the diff value is "0" here but I think the
		// existence of "#" on its own is enough to show it existed. This
		// protects us in the future from the zero value changing from
		// "0" to "" breaking us (if that were to happen).
		if _, ok := r.Diff.Attributes[prefix+"#"]; ok {
			exists = true
		}
	}

	return FieldReadResult{
		Value:  set,
		Exists: exists,
	}, nil
}
