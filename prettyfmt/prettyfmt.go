package prettyfmt

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"unicode"
)

// Bytes returns human friendly string representation of the number of bytes.
func Bytes(bytes int64) string {
	kilo := bytes / 1024
	if kilo == 0 {
		return fmt.Sprintf("%d", bytes)
	}
	mega := kilo / 1024
	if mega == 0 {
		return fmt.Sprintf("%dK", kilo)
	}
	giga := mega / 1024
	if giga == 0 {
		return fmt.Sprintf("%dM", mega)
	}
	return fmt.Sprintf("%dG", giga)
}

func Val(val interface{}) string {
	var buf bytes.Buffer
	valR8n := reflect.ValueOf(val)
	fmtVal(&buf, valR8n)
	return buf.String()
}

func fmtVal(buf *bytes.Buffer, valR8n reflect.Value) {
	switch valR8n.Kind() {
	case reflect.Map:
		fmtMap(buf, valR8n)
	case reflect.Slice:
		fmtSlice(buf, valR8n)
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		buf.WriteString(fmt.Sprintf("%d", valR8n.Int()))
	default:
		buf.WriteString(fmt.Sprintf("%v", valR8n.Interface()))
	}
}

func fmtMap(buf *bytes.Buffer, mapR8n reflect.Value) {
	mapKeyR8ns := mapR8n.MapKeys()
	if len(mapKeyR8ns) == 0 {
		buf.WriteString("{}")
	}

	sort.Slice(mapKeyR8ns, func(i, j int) bool {
		return mapKeyR8ns[i].String() < mapKeyR8ns[j].String()
	})

	buf.WriteString("{")
	firstKey := true
	for _, mapKeyR8n := range mapKeyR8ns {
		if firstKey {
			buf.WriteString("\n")
			firstKey = false
		}
		buf.WriteString("    ")
		buf.WriteString(mapKeyR8n.String())
		buf.WriteString(": ")

		mapValR8n := mapR8n.MapIndex(mapKeyR8n)
		switch mapValR8n.Kind() {
		case reflect.Slice:
			fmtSlice(buf, mapValR8n)
		default:
			buf.WriteString(mapValR8n.String())
		}

		buf.WriteString("\n")
	}
	buf.WriteString("}")
}

func fmtSlice(buf *bytes.Buffer, sliceR8n reflect.Value) {
	buf.WriteString("[")
	firstElem := true
	for i := 0; i < sliceR8n.Len(); i++ {
		if firstElem {
			firstElem = false
		} else {
			buf.WriteString(" ")
		}
		sliceElemR8n := sliceR8n.Index(i)
		fmtVal(buf, sliceElemR8n)
	}
	buf.WriteString("]")
}

const (
	collapseStateOutside = iota
	collapseStateSkip
	collapseStateID
)

// CollapseJSON takes as input the output of json.MarshalIndent function and
// puts all lists to the same line. It was not intended to solve this problem
// in general case and works on a very limited set of inputs. Specifically to
// make the output of the `GET /topics/<>/consumers` API method look compact.
func CollapseJSON(bytes []byte) []byte {
	state := collapseStateOutside
	j := 0
	for i := 0; i < len(bytes); i++ {
		c := rune(bytes[i])
		switch state {
		case collapseStateOutside:
			if c == '[' {
				state = collapseStateSkip
			}

		case collapseStateSkip:
			if c == ']' {
				state = collapseStateOutside
				break
			}
			if unicode.IsDigit(c) {
				state = collapseStateID
				break
			}
			continue

		case collapseStateID:
			if unicode.IsDigit(c) {
				break
			}
			if c == ']' {
				state = collapseStateOutside
				break
			}
			if c == ',' {
				state = collapseStateSkip
				break
			}
			continue
		}
		bytes[j] = bytes[i]
		j++
	}
	return bytes[:j]
}
