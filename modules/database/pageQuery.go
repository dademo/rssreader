package database

import (
	"errors"
	"fmt"
	"strings"
)

type Comparator int
type StringComparator int
type SortOrder int

const ( // Common values comparators
	ComparatorGreaterEqualThan = Comparator(iota)
	ComparatorGreaterThan
	ComparatorEquals
	ComparatorLowerThan
	ComparatorLowerEqualThan

	ComparatorDefault = ComparatorGreaterEqualThan
)

const ( // String comparators
	StrComparatorEquals = StringComparator(iota)
	StrComparatorContains
	StrComparatorMatches

	StrComparatorDefault = StrComparatorContains
)

const ( // Values sort order
	PageSortUndefined = SortOrder(iota)
	PageSortAsc
	PageSortDesc
)

const SortOptionSeparator = ","

type PageQuery struct {
	PageNo   uint
	PageSize uint
	Sort     []SortOption
}

type SortOption struct {
	field     string
	sortOrder SortOrder
}

func ParseSortOptions(value string) ([]SortOption, error) {

	allSortOptions := []SortOption{}
	// Format : "v1,asc,v2,desc" ; values will be upper-cased
	fieldsAndSort := strings.Split(value, SortOptionSeparator)
	if len(fieldsAndSort)%2 != 0 {
		if fieldsAndSort[len(fieldsAndSort)-1] != "" {
			return nil, errors.New("Bad number of key-values for sort order")
		} else {
			fieldsAndSort = fieldsAndSort[:len(fieldsAndSort)-1]
		}
	}

	for it := 0; it < len(fieldsAndSort); it += 2 {

		var sortOrder SortOrder
		field := fieldsAndSort[it]

		switch strings.ToUpper(fieldsAndSort[it+1]) {
		case "ASC":
			sortOrder = PageSortAsc
			break
		case "DESC":
			sortOrder = PageSortDesc
			break
		case "":
			sortOrder = PageSortUndefined
			break
		default:
			return nil, fmt.Errorf("Bad value for sort : expected 'ASC' or 'DESC', got [%s]", strings.ToUpper(fieldsAndSort[it*2+1]))
		}
		allSortOptions = append(allSortOptions, SortOption{
			field:     field,
			sortOrder: sortOrder,
		})
	}

	return allSortOptions, nil
}

func ParseComparator(value string) (Comparator, error) {

	switch strings.ToUpper(value) {
	case "GREATEROREQUAL", "GE":
		return ComparatorGreaterEqualThan, nil
	case "GREATERTHAN", "GT":
		return ComparatorGreaterThan, nil
	case "EQUALS", "EQ":
		return ComparatorEquals, nil
	case "LOWEROREQUAL", "LT":
		return ComparatorLowerThan, nil
	case "LOWERTHAN", "LE":
		return ComparatorLowerEqualThan, nil
	case "":
		return ComparatorDefault, nil
	default:
		return ComparatorDefault, fmt.Errorf("Bad comparator value for element [%s]", value)
	}
}

func ParseStringComparator(value string) (StringComparator, error) {

	switch strings.ToUpper(value) {
	case "EQUALS", "EQ":
		return StrComparatorEquals, nil
	case "CONTAINS":
		return StrComparatorContains, nil
	case "MATCHES":
		return StrComparatorMatches, nil
	case "":
		return StrComparatorDefault, nil
	default:
		return StrComparatorDefault, fmt.Errorf("Bad comparator value for element [%s]", value)
	}
}
