package model

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Word string

func NewWord(s string) Word {
	return Word(strings.ToLower(strings.TrimSpace(s)))
}

func (w *Word) Validate() error {
	if w == nil || len(*w) == 0 {
		return fmt.Errorf("empty word")
	}
	return nil
}

func (w *Word) String() string {
	return string(*w)
}

func (w *Word) MarshalText() (text []byte, err error) {
	return []byte(*w), nil
}

func (w *Word) UnmarshalText(b []byte) error {
	word := NewWord(string(b))
	if err := word.Validate(); err != nil {
		return err
	}
	*w = word
	return nil
}

func (w *Word) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: string(*w)}, nil
}

func (w *Word) UnmarshalDynamoDBAttributeValue(v types.AttributeValue) error {
	s, ok := v.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("unsupported type for Word: %t", v)
	}

	word := NewWord(s.Value)
	if err := word.Validate(); err != nil {
		return err
	}
	*w = word
	return nil
}
