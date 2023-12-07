package bencode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Decode(reader io.Reader) (any, error) {
	return decode(bufio.NewReader(reader))
}

func decode(reader *bufio.Reader) (any, error) {
	if reader == nil || reader.Size() == 0 {
		return nil, nil
	}

	var (
		result any
		err    error
	)

	op, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("malformed input, cannot read structure prefix: %w", err)
	}

	switch op {
	case 'd': // dictionary
		result, err = parseDict(reader)
	case 'l': // list
		result, err = parseList(reader)
	case 'i': // integer
		result, err = parseInteger(reader)
	default: // string
		_ = reader.UnreadByte()
		result, err = parseString(reader)
	}

	return result, err
}

// parseDict parses dict input which is actually a list of tuples in form of d<key_1><val_1>...<key_n><val_n>e
func parseDict(reader *bufio.Reader) (map[string]any, error) {
	result := make(map[string]any)

	for {
		peek, err := reader.Peek(1)
		if err != nil {
			return nil, fmt.Errorf("failed to peek for dict suffix: %w", err)
		}

		if string(peek) == "e" {
			break
		}

		key, err := decode(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dict key: %w", err)
		}

		val, err := decode(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dict val: %w", err)
		}
		result[key.(string)] = val
	}

	_, err := reader.Discard(1) // trim suffix
	if err != nil {
		return nil, fmt.Errorf("failed to discard dict suffix: %w", err)
	}

	return result, nil
}

// parseList parses list input in form of l<el_1><...<el_n>e
func parseList(reader *bufio.Reader) ([]any, error) {
	var result []any
	for {
		peek, err := reader.Peek(1)
		if err != nil {
			return nil, fmt.Errorf("failed to peek for dict suffix: %w", err)
		}

		if string(peek) == "e" {
			break
		}

		el, err := decode(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse list: %w", err)
		}
		result = append(result, el)
	}

	_, err := reader.Discard(1) // trim suffix
	if err != nil {
		return nil, fmt.Errorf("failed to discard list suffix: %w", err)
	}

	return result, nil
}

// parseInteger parses integer input in form of i<sign><value>e
func parseInteger(reader *bufio.Reader) (int, error) {
	strVal, err := reader.ReadSlice('e')
	if err != nil {
		return 0, fmt.Errorf("failed to read integer value: %w", err)
	}

	result, err := strconv.Atoi(strings.TrimRight(string(strVal), "e"))
	if err != nil {
		return 0, fmt.Errorf("failed to parse integer value: %w", err)
	}

	return result, nil
}

// parseString doesn't care about leading zeros because it should handle them
func parseString(reader *bufio.Reader) (string, error) {
	strLen, err := reader.ReadSlice(':')
	if err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}

	length, err := strconv.Atoi(strings.TrimRight(string(strLen), ":"))
	if err != nil {
		return "", fmt.Errorf("failed to parse string length: %w", err)
	}

	if length == 0 {
		return "", nil
	}

	result, err := reader.Peek(length)
	if err != nil {
		return "", fmt.Errorf("failed to read string value: %w", err)
	}

	_, err = reader.Discard(length)
	if err != nil {
		return "", fmt.Errorf("failed to discard string value: %w", err)
	}

	return string(result), nil
}
