package flatten

import (
	"errors"
	"fmt"
	"strings"
)

var ErrOrAndMustBeArray = errors.New("the value of $or and $and must be an array")

// Flatten flattens extended JSON which is used to build and store queries.
//
// Payloads that look like this
//
//	{
//		"person": {
//			"age": {
//				"$gte": 5
//			}
//		}
//	}
//
// would become
//
//	{
//		"person.age": {
//		  "$gte": 5
//		}
//	}
func Flatten(input map[string]interface{}) (map[string]interface{}, error) {
	return flatten("", input)
}

func FlattenWithPrefix(prefix string, input map[string]interface{}) (map[string]interface{}, error) {
	return flatten(prefix, input)
}

func flatten(prefix string, nested interface{}) (map[string]interface{}, error) {
	f := map[string]interface{}{}

	switch n := nested.(type) {
	case map[string]interface{}:
		for key, value := range n {
			if strings.HasPrefix(key, "$") {
				if key == "$or" || key == "$and" {
					switch a := value.(type) {
					case []map[string]interface{}:
						for i := range a {
							t, err := flatten("", a[i])
							if err != nil {
								return nil, err
							}

							a[i] = t
						}

						f[key] = a
						return f, nil
					case []interface{}:
						for i := range a {
							t, err := flatten("", a[i].(map[string]interface{}))
							if err != nil {
								return nil, err
							}

							a[i] = t
						}

						f[key] = a
						return f, nil
					default:
						return nil, ErrOrAndMustBeArray
					}
				}

				// op is not $and or $or
				continue
			}

			m, err := flatten(key, value)
			if err != nil {
				return nil, err
			}

			for mKey, mValue := range m {
				nKey := mKey
				if len(prefix) > 0 {
					nKey = fmt.Sprintf("%s.%s", prefix, mKey)
				}
				f[nKey] = mValue
			}

			// the map is empty so flatten the parent.child
			// and set the value to the new key
			if len(m) == 0 {
				if len(prefix) > 0 {
					key = fmt.Sprintf("%s.%s", prefix, key)
				}
				f[key] = value
			}
		}
	default:
		f[prefix] = n
	}

	return f, nil
}
